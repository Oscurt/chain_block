package main

import (
    "fmt"
    "os"
	"bufio"
	"context"
    "log"
    "strings"
    "strconv"
    "github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/multiformats/go-multiaddr"
    "blockchain/common"
    "time"
    "encoding/json"
)

const SeedNodesEnvVar = "LIBP2P_SEED_NODES"

func ConnectToRandomNode() (host.Host, *peer.AddrInfo, bool, error) {
    // Cargar nodos semilla del archivo .env
    randomNode, err := common.ChooseRandomNodeFromEnv("")
    if err != nil {
        return nil, nil, false, err
    }

    // Iniciar un host libp2p
    h, err := libp2p.New()
    if err != nil {
        return nil, nil, false, fmt.Errorf("error al iniciar el host libp2p: %v", err)
    }

    // Convertir la dirección del nodo remoto a multiaddr
    peerAddr, err := multiaddr.NewMultiaddr(randomNode)
    if err != nil {
        return nil, nil, false, fmt.Errorf("dirección multiaddr inválida: %v", err)
    }

    // Extraer la información del peer del nodo remoto
    peerInfo, err := peer.AddrInfoFromP2pAddr(peerAddr)
    if err != nil {
        return nil, nil, false, fmt.Errorf("error al obtener la información del peer: %v", err)
    }

    // Conectar al nodo remoto
    if err := h.Connect(context.Background(), *peerInfo); err != nil {
        return nil, nil, false, fmt.Errorf("error al conectar con el nodo remoto: %v", err)
    }

    fmt.Println("Conectado con éxito al nodo:", randomNode)
    return h, peerInfo, true, nil
}

func showMenu(h host.Host, peerInfo *peer.AddrInfo) {
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Println("\nMenú Blockchain:")
        fmt.Println("1. Crear cuenta")
        fmt.Println("2. Obtener saldo")
        fmt.Println("3. Enviar saldo")
        fmt.Println("4. Salir")
        fmt.Print("Ingrese su opción: ")

        scanner.Scan()
        option := scanner.Text()

        switch option {
        case "1":
            createAccount(h, peerInfo)
        case "2":
            getBalanceByAddress(h, peerInfo)
        case "3":
            sendBalance(h, peerInfo)
        case "4":
            fmt.Println("Saliendo...")
            return
        default:
            fmt.Println("Opción no válida")
        }
    }
}

func sendBalance(h host.Host, peerInfo *peer.AddrInfo) {
    log.Println("Intentando enviar saldo...")

    scanner := bufio.NewScanner(os.Stdin)

    // Solicitar la dirección del destinatario
    fmt.Print("Ingrese la dirección del destinatario: ")
    scanner.Scan()
    recipientAddress := scanner.Text()

    // Solicitar la dirección del remitente
    fmt.Print("Ingrese su dirección: ")
    scanner.Scan()
    senderAddress := scanner.Text()

    // Solicitar la cantidad a enviar
    fmt.Print("Ingrese la cantidad a enviar: ")
    scanner.Scan()
    amountStr := scanner.Text()

    // Convertir la cantidad a float64
    amount, err := strconv.ParseFloat(amountStr, 64)
    if err != nil {
        fmt.Println("Error al convertir la cantidad:", err)
        return
    }

    // Solicitar la clave privada
    fmt.Print("Ingrese su clave privada: ")
    scanner.Scan()
    privateKey := scanner.Text()

    // Abrir un stream al nodo conectado
    s, err := h.NewStream(context.Background(), peerInfo.ID, "/send-balance")
    if err != nil {
        fmt.Println("Error al abrir stream:", err)
        return
    }
    defer s.Close()

    // Crear la transacción y convertirla a JSON
    transaction := common.Transaction{
        Sender:    senderAddress,
        Recipient: recipientAddress,
        Ammount:   amount,
        Signature: privateKey, // Enviar la clave privada de forma plana
        TimeStamp: time.Now().Unix(),
    }

    transactionData, err := json.Marshal(transaction)
    if err != nil {
        fmt.Println("Error al codificar la transacción:", err)
        return
    }

    transactionData = append(transactionData, '\n')

    // Enviar la transacción al nodo
    _, err = s.Write(transactionData)
    if err != nil {
        fmt.Println("Error al enviar la transacción:", err)
        return
    }

    // Leer la respuesta del nodo
    buf := bufio.NewReader(s)
    response, err := buf.ReadString('\n')
    if err != nil {
        fmt.Println("Error al leer la respuesta:", err)
        return
    }

    log.Println("Respuesta del nodo:", response)
}

func getBalanceByAddress(h host.Host, peerInfo *peer.AddrInfo) {
    // Abrir un stream al nodo conectado
    s, err := h.NewStream(context.Background(), peerInfo.ID, "/get-balance")
    if err != nil {
        fmt.Println("Error al abrir stream:", err)
        return
    }
    defer s.Close()

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Print("Ingrese la dirección: ")
    scanner.Scan()
    address := scanner.Text()

    // Enviar la dirección al nodo
    _, err = s.Write([]byte(address + "\n"))
    if err != nil {
        fmt.Println("Error al enviar la dirección:", err)
        return
    }

    // Leer la respuesta del nodo
    buf := bufio.NewReader(s)
    response, err := buf.ReadString('\n')
    if err != nil {
        fmt.Println("Error al leer la respuesta:", err)
        return
    }

    // Convertir la respuesta a float64
    balance, err := strconv.ParseFloat(strings.TrimSpace(response), 64)
    if err != nil {
        fmt.Println("Error al convertir la respuesta:", err)
        return
    }

    log.Println("Saldo de la cuenta", address + ":", balance)
}


func createAccount(h host.Host, peerInfo *peer.AddrInfo) {
    log.Println("Intentando crear cuenta...")
    // Abrir un stream al nodo seleccionado
    s, err := h.NewStream(context.Background(), peerInfo.ID, "/create-account")
    if err != nil {
        fmt.Println("Error al abrir stream:", err)
        return
    }
    defer s.Close()

    log.Println("Stream abierto con éxito. Enviando solicitud de creación de cuenta...")
    // Enviar una solicitud de creación de cuenta
    _, err = s.Write([]byte("create_account\n"))
    if err != nil {
        fmt.Println("Error al enviar solicitud:", err)
        return
    }

    log.Println("Solicitud enviada. Esperando respuesta...")
    // Esperar y recibir la respuesta del nodo
    buf := bufio.NewReader(s)
    response, err := buf.ReadString('\n')
    if err != nil {
        fmt.Println("Error al leer la respuesta:", err)
        return
    }

    log.Println("Respuesta del nodo:", response)
}

func main() {
    h, peerInfo, isConnected, err := ConnectToRandomNode()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    if isConnected {
        showMenu(h, peerInfo)
    } else {
        fmt.Println("No se pudo establecer una conexión con ningún nodo.")
        os.Exit(1)
    }
}