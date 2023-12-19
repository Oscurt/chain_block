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
    "net/http"
    "github.com/gorilla/mux"
)

const SeedNodesEnvVar = "LIBP2P_SEED_NODES"
var h host.Host
var peerInfo *peer.AddrInfo

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

    log.Println("Conectando al nodo:", randomNode)

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

/*
func showMenu(h host.Host, peerInfo *peer.AddrInfo) {
    scanner := bufio.NewScanner(os.Stdin)
    for {
        fmt.Println("\nMenú Blockchain:")
        fmt.Println("1. Crear cuenta")
        fmt.Println("2. Obtener saldo")
        fmt.Println("3. Enviar saldo")
        fmt.Println("4. Obtener transaccion")
        fmt.Println("5. Salir")
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
            getTrans(h, peerInfo)
        case "5":
            fmt.Println("Saliendo...")
            return
        default:
            fmt.Println("Opción no válida")
        }
    }
}*/

func sendBalance(h host.Host, peerInfo *peer.AddrInfo, senderAddress string, recipientAddress string, amount float64, privateKey string) string {
    log.Println("Intentando enviar saldo...")

    // Abrir un stream al nodo conectado
    s, err := h.NewStream(context.Background(), peerInfo.ID, "/send-balance")
    if err != nil {
        fmt.Println("Error al abrir stream:", err)
        return ""
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

    transaction.Hash = common.GenerateTransactionHash(transaction)

    transactionData, err := json.Marshal(transaction)
    if err != nil {
        fmt.Println("Error al codificar la transacción:", err)
        return ""
    }

    transactionData = append(transactionData, '\n')

    // Enviar la transacción al nodo
    _, err = s.Write(transactionData)
    if err != nil {
        fmt.Println("Error al enviar la transacción:", err)
        return ""
    }

    // Leer la respuesta del nodo
    buf := bufio.NewReader(s)
    response, err := buf.ReadString('\n')
    if err != nil {
        fmt.Println("Error al leer la respuesta:", err)
        return ""
    }

    log.Println("Respuesta del nodo:", response)

    response = fmt.Sprintf("Transacción enviada con éxito. Hash: %s", transaction)

    return response
}

func getBalanceByAddress(h host.Host, peerInfo *peer.AddrInfo, address string) string {
    // Abrir un stream al nodo conectado
    s, err := h.NewStream(context.Background(), peerInfo.ID, "/get-balance")
    if err != nil {
        fmt.Println("Error al abrir stream:", err)
        return ""
    }
    defer s.Close()

    // Enviar la dirección al nodo
    _, err = s.Write([]byte(address + "\n"))
    if err != nil {
        fmt.Println("Error al enviar la dirección:", err)
        return ""
    }

    // Leer la respuesta del nodo
    buf := bufio.NewReader(s)
    response, err := buf.ReadString('\n')
    if err != nil {
        fmt.Println("Error al leer la respuesta:", err)
        return ""
    }

    // Convertir la respuesta a float64
    balance, err := strconv.ParseFloat(strings.TrimSpace(response), 64)
    if err != nil {
        fmt.Println("Error al convertir la respuesta:", err)
        return ""
    }

    log.Println("Saldo de la cuenta", address + ":", balance)
    response = fmt.Sprintf("Saldo de la cuenta %s: %f", address, balance)
    return response
}

func getTrans(h host.Host, peerInfo *peer.AddrInfo) {
    log.Println("Intentando obtener transacción...")

    // Abrir un stream al nodo seleccionado
    s, err := h.NewStream(context.Background(), peerInfo.ID, "/get-trans")
    if err != nil {
        fmt.Println("Error al abrir stream:", err)
        return
    }
    defer s.Close()

    log.Println("Stream abierto con éxito. Enviando solicitud de obtener transacción...")

    scanner := bufio.NewScanner(os.Stdin)

    fmt.Print("Ingrese el hash de la transacción: ")
    scanner.Scan()
    hash := scanner.Text()

    // Enviar el hash al nodo
    _, err = s.Write([]byte(hash + "\n"))
    if err != nil {
        fmt.Println("Error al enviar el hash:", err)
        return
    }

    // Leer la respuesta del nodo
    buf := bufio.NewReader(s)
    response, err := buf.ReadString('\n')
    if err != nil {
        fmt.Println("Error al leer la respuesta:", err)
        return
    }

    // Convertir la respuesta a una estructura de transacción
    var transaction common.Transaction
    err = json.Unmarshal([]byte(response), &transaction)
    if err != nil {
        fmt.Println("Error al decodificar la transacción:", err)
        return
    }

    // Imprimir la transacción
    fmt.Printf("Transacción obtenida: %+v\n", transaction)
}



func createAccount(h host.Host, peerInfo *peer.AddrInfo) string {
    log.Println("Intentando crear cuenta...")
    // Abrir un stream al nodo seleccionado
    s, err := h.NewStream(context.Background(), peerInfo.ID, "/create-account")
    if err != nil {
        fmt.Println("Error al abrir stream:", err)
        return ""
    }
    defer s.Close()

    log.Println("Stream abierto con éxito. Enviando solicitud de creación de cuenta...")
    // Enviar una solicitud de creación de cuenta
    _, err = s.Write([]byte("create_account\n"))
    if err != nil {
        fmt.Println("Error al enviar solicitud:", err)
        return ""
    }

    log.Println("Solicitud enviada. Esperando respuesta...")
    // Esperar y recibir la respuesta del nodo
    buf := bufio.NewReader(s)
    response, err := buf.ReadString('\n')
    if err != nil {
        fmt.Println("Error al leer la respuesta:", err)
        return ""
    }

    log.Println("Respuesta del nodo:", response)
    return response
}

func main() {
    var err error
    h, peerInfo, _, err = ConnectToRandomNode()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    r := mux.NewRouter()
    r.HandleFunc("/create_account", createAccountHandler).Methods("POST")
    r.HandleFunc("/get_balance", getBalanceHandler).Methods("GET")
    r.HandleFunc("/send_balance", sendBalanceHandler).Methods("POST")
    r.HandleFunc("/get_transaction", getTransactionHandler).Methods("GET")
    
    log.Println("Starting server on :8080")
    log.Fatal(http.ListenAndServe(":8080", r))
}

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
    response := createAccount(h, peerInfo)
    fmt.Fprintf(w, response)
}

func getBalanceHandler(w http.ResponseWriter, r *http.Request) {
    data := r.URL.Query()
    address := data.Get("address")
    response := getBalanceByAddress(h, peerInfo, address)
    fmt.Fprintf(w, response)
}

func sendBalanceHandler(w http.ResponseWriter, r *http.Request) {
    data := r.URL.Query()
    sender := data.Get("sender")
    recipient := data.Get("recipient")
    amountStr := data.Get("amount")
    privateKey := data.Get("privateKey")

    amount, err := strconv.ParseFloat(amountStr, 64)
    if err != nil {
        fmt.Println("Error al convertir la cantidad:", err)
        return
    }

    response := sendBalance(h, peerInfo, sender, recipient, amount, privateKey)

    fmt.Fprintf(w, response)
}

func getTransactionHandler(w http.ResponseWriter, r *http.Request) {
    //getTrans(h, peerInfo)
}