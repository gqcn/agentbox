// Package dockerhealth owns the narrow Docker SDK integration used by AgentBox
// runtime health checks. It does not expose Docker client handles to the parent
// service and does not manage Agent container lifecycle state.
package dockerhealth

import (
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/moby/moby/client"
)

// Info is the Docker daemon metadata exposed by a successful health probe.
type Info struct {
	APIVersion string
	OSType     string
}

// Runtime defines the minimal Docker health probe contract.
type Runtime interface {
	// Ping checks Docker daemon reachability. Errors are low-level causes that
	// callers wrap into plugin business errors before returning HTTP responses.
	Ping(ctx context.Context) (Info, error)
}

// runtimeImpl is the Docker SDK backed health runtime.
type runtimeImpl struct {
	client *client.Client
	err    error
}

var _ Runtime = (*runtimeImpl)(nil)

// New creates the Docker health runtime. Docker client construction failures
// are retained as runtime-unavailable causes instead of failing plugin startup.
func New() Runtime {
	cli, err := client.NewClientWithOpts(
		client.WithTLSClientConfigFromEnv(),
		client.WithHostFromEnv(),
		client.WithAPIVersionNegotiation(),
	)
	return &runtimeImpl{client: cli, err: err}
}

// Ping checks Docker daemon reachability using the negotiated client.
func (r *runtimeImpl) Ping(ctx context.Context) (Info, error) {
	if r == nil {
		return Info{}, gerror.New("docker health runtime is unavailable")
	}
	if r.err != nil {
		return Info{}, gerror.Wrap(r.err, "create docker client")
	}
	if r.client == nil {
		return Info{}, gerror.New("docker client is unavailable")
	}
	ping, err := r.client.Ping(ctx, client.PingOptions{})
	if err != nil {
		return Info{}, gerror.Wrap(err, "ping docker daemon")
	}
	return Info{
		APIVersion: ping.APIVersion,
		OSType:     ping.OSType,
	}, nil
}
