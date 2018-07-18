package overlay

import (
	"github.com/Pallinder/go-randomdata"
	"math/rand"
	"sync"
	"time"
)

var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
var randomGeneratorMutex = sync.Mutex{}

func generateRandomHash(id []byte) {
	randomGeneratorMutex.Lock()
	defer randomGeneratorMutex.Unlock()

	randomGenerator.Read(id)
}

func generateRandomIP() string {
	return randomdata.IpV4Address()
}
