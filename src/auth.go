package src

import (
    "fmt"
    "github.com/brianium/mnemonic"
    "github.com/brianium/mnemonic/entropy"
)

func Keys() mnemonic.Mnemonic {
	ent, _ := entropy.Random(256)
	mne, _ := mnemonic.New(ent)
	return mne
}