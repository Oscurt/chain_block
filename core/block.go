package core

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
    "encoding/json"
    "github.com/syndtr/goleveldb/leveldb"
    "blockchain/common"
)

func CalculateHash(block common.Block) string {
    data := fmt.Sprintf("%d%d%s%d", block.Header.Index, block.Header.TimeStamp, block.Header.PrevBlock, block.Header.Nonce)
    for _, transaction := range block.Transactions {
        data += fmt.Sprintf("%d%s%s%f%d", transaction.Index, transaction.Sender, transaction.Recipient, transaction.Ammount, transaction.Signature)
    }
    h := sha256.New()
    h.Write([]byte(data))
    return hex.EncodeToString(h.Sum(nil))
}

func GenerateBlock(index int64, PrevBlock string, transactions []common.Transaction, nonce int64) common.Block {
    header := common.Header{index, PrevBlock, "", time.Now().Unix(), nonce}
    block := common.Block{header, transactions}
    block.Header.Hash = CalculateHash(block)
    return block
}

func CreateGenesisBlock() (common.Block, float64) {
    var transactions []common.Transaction
    var totalAmmount float64 = 0

    // Crear transacciones iniciales para el bloque génesis
    for i := 0; i < 5; i++ {
        ammount := 1000000.0 // Cantidad inicial de moneda en la red
        totalAmmount += ammount
        genesisTransaction := common.Transaction{
            Index:     int64(i),
            Sender:    "0",
            Recipient: "OSCURT",
            Ammount:   ammount,
            Signature: "OSCURT",
            TimeStamp: time.Now().Unix(),
        }
        transactions = append(transactions, genesisTransaction)
    }

    // Crear el header del bloque génesis
    genesisHeader := common.Header{
        Index:     0,
        PrevBlock: "",
        Hash:      "OSCURT",
        TimeStamp: time.Now().Unix(),
        Nonce:     0,
    }

    // Crear el bloque génesis
    genesisBlock := common.Block{
        Header:       genesisHeader,
        Transactions: transactions,
    }

    // Calcular el hash del bloque génesis
    genesisBlock.Header.Hash = CalculateHash(genesisBlock)

    return genesisBlock, totalAmmount
}


func SaveBlock(db *leveldb.DB, block common.Block) error {
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

func LoadBlock(db *leveldb.DB, index int64) (*common.Block, error) {
    blockData, err := db.Get([]byte(fmt.Sprintf("%d", index)), nil)
    if err != nil {
        return nil, err
    }
    var block common.Block
    err = json.Unmarshal(blockData, &block)
    if err != nil {
        return nil, err
    }
    return &block, nil
}