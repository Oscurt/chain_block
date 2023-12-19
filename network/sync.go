package network

import (
	"fmt"
    "log"
    "bufio"
    "encoding/json"
    "strings"
    "context"
    "github.com/libp2p/go-libp2p/core/host"
    "github.com/libp2p/go-libp2p/core/network"
    "github.com/libp2p/go-libp2p/core/peer"
    "github.com/syndtr/goleveldb/leveldb"
    "blockchain/database"
    "blockchain/core"
    "blockchain/common"
)

const SyncProtocolID = "/blockchain/sync/1.0.0"

func SyncDatabase(h host.Host, localDBPath string, activeSeedNodes []string) error {
    var err error
    
    // Verificar si hay nodos activos para sincronizar
    if len(activeSeedNodes) > 0 {
        // Elegir un nodo activo al azar para la sincronización
        remoteNodeAddr, err := common.ChooseRandomNodeFromEnv(h.ID().String())
        if err != nil {
            return fmt.Errorf("error al elegir un nodo activo al azar: %v", err)
        }

        log.Println("Sincronizando con un nodo activo:", remoteNodeAddr)

        // Convertir la dirección del nodo remoto a peer.AddrInfo
        remotePeerInfo, err := peer.AddrInfoFromString(remoteNodeAddr)
        if err != nil {
            return fmt.Errorf("error al convertir la dirección del nodo remoto: %v", err)
        }

        // Abrir un stream con el nodo remoto
        s, err := h.NewStream(context.Background(), remotePeerInfo.ID, SyncProtocolID)
        if err != nil {
            return fmt.Errorf("error al abrir stream con el nodo remoto: %v", err)
        }
        defer s.Close()

        // Enviar solicitud de sincronización
        _, err = s.Write([]byte("sync_request\n"))
        if err != nil {
            return fmt.Errorf("error al enviar solicitud de sincronización: %v", err)
        }

        // Recibir datos del nodo remoto
        var allData map[string]interface{}
        if err := json.NewDecoder(s).Decode(&allData); err != nil {
            return fmt.Errorf("error al recibir datos del nodo remoto: %v", err)
        }

        err = updateLocalDatabase(localDBPath, allData)
        if err != nil {
            return fmt.Errorf("error al actualizar la base de datos local: %v", err)
        }

    } else {
        // Sincronizar con la base de datos maestra

        log.Println("No hay nodos activos para sincronizar, sincronizando con la base de datos maestra...")

        err = database.SyncWithMasterDB(localDBPath)
        if err != nil {
            return fmt.Errorf("error al sincronizar con la base de datos maestra: %v", err)
        }
    }

    return nil
}

func updateLocalDatabase(localDBPath string, data map[string]interface{}) error {
    db, err := leveldb.OpenFile(localDBPath, nil)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos local: %v", err)
    }
    defer db.Close()

    for key, value := range data {
        valueBytes, err := json.Marshal(value)
        if err != nil {
            return fmt.Errorf("error al serializar valor para la clave %s: %v", key, err)
        }
        if err := db.Put([]byte(key), valueBytes, nil); err != nil {
            return fmt.Errorf("error al actualizar la base de datos local para la clave %s: %v", key, err)
        }
    }

    db.Close()

    return nil
}

func SetupCreateAccountHandler(h host.Host) {
    h.SetStreamHandler("/create-account", func(s network.Stream) {
        defer s.Close()

        log.Println("Solicitud de creación de cuenta recibida. Leyendo solicitud...")
        // Leer la solicitud del cliente
        buf := bufio.NewReader(s)
        msg, err := buf.ReadString('\n')
        if err != nil {
            fmt.Println("Error al leer solicitud:", err)
            return
        }

        msg = strings.TrimSpace(msg)

        log.Println("Solicitud leída:", msg)
        if msg == "create_account" {
            log.Println("Procesando creación de cuenta...")
            // Crear una nueva cuenta
            user, err := core.NewUser("data/" + h.ID().String(), h, 0)
            if err != nil {
                fmt.Println("Error al crear usuario:", err)
                return
            }
            log.Println("Usuario creado con éxito:", user)

            // Enviar la información del usuario de vuelta al cliente
            userData, err := json.Marshal(user)
            if err != nil {
                fmt.Println("Error al codificar usuario:", err)
                return
            }

            log.Println("Enviando datos del usuario...")
            _, err = s.Write(append(userData, '\n'))
            if err != nil {
                fmt.Println("Error al enviar datos de usuario:", err)
                return
            }
            log.Println("Datos del usuario enviados con éxito.")
        } else {
            log.Println("Solicitud no válida.")
        }
    })
}

