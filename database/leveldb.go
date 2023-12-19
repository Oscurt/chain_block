package database

import (
    "github.com/syndtr/goleveldb/leveldb"
    "time"
    "fmt"
    "encoding/json"
    "blockchain/common"
    "strconv"
)

const MasterDBPath = "data/master"
const RetryInterval = 5 * time.Second

func InitDB(path string) (*leveldb.DB, error) {
    db, err := leveldb.OpenFile(path, nil)
    if err != nil {
        return nil, err
    }
    return db, nil
}

func IsEmpty(db *leveldb.DB) bool {
    iter := db.NewIterator(nil, nil)
    defer iter.Release()
    return !iter.Next()
}

func Get(db *leveldb.DB, key []byte) ([]byte, error) {
    return db.Get(key, nil)
}

func Put(db *leveldb.DB, key []byte, value []byte) error {
    //log.Printf("Putting key %s with value %s\n", key, value)
    return db.Put(key, value, nil)
}

func Delete(db *leveldb.DB, key []byte) error {
    return db.Delete(key, nil)
}

func SyncWithMasterDB(localDBPath string) error {
    masterDB, err := openDBWithRetry(MasterDBPath)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos maestra: %v", err)
    }
    defer masterDB.Close()

    localDB, err := leveldb.OpenFile(localDBPath, nil)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos local: %v", err)
    }
    defer localDB.Close()

    // Copiar todos los datos de la base de datos maestra a la base de datos local
    iter := masterDB.NewIterator(nil, nil)
    defer iter.Release()
    for iter.Next() {
        key := iter.Key()
        value := iter.Value()

        err = Put(localDB, key, value)
        if err != nil {
            return fmt.Errorf("error al copiar datos a la base de datos local: %v", err)
        }
    }
    if err := iter.Error(); err != nil {
        return fmt.Errorf("error al iterar sobre la base de datos maestra: %v", err)
    }

    return nil
}

// SyncNewEntriesToMasterDB sincroniza solo las nuevas entradas de la base de datos maestra a la local.
func SyncNewEntriesToMasterDB(localDBPath string) error {
    masterDB, err := openDBWithRetry(MasterDBPath)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos maestra: %v", err)
    }
    defer masterDB.Close()

    localDB, err := leveldb.OpenFile(localDBPath, nil)
    if err != nil {
        return fmt.Errorf("error al abrir la base de datos local: %v", err)
    }
    defer localDB.Close()

    err = syncNewEntriesUser(masterDB, localDB)

    masterDB.Close()
    localDB.Close()

    return err
}

// openDBWithRetry intenta abrir una base de datos con reintentos.
func openDBWithRetry(dbPath string) (*leveldb.DB, error) {
    var db *leveldb.DB
    var err error
    for {
        db, err = leveldb.OpenFile(dbPath, nil)
        if err == nil {
            return db, nil
        }
        fmt.Println("Esperando para acceder a la base de datos:", dbPath)
        time.Sleep(RetryInterval)
    }
}

// syncNewEntriesUser copia las entradas nuevas de masterDB a localDB.
func syncNewEntriesUser(masterDB, localDB *leveldb.DB) error {
    // Obtener la lista de usuarios de la base de datos maestra
    masterUsersData, err := masterDB.Get([]byte("USER"), nil)
    if err != nil && err != leveldb.ErrNotFound {
        return fmt.Errorf("error al obtener usuarios de la base de datos maestra: %v", err)
    }

    var masterUsers []*common.User
    if err != leveldb.ErrNotFound {
        err = json.Unmarshal(masterUsersData, &masterUsers)
        if err != nil {
            return fmt.Errorf("error al deserializar usuarios de la base de datos maestra: %v", err)
        }
    }

    // Obtener la lista de usuarios de la base de datos local
    localUsersData, err := localDB.Get([]byte("USER"), nil)
    if err != nil && err != leveldb.ErrNotFound {
        return fmt.Errorf("error al obtener usuarios de la base de datos local: %v", err)
    }

    var localUsers []*common.User
    if err != leveldb.ErrNotFound {
        err = json.Unmarshal(localUsersData, &localUsers)
        if err != nil {
            return fmt.Errorf("error al deserializar usuarios de la base de datos local: %v", err)
        }
    }

    // Sincronizar usuarios de la base de datos maestra a la local
    updatedUsers := mergeUserLists(localUsers, masterUsers)
    updatedUsersData, err := json.Marshal(updatedUsers)
    if err != nil {
        return fmt.Errorf("error al serializar usuarios actualizados: %v", err)
    }

    err = localDB.Put([]byte("USER"), updatedUsersData, nil)
    if err != nil {
        return fmt.Errorf("error al actualizar usuarios en la base de datos local: %v", err)
    }

    return nil
}

// mergeUserLists combina dos listas de usuarios, evitando duplicados.
func mergeUserLists(localUsers, masterUsers []*common.User) []*common.User {
    userMap := make(map[string]*common.User)

    // Agregar o actualizar usuarios locales basados en la lista maestra
    for _, masterUser := range masterUsers {
        userMap[masterUser.Address] = masterUser
    }

    // Agregar usuarios locales que no están en la lista maestra
    for _, localUser := range localUsers {
        if _, exists := userMap[localUser.Address]; !exists {
            userMap[localUser.Address] = localUser
        }
    }

    // Crear una lista combinada de usuarios
    var combinedUsers []*common.User
    for _, user := range userMap {
        combinedUsers = append(combinedUsers, user)
    }

    return combinedUsers
}

func GetLastBlockIndex(dbPATH string) (int64, error) {
    var lastBlockIndex int64 = -1

    masterDB, err := openDBWithRetry(dbPATH)
    if err != nil {
        return -1, fmt.Errorf("error al abrir la base de datos: %v", err)
    }

    iter := masterDB.NewIterator(nil, nil)
    for iter.Next() {
        key := iter.Key()
        if string(key) == "USER" {
            continue
        }

        currentIndex, err := strconv.ParseInt(string(key), 10, 64)
        if err != nil {
            return -1, fmt.Errorf("error al convertir la clave a int64: %v", err)
        }

        if currentIndex > lastBlockIndex {
            lastBlockIndex = currentIndex
        }
    }
    iter.Release()
    if err := iter.Error(); err != nil {
        return -1, fmt.Errorf("error al iterar sobre la base de datos: %v", err)
    }

    if lastBlockIndex == -1 {
        return -1, fmt.Errorf("no se encontraron bloques en la base de datos")
    }

    masterDB.Close()

    return lastBlockIndex, nil
}