package caravela

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/strabox/caravela/api/rest"
	"os"
)

const NodeIDSizeBits = 128

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

func SetLogs() {
	log.SetLevel(log.DebugLevel)
	// Set the format of the log text and the place to write
	logOutputFormatter := &log.TextFormatter{}
	logOutputFormatter.DisableColors = true
	logOutputFormatter.DisableTimestamp = true
	log.SetFormatter(logOutputFormatter)
	log.SetOutput(os.Stdout)
}

const author = "André Pires"
const email = "pardal.pires@tecnico.ulisboa.pt"

/*
Prints the banner of the CARAVELA Simulator system.
*/
func PrintSimulatorBanner() {
	fmt.Printf("##################################################################\n")
	fmt.Printf("#      CARAVELA: A Cloud @ Edge (SIMULATOR)            000000       #\n")
	fmt.Printf("#            Author: %s                 00000000000     #\n", author)
	fmt.Printf("#  Email: %s           | ||| |      #\n", email)
	fmt.Printf("#              IST/INESC-ID                        || ||| ||     #\n")
	fmt.Printf("##################################################################\n")
}