func SetupSyncHandler(h host.Host) {
    h.SetStreamHandler(SyncProtocolID, func(s network.Stream) {
        defer s.Close()

        log.Println("Solicitud de sincronización recibida.")

        // Leer la solicitud del nodo solicitante
        buf := bufio.NewReader(s)
        msg, err := buf.ReadString('\n')
        if err != nil {
            log.Println("Error al leer solicitud:", err)
            return
        }

        if strings.TrimSpace(msg) == "sync_request" {
            // Extraer toda la información de la base de datos
            dbData, err := extractDBData("data/" + h.ID().String())
            if err != nil {
                log.Println("Error al extraer datos de la base de datos:", err)
                return
            }

            // Enviar datos al nodo solicitante
            _, err = s.Write(dbData)
            if err != nil {
                log.Println("Error al enviar datos de sincronización:", err)
                return
            }

            log.Println("Datos de sincronización enviados con éxito.")
        } else {
            log.Println("Solicitud no reconocida.")
        }
    })
}

func SetupGetTransHandler(h host.Host) {
    h.SetStreamHandler("/get-trans", func(s network.Stream) {
        defer s.Close()

        log.Println("Solicitud de obtener transacción recibida.")
        
        // Leer el hash de la transacción del cliente
        buf := bufio.NewReader(s)
        hash, err := buf.ReadString('\n')
        if err != nil {
            fmt.Println("Error al leer el hash:", err)
            return
        }
        hash = strings.TrimSpace(hash)

        // Buscar la transacción en la base de datos
        transaction, err := getTransactionByHash("data/"+h.ID().String(), hash)
        if err != nil {
            fmt.Println("Error al obtener la transacción:", err)
            return
        }

        // Enviar la transacción al cliente
        transactionData, err := json.Marshal(transaction)
        if err != nil {
            fmt.Println("Error al codificar la transacción:", err)
            return
        }

        _, err = s.Write(append(transactionData, '\n'))
        if err != nil {
            fmt.Println("Error al enviar la transacción:", err)
            return
        }

        log.Println("Transacción enviada con éxito.")
    })
}


func getTransactionByHash(dbPath, hash string) (*common.Transaction, error) {
    db, err := leveldb.OpenFile(dbPath, nil)
    if err != nil {
        return nil, fmt.Errorf("error al abrir la base de datos: %v", err)
    }
    defer db.Close()

    iter := db.NewIterator(nil, nil)
    for iter.Next() {
        // Ignorar la clave "USER"
        if string(iter.Key()) == "USER" {
            continue
        }

        var block common.Block
        err := json.Unmarshal(iter.Value(), &block)
        if err != nil {
            continue // o manejar el error
        }

        for _, tx := range block.Transactions {
            if tx.Hash == hash {
                return &tx, nil
            }
        }
    }
    iter.Release()
    if err := iter.Error(); err != nil {
        return nil, err
    }

    return nil, fmt.Errorf("transacción no encontrada")
}


func SetupGetBalanceHandler(h host.Host) {
    h.SetStreamHandler("/get-balance", func(s network.Stream) {
        defer s.Close()

        log.Println("Solicitud de obtener saldo recibida.")
        // Leer la dirección del cliente
        buf := bufio.NewReader(s)
        address, err := buf.ReadString('\n')
        if err != nil {
            fmt.Println("Error al leer la dirección:", err)
            return
        }
        address = strings.TrimSpace(address)

        // Obtener el saldo de la dirección
        balance, err := getBalance(address, "data/"+h.ID().String())
        if err != nil {
            fmt.Println("Error al obtener el saldo:", err)
            return
        }

        // Enviar el saldo al cliente
        _, err = s.Write([]byte(fmt.Sprintf("%f\n", balance)))
        if err != nil {
            fmt.Println("Error al enviar el saldo:", err)
            return
        }

        log.Println("Saldo enviado con éxito.")
    })
}

