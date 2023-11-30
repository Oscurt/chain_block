package database

import (
    "github.com/syndtr/goleveldb/leveldb"
)

var DB leveldb.DB

func GetDB()leveldb.DB {
    return DB
}

func InitDB(path string) {
    db, err := leveldb.OpenFile(path, nil)
    if err != nil {
        panic(err)
    }
    DB = db
}

func CloseDB() {
    DB.Close()
}

func Get(key []byte) ([]byte, error) {
    return DB.Get(key, nil)
}

func Put(key []byte, value []byte) error {
    return DB.Put(key, value, nil)
}

func Delete(key []byte) error {
    return DB.Delete(key, nil)
}