package network

import (
    "context"
    "log"
    "fmt"
    "strings"
    "encoding/json"
    "os"
    "github.com/joho/godotenv"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/multiformats/go-multiaddr"
    "blockchain/common"
)

const MaxSeedNodes = 20
const SeedNodesEnvVar = "LIBP2P_SEED_NODES"
const UserBroadcastProtocolID = "/user/broadcast/1.0.0"
const EnvFilePath = ".env"

func UpdateSeedNodes(currentNodeAddr string) {
    seedNodes, _ := godotenv.Read()
    existingNodes := seedNodes[SeedNodesEnvVar]
    nodes := strings.Split(existingNodes, ",")

    if !strings.Contains(existingNodes, currentNodeAddr) && len(nodes) < MaxSeedNodes {
        updatedNodes := existingNodes
        if len(updatedNodes) > 0 {
            updatedNodes += ","
        }
        updatedNodes += currentNodeAddr
        _ = godotenv.Write(map[string]string{SeedNodesEnvVar: updatedNodes}, EnvFilePath)
    }
}

func RemoveNodeFromEnv(filePath, currentNodeAddr string) {
    seedNodes, err := godotenv.Read()
    if err != nil {
        log.Printf("Error al leer el archivo .env: %v\n", err)
        return
    }

    nodes := strings.Split(seedNodes[SeedNodesEnvVar], ",")
    updatedNodes := make([]string, 0)

    for _, addr := range nodes {
        if addr != currentNodeAddr && addr != "" {
            updatedNodes = append(updatedNodes, addr)
        }
    }

    if len(updatedNodes) == 0 {
        // Si no quedan nodos semilla, eliminar el archivo .env
        err := os.Remove(filePath)
        if err != nil {
            log.Printf("Error al eliminar el archivo .env: %v\n", err)
        } else {
            log.Println("Archivo .env eliminado, ya que no quedan nodos semilla.")
        }
    } else {
        updatedNodesStr := strings.Join(updatedNodes, ",")
        if err := godotenv.Write(map[string]string{SeedNodesEnvVar: updatedNodesStr}, filePath); err != nil {
            log.Printf("Error al escribir en el archivo .env: %v\n", err)
        } else {
            log.Printf("Nodo eliminado del archivo .env: %s\n", currentNodeAddr)
        }
    }
}

func ConnectToSeedNodes(ctx context.Context, h host.Host, activeSeedNodes []string) {
    for _, addr := range activeSeedNodes {
        peerAddr, _ := multiaddr.NewMultiaddr(addr)
        peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)
        if err := h.Connect(ctx, *peerinfo); err != nil {
            log.Printf("Fallo al conectar con el nodo semilla activo: %s\n", addr)
        } else {
            log.Printf("Conectado con éxito al nodo semilla: %s\n", addr)
        }
    }
}

func VerifySeedNodes(ctx context.Context, h host.Host, currentNodeAddr string) []string {
    seedNodes, _ := godotenv.Read()
    nodes := strings.Split(seedNodes[SeedNodesEnvVar], ",")

    activeSeedNodes := make([]string, 0)

    for _, addr := range nodes {
        if addr == "" || addr == currentNodeAddr {
            continue
        }
        peerAddr, _ := multiaddr.NewMultiaddr(addr)
        peerinfo, _ := peer.AddrInfoFromP2pAddr(peerAddr)

        if err := h.Connect(ctx, *peerinfo); err != nil {
            log.Printf("Nodo semilla inactivo: %s\n", addr)
        } else {
            log.Printf("Nodo semilla activo: %s\n", addr)
            activeSeedNodes = append(activeSeedNodes, addr)
            h.Network().ClosePeer(peerinfo.ID)
        }
    }

    return activeSeedNodes
}

func SetupBroadcastStreamHandler(h host.Host) {
    h.SetStreamHandler(UserBroadcastProtocolID, func(s network.Stream) {
        defer s.Close()

        var user common.User
        if err := json.NewDecoder(s).Decode(&user); err != nil {
            fmt.Printf("Error al recibir datos de usuario: %v\n", err)
            return
        }

        fmt.Printf("Datos de usuario recibidos de %s: %v\n", s.Conn().RemotePeer(), user)
    })
}