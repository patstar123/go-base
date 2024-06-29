package spr

import (
	"encoding/gob"
	"net"
	"strconv"
	"sync"
)

const (
	BaseChildName = "baseChild"
	StopMethod    = BaseChildName + ".RpcStopChild"
	PingMethod    = BaseChildName + ".RpcPing"

	debugEnabled      = false
	debugRpcPort      = "46000"
	debugCallbackPort = "46001"
)

type RunnerTyper interface {
	CustomTypeValues() []any
}

type RpcSvr = RunnerTyper

func LoadRpcTypes(typer RunnerTyper) {
	values := typer.CustomTypeValues()
	for _, value := range values {
		gob.Register(value)
	}
}

var gMutex = sync.Mutex{}
var gMinPort = 49152
var gMaxPort = 65535
var gLastPort = 0

func SetRpcPortRange(min, max int) {
	gMutex.Lock()
	defer gMutex.Unlock()
	gMinPort = min
	gMaxPort = max
}

func findAvailablePort() int {
	gMutex.Lock()
	defer gMutex.Unlock()

	start := max(gMinPort, gLastPort+1)
	port := start
	for ; port <= gMaxPort; port++ {
		if isPortAvailable(port) {
			gLastPort = port
			return port
		}
	}

	if start == gMinPort {
		return -1
	}

	for port = gMinPort; port < start; port++ {
		if isPortAvailable(port) {
			gLastPort = port
			return port
		}
	}

	return -1
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", "127.0.0.1:"+strconv.Itoa(port))
	if err != nil {
		return false
	} else {
		ln.Close()
		return true
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
