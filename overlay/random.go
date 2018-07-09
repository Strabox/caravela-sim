package overlay

import (
	"math/rand"
	"strconv"
	"sync"
	"time"
)

var randomGenerator = rand.New(rand.NewSource(time.Now().UnixNano()))
var randomGeneratorMutex = sync.Mutex{}
var randomIP = 0

func generateRandomHash(id []byte) {
	randomGeneratorMutex.Lock()
	defer randomGeneratorMutex.Unlock()

	randomGenerator.Read(id)
}

func generateRandomIP() string {
	randomGeneratorMutex.Lock()
	defer randomGeneratorMutex.Unlock()

	res := randomIP
	randomIP++
	return strconv.Itoa(res)
}
