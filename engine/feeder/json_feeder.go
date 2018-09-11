package feeder

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/strabox/caravela-sim/configuration"
	"github.com/strabox/caravela-sim/engine/metrics"
	"github.com/strabox/caravela-sim/mocks/caravela"
	"github.com/strabox/caravela-sim/util"
	"github.com/strabox/caravela/api/types"
	caravelaConfigs "github.com/strabox/caravela/configuration"
	"github.com/strabox/caravela/node"
	caravelaUtil "github.com/strabox/caravela/util"
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

const logJsonFeederTag = "JS-FEEDER"

// jsonFeeder generates a stream of user requests reading from a json file.
type jsonFeeder struct {
	collector            *metrics.Collector             // Metrics collector that collects system level metrics.
	reqInjectionNode     sync.Map                       // Map of ContainerID<->NodeIndex.
	systemTotalResources types.Resources                // Caravela's maximum resources.
	randomGenerator      *rand.Rand                     // Pseudo-random generator.
	simConfigs           *configuration.Configuration   // Simulator's configurations.
	caravelaConfigs      *caravelaConfigs.Configuration // Caravela's configurations.
}

// newJsonFeeder creates a new json feeder.
func newJsonFeeder(simConfigs *configuration.Configuration, caravelaConfigs *caravelaConfigs.Configuration, rngSeed int64) (Feeder, error) {
	return &jsonFeeder{
		collector:        nil,
		reqInjectionNode: sync.Map{},
		randomGenerator:  rand.New(caravelaUtil.NewSourceSafe(rand.NewSource(rngSeed))),
		simConfigs:       simConfigs,
		caravelaConfigs:  caravelaConfigs,
	}, nil
}

func (j *jsonFeeder) Init(metricsCollector *metrics.Collector, systemTotalResources types.Resources) {
	j.collector = metricsCollector
	j.systemTotalResources = systemTotalResources
}

func (j *jsonFeeder) Start(ticksChannel <-chan chan RequestTask) {
	type Request struct {
		Time      int64   `json:"Time"`
		JobID     int64   `json:"job id"`
		EventType int     `json:"event type"`
		User      string  `json:"user"`
		CPUs      float64 `json:"CPU request"`
		Memory    float64 `json:"memory request"`
	}
	tick := 0

	fileReader, err := os.Open("RequestStream.js")
	if err != nil {
		panic(fmt.Errorf("json feeder invalid request stream file: %s", err))
	}

	jsonRequestStream := json.NewDecoder(fileReader)
	_, err = jsonRequestStream.Token() // Read open bracket
	if err != nil {
		panic(err)
	}

	for {
		select {
		case newTickChan, more := <-ticksChannel: // Send all the requests for this tickChan
			if more {
				tickCpusAcc := 0
				tickMemoryAcc := 0

				// Generate the requests from the json request stream.
				for jsonRequestStream.More() && (j.ratioSystemResources(tickCpusAcc, tickMemoryAcc) < 0.05) {
					var reqJson Request
					err := jsonRequestStream.Decode(&reqJson)
					if err != nil {
						panic(err)
					}

					requestID := strconv.FormatInt(reqJson.JobID, 10)
					reqResources := j.generateRequestResources(reqJson.CPUs, reqJson.Memory)
					tickCpusAcc += reqResources.CPUs
					tickMemoryAcc += reqResources.Memory

					if reqJson.EventType == 1 && !j.requestExists(requestID) { // Deploy container request.

						newTickChan <- func(nodeIndex int, injectedNode *node.Node, currentTime time.Duration) {
							requestCtx := context.WithValue(context.Background(), types.RequestIDKey, requestID)
							j.collector.CreateRunRequest(nodeIndex, requestID, reqResources, currentTime)
							contStatus, err := injectedNode.SubmitContainers(
								requestCtx,
								[]types.ContainerConfig{{
									ImageKey:     util.RandomName(),
									Name:         util.RandomName(),
									PortMappings: caravela.EmptyPortMappings(),
									Args:         caravela.EmptyContainerArgs(),
									Resources:    reqResources,
									GroupPolicy:  types.SpreadGroupPolicy,
								}})
							if err == nil {
								j.reqInjectionNode.Store(contStatus[0].ContainerID, injectedNode)
								j.collector.RunRequestSucceeded()
							}
							j.collector.ArchiveRunRequest(requestID)
						}

					} else if (reqJson.EventType == 2 || reqJson.EventType == 3 || reqJson.EventType == 4 ||
						reqJson.EventType == 5 || reqJson.EventType == 6) && j.requestExists(requestID) { // Stop container request.

						j.reqInjectionNode.Range(func(key, value interface{}) bool {
							containerID, _ := key.(string)
							injectionNode, _ := value.(*node.Node)
							newTickChan <- func(_ int, _ *node.Node, _ time.Duration) {
								err := injectionNode.StopContainers(context.Background(), []string{containerID})
								if err != nil {
									//util.Log.Infof(util.LogTag(logRandFeederTag)+"Stop container FAILED, err: %s", err)
								}
								j.reqInjectionNode.Delete(containerID)
							}
							return false
						})

					}

				}

				if !jsonRequestStream.More() {
					fileReader.Close() // Close the request file.
					close(newTickChan) // No more user requests for this tick.
					return
				}

				close(newTickChan) // No more user requests for this tick.
			} else { // Simulator closed ticks channel.
				fileReader.Close() // Close the request file.
				return             // Stop feeding engine
			}
		}
		tick++
	}
}

// generateRequestResources ...
func (j *jsonFeeder) generateRequestResources(normalizedCpus, normalizedMemory float64) types.Resources {
	cpuClasses := make([]int, len(j.caravelaConfigs.ResourcesPartitions().CPUClasses))
	maxCpus := 0
	maxMemory := 0

	cpuClassAcc := 0
	for i, cpuClass := range j.caravelaConfigs.ResourcesPartitions().CPUClasses {
		cpuClasses[i] = cpuClass.Percentage + cpuClassAcc
		cpuClassAcc += cpuClass.Percentage
		for _, cpus := range cpuClass.CPUCores {
			if cpus.Value > maxCpus {
				maxCpus = cpus.Value
			}
			for _, memory := range cpus.Memory {
				if memory.Value > maxMemory {
					maxMemory = memory.Value
				}
			}
		}
	}

	chosenCpuClass := 0
	randInt := j.randomGenerator.Intn(101)
	for i, cpuClass := range cpuClasses {
		if randInt <= cpuClass {
			chosenCpuClass = i
		}
	}

	return types.Resources{
		CPUClass: types.CPUClass(chosenCpuClass),
		CPUs:     int(math.Ceil(normalizedCpus * float64(maxCpus))),
		Memory:   int(math.Ceil(normalizedMemory * float64(maxMemory))),
	}
}

// ratioSystemResources returns the ratio of resources given considered the system's total resources.
func (j *jsonFeeder) ratioSystemResources(cpus, memory int) float64 {
	cpusRatio := float64(cpus) / float64(j.systemTotalResources.CPUs)
	memoryRatio := float64(memory) / float64(j.systemTotalResources.Memory)
	return (cpusRatio + memoryRatio) / 2
}

// requestExists verifies if a request was already injected in the request stream.
func (j *jsonFeeder) requestExists(requestID string) bool {
	_, ok := j.reqInjectionNode.Load(requestID)
	return ok
}
