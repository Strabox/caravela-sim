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
	"math"
	"os"
	"strconv"
	"sync"
	"time"
)

const logJsonFeederTag = "JS-FEEDER"

// jsonFeeder generates a stream of user requests reading from a json file.
type jsonFeeder struct {
	collector              *metrics.Collector             // Metrics collector that collects system level metrics.
	containerInjectionNode sync.Map                       // Map of ContainerID<->NodeIndex.
	currentRequests        sync.Map                       // Map of RequestID<->ContainerID.
	systemTotalResources   types.Resources                // Caravela's maximum resources.
	simConfigs             *configuration.Configuration   // Simulator's configurations.
	caravelaConfigs        *caravelaConfigs.Configuration // Caravela's configurations.
}

// newJsonFeeder creates a new json feeder.
func newJsonFeeder(simConfigs *configuration.Configuration, caravelaConfigs *caravelaConfigs.Configuration, _ int64) (Feeder, error) {
	return &jsonFeeder{
		collector:              nil,
		containerInjectionNode: sync.Map{},
		currentRequests:        sync.Map{},
		simConfigs:             simConfigs,
		caravelaConfigs:        caravelaConfigs,
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
	const maxJsonFiles = 20
	jsonFileCounter := 0
	tick := 0

	jsonRequestStream, currentJsonFile := j.getJsonRequestStream(fmt.Sprintf("in/Stream_%d.js", jsonFileCounter))

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

					if reqJson.EventType == 1 && !j.requestExists(requestID) { // Deploy container request.

						tickCpusAcc += reqResources.CPUs
						tickMemoryAcc += reqResources.Memory
						j.currentRequests.Store(requestID, nil)

						newTickChan <- func(nodeIndex int, injectedNode *node.Node, currentTime time.Duration) {
							if _, exist := j.currentRequests.Load(requestID); !exist {
								requestCtx := context.WithValue(context.Background(), types.RequestIDKey, requestID)
								j.collector.CreateRunRequest(nodeIndex, requestID, reqResources)
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
									j.containerInjectionNode.Store(contStatus[0].ContainerID, injectedNode)
									j.currentRequests.Store(requestID, contStatus[0].ContainerID)
								}
								j.currentRequests.Delete(requestID)
								j.collector.ArchiveRunRequest(requestID, err == nil)
							}
						}

					} else if (reqJson.EventType == 2 || reqJson.EventType == 3 || reqJson.EventType == 4 ||
						reqJson.EventType == 5 || reqJson.EventType == 6) && j.requestExists(requestID) { // Stop container request.

						contID, exist := j.currentRequests.Load(requestID)
						containerID, ok := contID.(string)
						if !exist || !ok {
							j.currentRequests.Delete(requestID)
							continue
						}

						injNode, exist := j.containerInjectionNode.Load(containerID)
						if !exist {
							continue
						}
						injectionNode, ok := injNode.(*node.Node)
						if !ok {
							panic("node to *node.Node")
						}

						newTickChan <- func(_ int, _ *node.Node, _ time.Duration) {
							err := injectionNode.StopContainers(context.Background(), []string{containerID})
							if err != nil {
								//util.Log.Infof(util.LogTag(logRandFeederTag)+"Stop container FAILED, err: %s", err)
							}
							j.currentRequests.Delete(requestID)
							j.containerInjectionNode.Delete(containerID)
						}

					}

				}

				if !jsonRequestStream.More() && (jsonFileCounter == maxJsonFiles) {
					util.Log.Fatalf(util.LogTag(logJsonFeederTag) + "No more requests in the json files!!!")
					currentJsonFile.Close() // Close the request file.
					close(newTickChan)      // No more user requests for this tick.
					return
				} else if !jsonRequestStream.More() {
					jsonFileCounter++
					currentJsonFile.Close() // Close the request file.
					jsonRequestStream, currentJsonFile = j.getJsonRequestStream(fmt.Sprintf("in/Stream_%d.js", jsonFileCounter))
				}

				close(newTickChan) // No more user requests for this tick.
			} else { // Simulator closed ticks channel.
				currentJsonFile.Close() // Close the request file.
				return                  // Stop feeding engine
			}
		}
		tick++
	}
}

func (j *jsonFeeder) getJsonRequestStream(filePath string) (*json.Decoder, *os.File) {
	fileReader, err := os.Open(filePath)
	if err != nil {
		panic(fmt.Errorf("json feeder invalid request stream file: %s", err))
	}

	jsonRequestStream := json.NewDecoder(fileReader)
	_, err = jsonRequestStream.Token() // Read open bracket
	if err != nil {
		panic(err)
	}
	return jsonRequestStream, fileReader
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

	/*
		chosenCpuClass := 0
		randInt := j.randomGenerator.Intn(101)
		for i, cpuClass := range cpuClasses {
			if randInt <= cpuClass {
				chosenCpuClass = i
				break
			}
		}
	*/

	return types.Resources{
		CPUClass: 0,
		CPUs:     int(math.Ceil((4 * normalizedCpus) * float64(maxCpus))),
		Memory:   int(math.Ceil((4 * normalizedMemory) * float64(maxMemory))),
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
	_, exist := j.currentRequests.Load(requestID)
	return exist
}
