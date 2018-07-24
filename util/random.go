package util

import (
	"github.com/Pallinder/go-randomdata"
	"math/rand"
	"sync"
	"time"
)

var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
var randomGeneratorMutex = sync.Mutex{}

func RandomHash(id []byte) {
	randomGeneratorMutex.Lock()
	defer randomGeneratorMutex.Unlock()

	randomGenerator.Read(id)
}

func RandomIP() string {
	return randomdata.IpV4Address()
}

func RandomName() string {
	return randomdata.SillyName()
}

func RandomString(size int) string {
	return randomdata.Letters(size)
}

func RandomInteger(min, max int) int {
	return randomdata.Number(min, max)
}
