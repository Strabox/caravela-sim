package util

import (
	"github.com/Pallinder/go-randomdata"
	caravelaUtil "github.com/strabox/caravela/util"
	"math/rand"
	"time"
)

var randomGenerator = rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(time.Now().UnixNano())))

func RandomHash(id []byte) {
	randomGenerator.Read(id)
}

func init() {
	randomdata.CustomRand(rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(time.Now().UnixNano()))))
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
