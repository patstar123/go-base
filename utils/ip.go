package utils

import (
	"fmt"
	"golang.org/x/exp/slices"
	"net"
	"strings"
)

type TransProtocol string

const (
	ProTcp       TransProtocol = "tcp"
	ProUdp       TransProtocol = "udp"
	ProTcpClient TransProtocol = "tcp-c"
	ProTcpServer TransProtocol = "tcp-s"
)

// CheckTCPPortAvailable 检查指定端口是否可用
func CheckTCPPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	ln.Close()
	return true
}

// FindAvailableTCPPort 在指定范围内查找可用的TCP端口
func FindAvailableTCPPort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		if CheckTCPPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available TCP port found in range %d-%d", start, end)
}

// CheckUDPPortAvailable 检查指定端口是否可用
func CheckUDPPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// FindAvailableUDPPort 在指定范围内查找可用的UDP端口
func FindAvailableUDPPort(start, end int) (int, error) {
	for port := start; port <= end; port++ {
		if CheckUDPPortAvailable(port) {
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available UDP port found in range %d-%d", start, end)
}

// FindAvailablePort 在指定范围内查找可用的端口
func FindAvailablePort(start, end int, protocol TransProtocol) (int, error) {
	if protocol == ProUdp {
		return FindAvailableUDPPort(start, end)
	} else if protocol == ProTcp {
		return FindAvailableTCPPort(start, end)
	} else {
		panic("unknown protocol: " + protocol)
	}
}

func GetLocalIPAddresses(includeLoopback bool, preferredInterfaces []string) ([]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	loopBacks := make([]string, 0)
	addresses := make([]string, 0)
	for _, iface := range ifaces {
		if len(preferredInterfaces) != 0 && !slices.Contains(preferredInterfaces, iface.Name) {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch typedAddr := addr.(type) {
			case *net.IPNet:
				ip = typedAddr.IP.To4()
			case *net.IPAddr:
				ip = typedAddr.IP.To4()
			default:
				continue
			}
			if ip == nil {
				continue
			}
			if ip.IsLoopback() {
				loopBacks = append(loopBacks, ip.String())
			} else {
				addresses = append(addresses, ip.String())
			}
		}
	}

	if includeLoopback {
		addresses = append(addresses, loopBacks...)
	}

	if len(addresses) > 0 {
		return addresses, nil
	}
	if len(loopBacks) > 0 {
		return loopBacks, nil
	}
	return nil, fmt.Errorf("could not find local IP address")
}

func GetLocalIPAddresses2(includeLoopback bool) (map[string][]string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	allAddresses := make(map[string][]string)
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		addresses := make([]string, 0)
		for _, addr := range addrs {
			var ip net.IP
			switch typedAddr := addr.(type) {
			case *net.IPNet:
				ip = typedAddr.IP.To4()
			case *net.IPAddr:
				ip = typedAddr.IP.To4()
			default:
				continue
			}
			if ip == nil {
				continue
			}

			if !ip.IsLoopback() || includeLoopback {
				addresses = append(addresses, ip.String())
			}
		}

		if len(addresses) > 0 {
			allAddresses[iface.Name] = addresses
		}
	}

	return allAddresses, nil
}

func FindMostSuitableIp(addresses map[string][]string, preferredInterfaces, excludedIpPrefixs []string) string {
	if preferredInterfaces != nil {
		for _, preferred := range preferredInterfaces {
			addrList, ok := addresses[preferred]
			if ok {
				for _, address := range addrList {
					excluded := false
					if excludedIpPrefixs != nil {
						for _, excludedIpPrefix := range excludedIpPrefixs {
							if strings.HasPrefix(address, excludedIpPrefix) {
								excluded = true
								break
							}
						}
					}

					if !excluded {
						return address
					}
				}
			}
		}
	}

	for _, addrList := range addresses {
		for _, address := range addrList {
			excluded := false
			if excludedIpPrefixs != nil {
				for _, excludedIpPrefix := range excludedIpPrefixs {
					if strings.HasPrefix(address, excludedIpPrefix) {
						excluded = true
						break
					}
				}
			}

			if !excluded {
				return address
			}
		}
	}

	for _, addrList := range addresses {
		for _, address := range addrList {
			return address
		}
	}

	return ""
}
