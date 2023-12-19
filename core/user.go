package core

import (
    "github.com/tyler-smith/go-bip32"
    "github.com/tyler-smith/go-bip39"
    "crypto/sha256"
    "golang.org/x/crypto/ripemd160"
    "fmt"
    "log"
    "github.com/syndtr/goleveldb/leveldb"
    "github.com/libp2p/go-libp2p/core/host"
    "encoding/json"
    "blockchain/common"
    "blockchain/database"
)

func generateMnemonic() string {
    entropy, _ := bip39.NewEntropy(128)
    mnemonic, _ := bip39.NewMnemonic(entropy)
    return mnemonic
}

func derivePrivateKey(mnemonic string) *bip32.Key {
    seed := bip39.NewSeed(mnemonic, "")
    masterKey, _ := bip32.NewMasterKey(seed)
    return masterKey
}

func derivePublicKey(privateKey *bip32.Key) *bip32.Key {
    return privateKey.PublicKey()
}

func deriveAddress(publicKey *bip32.Key) string {

    sha256 := sha256.New()
    sha256.Write(publicKey.Key)
    hash := sha256.Sum(nil)

    ripemd160 := ripemd160.New()
    ripemd160.Write(hash)
    hash = ripemd160.Sum(nil)

    address := fmt.Sprintf("%x", hash)

    return address
}

func NewUser(db string, h host.Host, master float64) (common.User, error){

    mnemonic := generateMnemonic()
    privateKey := derivePrivateKey(mnemonic)
    publicKey := derivePublicKey(privateKey)
    address := deriveAddress(publicKey)

    user := &common.User{
        PrivateKey: privateKey,
        PublicKey: publicKey,
        Address: address,
        Balance: master,
    }

    log.Println("Usuario generado:", user)
    err := SaveUser(db, user)
    if err != nil {
        return common.User{}, err
    }

    log.Println("Usuario guardado con éxito en la base de datos.")

    return *user, nil
}

func updateUserList(db *leveldb.DB, user *common.User) error {
    data, err := db.Get([]byte("USER"), nil)
    if err != nil && err != leveldb.ErrNotFound {
        return err
    }

    var users []*common.User
    if err != leveldb.ErrNotFound {
        err = json.Unmarshal(data, &users)
        if err != nil {
            return err
        }
    }

    users = append(users, user)
    updatedData, err := json.Marshal(users)
    if err != nil {
        return err
    }

    return db.Put([]byte("USER"), updatedData, nil)
}

func SaveUser(dbPath string, user *common.User) error {
    // Abrir la base de datos maestra
    masterDB, err := leveldb.OpenFile("data/master", nil)
    if err != nil {
        log.Fatalf("Failed to initialize master DB: %v", err)
    }
    defer masterDB.Close()

    // Leer y actualizar la lista de usuarios en la base de datos maestra
    err = updateUserList(masterDB, user)
    if err != nil {
        return err
    }

    masterDB.Close()

    // Sincronizar la base de datos local con la maestra
    err = database.SyncNewEntriesToMasterDB(dbPath)
    if err != nil {
        log.Printf("Error al sincronizar con la base de datos maestra: %v", err)
        return err
    }

    return nil
}

func LoadUser(db *leveldb.DB, address string) (*common.User, error) {
    userDataBytes, err := db.Get([]byte(address), nil)
    if err != nil {
        return nil, err
    }
    var userData common.User
    err = json.Unmarshal(userDataBytes, &userData)
    if err != nil {
        return nil, err
    }
    return &userData, nil
}