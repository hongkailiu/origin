package cluster

import (
	"time"
)

// All metrics are defined in the cluster-loader (cl) config file
// except the ones mentioned otherwise
type Metrics interface {
	passCheck() bool
	printLog()
}

type BaseMetrics struct {
	Name string `json:"name"`
	Type string `json:"type"`
	result bool // our cl vendors the value
}

type TestDuration struct {
	BaseMetrics
	StartTime time.Time `json:"startTime"` // our cl vendors the value
	TestDuration time.Duration `json:"testDuration"` // our cl vendors the value
	DurationToPass time.Duration `json:"durationToPass"`
}


type ResourceState struct {
	BaseMetrics
	// k8s/oc resources type, eg, current version of cl supports build, dc
	// may allow for pod, rc ...
	ResourceType string `json:"resourceType"`
	Number int `json:"number"`
	// operator, 'eq' by default, support 'ge' and 'le' as well
	Operator string `json:"operator"`
	// the desired state of the resources, eg, Running, failed
	// might support regex
	// if nil, then it means ignoring status
	State string `json:"state"`
	// selector of resources
	Selector map[string]string `json:"selector"`
}

type Executable struct {
	BaseMetrics
	Path string `json:"path"`
	Timeout time.Duration `json:"timeout"`
	Stdout string `json:"stdout"` // our cl vendors the value
	Stderr string `json:"stderr"` // our cl vendors the value
}

func (td TestDuration) passCheck() bool {
	// TODO
	// if DurationToPass is defined by config
	// then compare with TestDuration to determine return value
	return false
}

func (td TestDuration) printLog() {
	// TODO
	// logging the object in json
}

func (rs ResourceState) passCheck() bool {
	// TODO
	// compare with operator Operator the number of desired resources
	// using the selector Selector
	// with the resource type ResourceType in the desired the state State
	// and the desired number Number
	return false
}

func (rs ResourceState) printLog() {
	// TODO
	// logging the object in json
}

func (e Executable) passCheck() bool {
	// run the executable specified by the path Path
	// with a timeout Timeout
	// if the exit code is 0, then return true
	// return false otherwise
	return false
}

func (e Executable) printLog() {
	// TODO
	// logging the object in json
}

// expected to be run in cl
func CheckMetrics(metrics []Metrics) {
	for _, m := range metrics {
		m.passCheck()
		m.printLog()
	}
}