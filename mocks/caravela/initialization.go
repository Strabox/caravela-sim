package caravela

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela-sim/util"
	caravelaConfig "github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
	"os"
)

func Init(logLevel string) {
	// Init logs
	log.SetLevel(util.LogLevel(logLevel))
	log.SetFormatter(util.LogFormatter(true, true))
	log.SetOutput(os.Stdout)

	// Caravela configurations
	caravelaConfigs := Configuration()

	caravelaConfigs.Print()

	// Initialize CARAVELA's GUID package
	guid.Init(caravelaConfigs.ChordHashSizeBits())
}

func Configuration() *caravelaConfig.Configuration {
	caravelaConfigs, err := caravelaConfig.ReadFromFile(util.RandomIP())
	if err != nil {
		panic(fmt.Errorf("problem reading CARAVELA's config file: %s", err))
	}
	return caravelaConfigs
}
