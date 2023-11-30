package core

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
    "encoding/json"
    "github.com/syndtr/goleveldb/leveldb"
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

func CalculateHash(block Block) string {
    data := fmt.Sprintf("%d%d%s%d", block.Header.Index, block.Header.TimeStamp, block.Header.PrevBlock, block.Header.Nonce)
    for _, transaction := range block.Transactions {
        data += fmt.Sprintf("%d%s%s%f%d", transaction.Index, transaction.Sender, transaction.Recipient, transaction.Ammount, transaction.Signature)
    }
    h := sha256.New()
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}

func GenerateBlock(index int64, PrevBlock string, transactions []Transaction, nonce int64) Block {
    header := Header{index, PrevBlock, "", time.Now().Unix(), nonce}
    block := Block{header, transactions}
    block.Header.Hash = CalculateHash(block)
    return block
}

func SaveBlock(db leveldb.DB, blockBlock) error {
    blockData, err := json.Marshal(block)
    if err != nil {
        return err
    }
    err = db.Put([]byte(fmt.Sprintf("%d", block.Index)), blockData, nil)
    if err != nil {
        return err
    }
    return nil
}

func LoadBlock(db leveldb.DB, index int64) (Block, error) {
    blockData, err := db.Get([]byte(fmt.Sprintf("%d", index)), nil)
    if err != nil {
        return nil, err
    }
    var block Block
    err = json.Unmarshal(blockData, &block)
    if err != nil {
        return nil, err
    }
    return &block, nil
}