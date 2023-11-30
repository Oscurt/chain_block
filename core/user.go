package core

import (
    "github.com/tyler-smith/go-bip32"
    "github.com/tyler-smith/go-bip39"
    "crypto/sha256"
    "golang.org/x/crypto/ripemd160"
    "fmt"
    "github.com/syndtr/goleveldb/leveldb"
    "encoding/json"
)

type User struct {
    PrivateKey         *bip32.Key
    PublicKey          *bip32.Key
    Address            string
    Balance            float64
}
func generateMnemonic() string {
    entropy,  := bip39.NewEntropy(128)
    mnemonic,  := bip39.NewMnemonic(entropy)
    return mnemonic
}

func derivePrivateKey(mnemonic string) bip32.Key {
    seed := bip39.NewSeed(mnemonic, "")
    masterKey, _ := bip32.NewMasterKey(seed)
    return masterKey
}

func derivePublicKey(privateKeybip32.Key) bip32.Key {
    return privateKey.PublicKey()
}

func deriveAddress(publicKeybip32.Key) string {

    sha256 := sha256.New()
    sha256.Write(publicKey.Key)
    hash := sha256.Sum(nil)

    ripemd160 := ripemd160.New()
    ripemd160.Write(hash)
    hash = ripemd160.Sum(nil)

    address := fmt.Sprintf("%x", hash)

    return address
}

func NewUser(db leveldb.DB)User {

    mnemonic := generateMnemonic()
    privateKey := derivePrivateKey(mnemonic)
    publicKey := derivePublicKey(privateKey)
    address := deriveAddress(publicKey)

    user := &User{
        PrivateKey: privateKey,
        PublicKey: publicKey,
        Address: address,
        Balance: 0,
    }

    SaveUser(db, user)

    return user
}

func SaveUser(db leveldb.DB, userUser) error {
    userData := &User{
        PrivateKey: nil,
        PublicKey: user.PublicKey,
        Address: user.Address,
        Balance: user.Balance,
    }
    userDataBytes, err := json.Marshal(userData)
    if err != nil {
        return err
    }
    err = db.Put([]byte(user.Address), userDataBytes, nil)
    if err != nil {
        return err
    }
    return nil
}

func LoadUser(db leveldb.DB, address string) (User, error) {
    userDataBytes, err := db.Get([]byte(address), nil)
    if err != nil {
        return nil, err
    }
    var userData User
    err = json.Unmarshal(userDataBytes, &userData)
    if err != nil {
        return nil, err
    }
    return &userData, nil
}
