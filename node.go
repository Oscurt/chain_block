package main

import (
    "context"
    "log"
    "fmt"
    "os"
    "os/signal"
    "github.com/libp2p/go-libp2p"
    "github.com/multiformats/go-multiaddr"
    "blockchain/network"
    "blockchain/database"
    "blockchain/core"
)

func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt)

    // Crea un nuevo nodo host

    h, err := libp2p.New()
    if err != nil {
        panic(err)
    }
    defer h.Close()

    network.SetupBroadcastStreamHandler(h)

    hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ipfs/%s", h.ID()))

    fullAddr := h.Addrs()[0].Encapsulate(hostAddr).String()
    log.Printf("Dirección completa de este nodo: %s\n", fullAddr)

    // Genera base de datos general

    master, err := database.InitDB("data/master")

    if err != nil {
        log.Fatalf("Failed to initialize DB: %v", err)
    }
    defer master.Close()

    // genera bloque genesis si no existe

    isEmpty := database.IsEmpty(master)

    if isEmpty {
        genesisBlock, amount := core.CreateGenesisBlock()
        err := core.SaveBlock(master, genesisBlock)
        if err != nil {
            log.Fatalf("Error al guardar el bloque génesis: %v", err)
        }
        log.Println("Bloque génesis creado y guardado con éxito.")

        master.Close()

        _, err = core.NewUser("data/" + h.ID().String(), h, amount)

    } else {
        log.Println("La blockchain ya existe, no se necesita crear un bloque génesis.")
    }

    master.Close()

    defer network.RemoveNodeFromEnv(network.EnvFilePath, fullAddr)

    network.UpdateSeedNodes(fullAddr)
    activeSeedNodes := network.VerifySeedNodes(ctx, h, fullAddr)
    network.ConnectToSeedNodes(ctx, h, activeSeedNodes)

    // Genera base de datos para el nodo

    db, err := database.InitDB("data/" + h.ID().String())
    if err != nil {
        log.Fatalf("Failed to initialize DB: %v", err)
    }
    defer func() {
        // Sincroniza la base de datos local con la base de datos maestra antes de cerrar
        err := database.SyncWithMasterDB("data/" + h.ID().String())
        if err != nil {
            log.Printf("Error al sincronizar con la base de datos maestra: %v", err)
        }

        // Cierra y elimina la base de datos local
        db.Close()
        err = os.RemoveAll("data/" + h.ID().String())
        if err != nil {
            log.Printf("Error al eliminar la base de datos local: %v", err)
        }
    }()

    db.Close()

    // Sincroniza la base de datos

    err = network.SyncDatabase(h, "data/" + h.ID().String(), activeSeedNodes)

    network.SetupCreateAccountHandler(h)
    network.SetupSyncHandler(h)
    network.SetupGetBalanceHandler(h)
    network.SetupSendHandler(h)
    network.SetupGetTransHandler(h)

    go func() {
        <-sigChan
        log.Println("Señal de interrupción recibida, limpiando...")
        cancel()
    }()

    <-ctx.Done()
}
