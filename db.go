package main

import (
	"sync"

	"git.mills.io/prologic/bitcask"
)

var db *bitcask.Bitcask
var dbErr error
var once sync.Once

func GetDB() (*bitcask.Bitcask, error) {
	once.Do(func() { db, dbErr = bitcask.Open("/Users/raghuveernaraharisetti/Documents/lc-db/") })
	return db, dbErr
}