func SetupSendHandler(h host.Host) {
    h.SetStreamHandler("/send-balance", func(s network.Stream) {
        defer s.Close()

        log.Println("Solicitud de envío de saldo recibida.")

        // Leer la transacción del cliente
        buf := bufio.NewReader(s)
        transactionData, err := buf.ReadString('\n')
        if err != nil {
            log.Printf("Error al leer la transacción: %v\n", err)
            return
        }

        if transactionData == "" {
            log.Println("No se recibieron datos de transacción.")
            return
        }

        log.Printf("Datos de transacción recibidos: %s\n", transactionData)

        var transaction common.Transaction
        err = json.Unmarshal([]byte(transactionData), &transaction)
        if err != nil {
            log.Printf("Error al decodificar la transacción: %v\n", err)
            return
        }

        log.Printf("Transacción decodificada: %+v\n", transaction)

        // Procesar la transacción
        err = processTransaction(transaction, "data/"+h.ID().String())
        if err != nil {
            log.Printf("Error al procesar la transacción: %v\n", err)
            response := fmt.Sprintf("Error al procesar la transacción: %v\n", err)
            _, err = s.Write([]byte(response))
            if err != nil {
                log.Printf("Error al enviar respuesta: %v\n", err)
                return
            }
        }

        // Enviar respuesta al cliente
        response := "Transacción procesada con éxito."
        _, err = s.Write([]byte(response + "\n"))
        if err != nil {
            log.Printf("Error al enviar respuesta: %v\n", err)
            return
        }

        log.Println("Respuesta enviada al cliente.")
    })
}

func processTransaction(transaction common.Transaction, dbPath string) error {
    err := updateBalances(transaction, dbPath)
    if err != nil {
        return fmt.Errorf("error al actualizar saldos: %v", err)
    }

    lastblock, err := database.GetLastBlockIndex(dbPath)
    if err != nil {
        return fmt.Errorf("error al obtener el último bloque: %v", err)
    }

    log.Println("Último bloque:", lastblock)

    // verifica si es bloque genesis o index 0

    dbMaster, err := leveldb.OpenFile("data/master", nil)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos maestra: %v", err)
    }
    defer dbMaster.Close()

    block, err := core.LoadBlock(dbMaster, lastblock)
    if err != nil {
        return fmt.Errorf("error al cargar el último bloque: %v", err)
    }

    dbMaster.Close()

    // ver cantidad de transacciones en el bloque

    if len(block.Transactions) >= 5 {
        log.Println("Creando nuevo bloque...")
        // crear nuevo bloque
        newBlock := core.GenerateBlock(lastblock+1, block.Header.Hash, []common.Transaction{transaction}, lastblock+1)
        // agremos la transacción al nuevo bloque
        newBlock.Transactions = append(newBlock.Transactions, transaction)

        // guardar el nuevo bloque
        // save en master
        masterDB, err := leveldb.OpenFile("data/master", nil)
        if err != nil {
            return fmt.Errorf("error al abrir la base de datos maestra: %v", err)
        }
        defer masterDB.Close()

        localDB, err := leveldb.OpenFile(dbPath, nil)
        if err != nil {
            return fmt.Errorf("error al abrir la base de datos local: %v", err)
        }
        defer localDB.Close()

        err = core.SaveBlock(masterDB, newBlock)
        err = core.SaveBlock(localDB, newBlock)

        masterDB.Close()
        localDB.Close()

    } else {
        // agregar transacción al bloque
        log.Println("Agregando transacción al bloque...")
        block.Transactions = append(block.Transactions, transaction)
        
        masterDB, err := leveldb.OpenFile("data/master", nil)
        if err != nil {
            return fmt.Errorf("error al abrir la base de datos maestra: %v", err)
        }
        defer masterDB.Close()

        localDB, err := leveldb.OpenFile(dbPath, nil)
        if err != nil {
            return fmt.Errorf("error al abrir la base de datos local: %v", err)
        }
        defer localDB.Close()

        err = core.SaveBlock(masterDB, *block)
        err = core.SaveBlock(localDB, *block)

        masterDB.Close()
        localDB.Close()
    }

    return nil
}

