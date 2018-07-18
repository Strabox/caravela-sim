package caravela

import (
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/rest"
	"os"
)

const LocalIP = "127.0.0.1"
const FakePort = 8000
const FakeDockerAPIVersion = "1.35"

const FakeContainerImageKey = "fake_image_key"

func EmptyPortMappings() []rest.PortMapping {
	return make([]rest.PortMapping, 0)
}

func EmptyContainerArgs() []string {
	return make([]string, 0)
}

func InitLogs(logLevel string) {
	log.SetLevel(util.LogLevel(logLevel))
	log.SetFormatter(util.LogFormatter(true, true))
	log.SetOutput(os.Stdout)
}
