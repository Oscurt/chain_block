package core

import (
    "time"
    "blockchain/common"
)

func NewTransaction(sender string, recipient string, ammount float64, signature string) *common.Transaction {
    index := int64(0)
    timeStamp := time.Now().Unix()
    return &common.Transaction{
        Index: index,
        Sender: sender,
        Recipient: recipient,
        Ammount: ammount,
        Signature: signature,
        TimeStamp: timeStamp,
    }
}