func updateBalances(transaction common.Transaction, dbPath string) error {
    // Abrir la base de datos maestra
    masterDB, err := leveldb.OpenFile("data/master", nil)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos maestra: %v", err)
    }
    defer masterDB.Close()

    // Obtener los usuarios de la base de datos maestra
    usersData, err := masterDB.Get([]byte("USER"), nil)
    if err != nil {
        return fmt.Errorf("error al obtener datos de usuarios: %v", err)
    }

    var users []*common.User
    err = json.Unmarshal(usersData, &users)
    if err != nil {
        return fmt.Errorf("error al deserializar usuarios: %v", err)
    }

    // Buscar el remitente y el destinatario en la lista de usuarios
    var sender, recipient *common.User
    for _, user := range users {
        if user.Address == transaction.Sender {
            sender = user
        } else if user.Address == transaction.Recipient {
            recipient = user
        }
    }

    // Verificar si el remitente tiene saldo suficiente

    if sender == nil {
        return fmt.Errorf("remitente no encontrado")
    }

    if recipient == nil {
        return fmt.Errorf("destinatario no encontrado")
    }

    if sender.Balance < transaction.Ammount {
        return fmt.Errorf("saldo insuficiente")
    }

    // Actualizar los saldos
    sender.Balance -= transaction.Ammount
    recipient.Balance += transaction.Ammount

    // Actualizar la base de datos local
    updatedUsersData, err := json.Marshal(users)
    if err != nil {
        return fmt.Errorf("error al serializar usuarios actualizados: %v", err)
    }

    err = masterDB.Put([]byte("USER"), updatedUsersData, nil)
    if err != nil {
        return fmt.Errorf("error al actualizar usuarios en la base de datos local: %v", err)
    }

    // Actualizar la base de datos local

    // Abrir la base de datos local

    db, err := leveldb.OpenFile(dbPath, nil)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos local: %v", err)
    }

    // Actualizar el saldo del remitente

    err = db.Put([]byte("USER"), updatedUsersData, nil)

    masterDB.Close()
    db.Close()

    return nil
}

func getBalance(address, dbPath string) (float64, error) {
    db, err := leveldb.OpenFile(dbPath, nil)
    if err != nil {
        return 0, fmt.Errorf("error al abrir la base de datos: %v", err)
    }
    defer db.Close()

    // Aquí asumimos que los datos del usuario están almacenados bajo la clave "USER"
    data, err := db.Get([]byte("USER"), nil)
    if err != nil {
        return 0, fmt.Errorf("error al obtener datos de usuarios: %v", err)
    }

    var users []*common.User
    err = json.Unmarshal(data, &users)
    if err != nil {
        return 0, fmt.Errorf("error al deserializar usuarios: %v", err)
    }

    for _, user := range users {
        if user.Address == address {
            return user.Balance, nil
        }
    }

    db.Close()

    return 0, fmt.Errorf("dirección no encontrada")
}

// extractDBData extrae los datos de la base de datos para la sincronización.
func extractDBData(dbPath string) ([]byte, error) {
    // Abrir la base de datos
    db, err := leveldb.OpenFile(dbPath, nil)
    if err != nil {
        return nil, fmt.Errorf("error al abrir la base de datos: %v", err)
    }
    defer db.Close()

    // Crear un mapa para almacenar todos los datos
    allData := make(map[string]interface{})

    // Iterar sobre todos los elementos en la base de datos
    iter := db.NewIterator(nil, nil)
    for iter.Next() {
        key := iter.Key()
        value := iter.Value()

        var data interface{}
        if err := json.Unmarshal(value, &data); err != nil {
            return nil, fmt.Errorf("error al deserializar el valor para la clave %s: %v", key, err)
        }

        allData[string(key)] = data
    }
    iter.Release()
    if err := iter.Error(); err != nil {
        return nil, fmt.Errorf("error al iterar sobre la base de datos: %v", err)
    }

    // Serializar el mapa completo a JSON
    jsonData, err := json.Marshal(allData)
    if err != nil {
        return nil, fmt.Errorf("error al serializar los datos de la base de datos: %v", err)
    }

    return jsonData, nil
}