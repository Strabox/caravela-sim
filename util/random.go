package util

import (
	"github.com/Pallinder/go-randomdata"
	caravelaUtil "github.com/strabox/caravela/util"
	"math/rand"
	"time"
)

// init initializes the random generator external package.
func init() {
	// This util random generator is initialized with a "random" seed because it is used to generate dummy information
	// that does not affect the critical random dependent selections.
	randomdata.CustomRand(rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(time.Now().UnixNano()))))
}

// RandomIP returns a random IPV4 Address.
func RandomIP() string {
	return randomdata.IpV4Address()
}

// RandomName returns a random name.
func RandomName() string {
	return randomdata.SillyName()
}

// RandomString returns a random set of characters.
func RandomString(size int) string {
	return randomdata.Letters(size)
}

// RandomInteger returns a integer in interval [min,max].
func RandomInteger(min, max int) int {
	return randomdata.Number(min, max+1)
}
