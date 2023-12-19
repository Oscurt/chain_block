package common

import (
    "github.com/tyler-smith/go-bip32"
)

type Header struct {
    Index       int64
    PrevBlock   string
    Hash        string
    TimeStamp   int64
    Nonce       int64
}

type Block struct {
    Header
    Transactions    []Transaction
}

type Transaction struct {
    Index       int64
    Sender      string
    Recipient   string
    Ammount     float64
    Signature   string
    TimeStamp   int64
    Hash        string
}

type User struct {
    PrivateKey         *bip32.Key
    PublicKey          *bip32.Key
    Address            string
    Balance            float64
}