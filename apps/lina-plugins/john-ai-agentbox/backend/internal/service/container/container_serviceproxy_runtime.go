// This file implements read-only Agent runtime service discovery on top of the
// plugin-labelled Docker runtime. It inspects only containers owned by the
// current AgentBox user and Agent, and it does not create proxy relays, bridges,
// or tunnels.

package container

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	dockercontainer "github.com/moby/moby/api/types/container"

	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
)

var _ serviceproxysvc.RuntimeBackend = (*dockerRuntimeBackend)(nil)

// RuntimeServices lists bounded TCP listen sockets for one visible Agent runtime.
func (b *dockerRuntimeBackend) RuntimeServices(ctx context.Context, userID string, agentID string) ([]serviceproxysvc.RuntimeServiceInfo, error) {
	inspected, err := b.inspectRunningAgentContainer(ctx, userID, agentID)
	if err != nil {
		return nil, err
	}
	output, err := b.runAgentCommand(ctx, userID, agentID, b.workspaceRootPath(), []string{
		"/bin/sh",
		"-c",
		agentServiceDiscoveryScript,
		"agentbox-service-discovery",
	})
	if err != nil {
		return nil, err
	}
	return parseAgentRuntimeServices(output, agentID, inspected, time.Now(), b.config.Service.RuntimeServiceListLimit)
}

const agentServiceDiscoveryScript = `set -eu
find_proc_name() {
  inode=$1
  for fd in /proc/[0-9]*/fd/[0-9]*; do
    target=$(readlink "$fd" 2>/dev/null || true)
    if [ "$target" = "socket:[$inode]" ]; then
      pid=${fd#/proc/}
      pid=${pid%%/*}
      name=$(cat "/proc/$pid/comm" 2>/dev/null || true)
      printf '%s\t%s\n' "$pid" "$name"
      return 0
    fi
  done
  printf '\t\n'
}
scan_file() {
  file=$1
  family=$2
  [ -r "$file" ] || return 0
  awk -v family="$family" 'NR > 1 && $4 == "0A" {
    split($2, local, ":")
    printf "%s\t%s\t%s\t%s\n", family, local[1], local[2], $10
  }' "$file"
}
{
  scan_file /proc/net/tcp tcp4
  scan_file /proc/net/tcp6 tcp6
} | while IFS="$(printf '\t')" read -r family address port_hex inode; do
  [ -n "$address" ] || continue
  proc=$(find_proc_name "$inode")
  pid=${proc%%	*}
  name=${proc#*	}
  printf '%s\t%s\t%s\t%s\t%s\n' "$family" "$address" "$port_hex" "$pid" "$name"
done`

type discoveredRuntimeService struct {
	family      string
	address     string
	port        int
	pid         string
	processName string
}

