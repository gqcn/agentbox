// This file verifies read-only Agent runtime service discovery parsing without
// requiring a live Docker daemon.

package container

import (
	"net/netip"
	"testing"
	"time"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"

	serviceproxysvc "john-ai-agentbox/backend/internal/service/serviceproxy"
)

// TestParseAgentRuntimeServicesGroupsPorts verifies /proc/net listeners are grouped into bounded service projections.
func TestParseAgentRuntimeServicesGroupsPorts(t *testing.T) {
	output := "" +
		"tcp4\t00000000\t1770\t123\tvite\n" +
		"tcp4\t0100007F\t0BB8\t124\tnode\n"
	inspected := dockercontainer.InspectResponse{}
	items, err := parseAgentRuntimeServices(output, "agt-owned", inspected, time.Unix(10, 0), serviceproxysvc.DefaultRuntimeServiceListLimit)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("expected two services, got %#v", items)
	}
	if items[0].ID != "svc-3000" || items[0].AccessStatus != serviceproxysvc.AgentServiceAccessBridgeRequired {
		t.Fatalf("unexpected loopback service: %#v", items[0])
	}
	if items[0].ListenAddresses[0].UnavailableReason == "" || items[0].ProxyURL != "" || items[0].TunnelURL != "" {
		t.Fatalf("loopback service should require bridge without relay URLs: %#v", items[0])
	}
	if items[1].ID != "svc-6000" || items[1].AccessStatus != serviceproxysvc.AgentServiceAccessDirect {
		t.Fatalf("unexpected direct service: %#v", items[1])
	}
	if items[1].LastCheckedAt != time.Unix(10, 0).UnixMilli() {
		t.Fatalf("unexpected check time: %#v", items[1])
	}
}

// TestParseAgentRuntimeServicesRecognizesContainerNetworkAddress verifies container IP listeners are direct.
func TestParseAgentRuntimeServicesRecognizesContainerNetworkAddress(t *testing.T) {
	containerIP := netip.MustParseAddr("172.18.0.4")
	inspected := dockercontainer.InspectResponse{
		NetworkSettings: &dockercontainer.NetworkSettings{
			Networks: map[string]*network.EndpointSettings{
				"bridge": {IPAddress: containerIP},
			},
		},
	}
	items, err := parseAgentRuntimeServices("tcp4\t040012AC\t1F90\t321\tserver\n", "agt-owned", inspected, time.Unix(10, 0), serviceproxysvc.DefaultRuntimeServiceListLimit)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].AccessStatus != serviceproxysvc.AgentServiceAccessDirect {
		t.Fatalf("expected container network listener to be direct, got %#v", items)
	}
}

// TestParseProcNetAddress verifies Linux little-endian IPv4 parsing.
func TestParseProcNetAddress(t *testing.T) {
	address, err := parseProcNetAddress("tcp4", "0100007F")
	if err != nil {
		t.Fatal(err)
	}
	if address != "127.0.0.1" {
		t.Fatalf("unexpected parsed address %q", address)
	}
}
