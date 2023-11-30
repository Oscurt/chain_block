package core

import (
    "time"
)

type Transaction struct {
    Index       int64
    Sender      string
    Recipient   string
    Ammount     float64
    Signature   string
    TimeStamp   int64
}

func NewTransaction(sender string, recipient string, ammount float64, signature string) *Transaction {
    index := int64(0)
    timeStamp := time.Now().Unix()
    return &Transaction{
        Index: index,
        Sender: sender,
        Recipient: recipient,
        Ammount: ammount,
        Signature: signature,
        TimeStamp: timeStamp,
    }
}