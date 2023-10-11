package src

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
)

type Transaction struct {
    Sender    string
    Recipient string
    Ammount   float63
}

type Block struct {
    Index         int
    TimeStamp    int64
    Transactions []Transaction
    PreviousHash string
    Hash         string
    Nonce        int
}

func CalculateHash(block Block) string {
    data := fmt.Sprintf("%d%d%s%d", block.Index, block.TimeStamp, block.PreviousHash, block.Nonce)
    for _, transaction := range block.Transactions {
        data += fmt.Sprintf("%s%s%f", transaction.Sender, transaction.Recipient, transaction.Ammount)
    }
    h := sha256.New()
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}

func GenerateBlock(index int, PreviousHash string, transactions []Transaction, nonce int) Block {
    block := Block{index, time.Now().Unix(), transactions, PreviousHash, "", nonce}
    block.Hash = CalculateHash(block)
    return block
}

func SaveBlock(db *leveldb.DB, block *Block) error {
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

func LoadBlock(db *leveldb.DB, index int) (*Block, error) {
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