package main

import (
	"log"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

const MasterDBPath = "data/12D3KooWPtCk6o3jWQ7aaesr1Sj7c5TgLCcG7huHCGFHiJhbzrE5"

func main() {
	masterDB, err := leveldb.OpenFile(MasterDBPath, nil)
    if err != nil {
		defer fmt.Errorf("error al abrir la base de datos maestra: %v", err)
    }
	defer masterDB.Close()

    iter := masterDB.NewIterator(nil, nil) // Iterar sobre todas las claves
    for iter.Next() {
        key := iter.Key()
        value := iter.Value()

        log.Printf("key %s with value %s\n", key, value)
    }
    iter.Release()
    if err := iter.Error(); err != nil {
        defer fmt.Errorf("error al iterar sobre la base de datos maestra: %v", err)
    }

    masterDB.Close()

}