func parseAgentRuntimeServices(output string, agentID string, inspected dockercontainer.InspectResponse, checkedAt time.Time, limit int) ([]serviceproxysvc.RuntimeServiceInfo, error) {
	if limit <= 0 {
		limit = serviceproxysvc.DefaultRuntimeServiceListLimit
	}
	records := make([]discoveredRuntimeService, 0)
	for _, line := range strings.Split(output, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		record, err := parseAgentRuntimeServiceLine(line)
		if err != nil {
			return nil, err
		}
		if record.port <= 0 || record.port > 65535 {
			continue
		}
		records = append(records, record)
	}
	sort.SliceStable(records, func(i, j int) bool {
		if records[i].port != records[j].port {
			return records[i].port < records[j].port
		}
		return records[i].address < records[j].address
	})
	grouped := map[int]*serviceproxysvc.RuntimeServiceInfo{}
	ports := make([]int, 0)
	lastCheckedAt := checkedAt.UnixMilli()
	for _, record := range records {
		item, ok := grouped[record.port]
		if !ok {
			item = &serviceproxysvc.RuntimeServiceInfo{
				ID:            fmt.Sprintf("svc-%d", record.port),
				AgentID:       strings.TrimSpace(agentID),
				Port:          record.port,
				Protocol:      serviceproxysvc.AgentServiceProtocolUnknown,
				AccessStatus:  serviceproxysvc.AgentServiceAccessUnavailable,
				PID:           record.pid,
				ProcessName:   record.processName,
				LastCheckedAt: lastCheckedAt,
			}
			grouped[record.port] = item
			ports = append(ports, record.port)
		}
		if item.PID == "" && record.pid != "" {
			item.PID = record.pid
		}
		if item.ProcessName == "" && record.processName != "" {
			item.ProcessName = record.processName
		}
		item.ListenAddresses = append(item.ListenAddresses, listenAddressFromDiscovered(record, inspected))
	}
	sort.Ints(ports)
	out := make([]serviceproxysvc.RuntimeServiceInfo, 0, len(ports))
	for _, port := range ports {
		item := grouped[port]
		item.AccessStatus = aggregateServiceAccessStatus(item.ListenAddresses)
		item.UnavailableReason = serviceUnavailableReason(item.ListenAddresses)
		out = append(out, *item)
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

func parseAgentRuntimeServiceLine(line string) (discoveredRuntimeService, error) {
	parts := strings.Split(line, "\t")
	if len(parts) != 5 {
		return discoveredRuntimeService{}, gerror.New("agent service discovery output is invalid")
	}
	port64, err := strconv.ParseUint(parts[2], 16, 16)
	if err != nil {
		return discoveredRuntimeService{}, gerror.Wrap(err, "parse agent service port")
	}
	address, err := parseProcNetAddress(parts[0], parts[1])
	if err != nil {
		return discoveredRuntimeService{}, err
	}
	return discoveredRuntimeService{
		family:      strings.TrimSpace(parts[0]),
		address:     address,
		port:        int(port64),
		pid:         strings.TrimSpace(parts[3]),
		processName: strings.TrimSpace(parts[4]),
	}, nil
}

func parseProcNetAddress(family string, hexAddress string) (string, error) {
	hexAddress = strings.TrimSpace(hexAddress)
	if family == "tcp4" {
		if len(hexAddress) != 8 {
			return "", gerror.New("agent service IPv4 address is invalid")
		}
		bytes := make([]byte, 4)
		for i := 0; i < 4; i++ {
			value, err := strconv.ParseUint(hexAddress[i*2:i*2+2], 16, 8)
			if err != nil {
				return "", gerror.Wrap(err, "parse agent service IPv4 address")
			}
			bytes[3-i] = byte(value)
		}
		return net.IP(bytes).String(), nil
	}
	if family == "tcp6" {
		if len(hexAddress) != 32 {
			return "", gerror.New("agent service IPv6 address is invalid")
		}
		bytes := make([]byte, 16)
		for word := 0; word < 4; word++ {
			base := word * 8
			for i := 0; i < 4; i++ {
				value, err := strconv.ParseUint(hexAddress[base+i*2:base+i*2+2], 16, 8)
				if err != nil {
					return "", gerror.Wrap(err, "parse agent service IPv6 address")
				}
				bytes[word*4+3-i] = byte(value)
			}
		}
		return net.IP(bytes).String(), nil
	}
	return "", gerror.New("agent service network family is invalid")
}

func listenAddressFromDiscovered(record discoveredRuntimeService, inspected dockercontainer.InspectResponse) serviceproxysvc.ListenAddress {
	status := serviceproxysvc.AgentServiceAccessBridgeRequired
	reason := ""
	if serviceAddressDirect(record.address, inspected) {
		status = serviceproxysvc.AgentServiceAccessDirect
	} else {
		reason = "loopback listener requires a bridge before proxy or tunnel access"
	}
	return serviceproxysvc.ListenAddress{
		Address:           record.address,
		Port:              record.port,
		Network:           record.family,
		AccessStatus:      status,
		UnavailableReason: reason,
	}
}

func serviceAddressDirect(address string, inspected dockercontainer.InspectResponse) bool {
	ip := net.ParseIP(strings.TrimSpace(address))
	if ip == nil {
		return false
	}
	if ip.IsUnspecified() {
		return true
	}
	if ip.IsLoopback() {
		return false
	}
	if inspected.NetworkSettings == nil {
		return true
	}
	for _, network := range inspected.NetworkSettings.Networks {
		if network == nil {
			continue
		}
		if network.IPAddress.IsValid() && ip.Equal(net.IP(network.IPAddress.AsSlice())) {
			return true
		}
		if network.GlobalIPv6Address.IsValid() && ip.Equal(net.IP(network.GlobalIPv6Address.AsSlice())) {
			return true
		}
	}
	return true
}

func aggregateServiceAccessStatus(addresses []serviceproxysvc.ListenAddress) string {
	if len(addresses) == 0 {
		return serviceproxysvc.AgentServiceAccessUnavailable
	}
	for _, address := range addresses {
		if address.AccessStatus == serviceproxysvc.AgentServiceAccessDirect {
			return serviceproxysvc.AgentServiceAccessDirect
		}
	}
	return addresses[0].AccessStatus
}

func serviceUnavailableReason(addresses []serviceproxysvc.ListenAddress) string {
	for _, address := range addresses {
		if address.UnavailableReason != "" {
			return address.UnavailableReason
		}
	}
	return ""
}
