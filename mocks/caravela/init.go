package caravela

import (
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/strabox/caravela-sim/util"
	caravelaConfig "github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node/common/guid"
	"os"
)

// Init initializes all the packages and static configurations of the caravela's project.
func Init(logLevel string, caravelaConfigs *caravelaConfig.Configuration) {
	// Init logs
	log.SetLevel(util.LogLevel(logLevel))
	log.SetFormatter(util.LogFormatter(true, true))
	log.SetOutput(os.Stdout)

	caravelaConfigs.Print()

	// Initialize CARAVELA's GUID package
	guid.Init(caravelaConfigs.ChordHashSizeBits(), int64(caravelaConfigs.GUIDEstimatedNetworkSize()),
		int64(caravelaConfigs.GUIDScaleFactor()))
}

// Configuration is a wrapper for obtaining the CARAVELA's configurations structure from the
// default file.
func Configuration() *caravelaConfig.Configuration {
	caravelaConfigs, err := caravelaConfig.ReadFromFile(util.RandomIP(), caravelaConfig.DefaultFilePath)
	if err != nil {
		panic(errors.New("problem reading CARAVELA's config file, error: " + err.Error()))
	}
	return caravelaConfigs
}
