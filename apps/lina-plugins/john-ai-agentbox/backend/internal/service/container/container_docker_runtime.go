// This file adapts the Docker SDK to the safe AgentBox container lifecycle
// subset. It never lists or mutates arbitrary Docker containers: every operation
// is constrained by plugin-owned labels and the current AgentBox user label.

package container

import (
	"bytes"
	"context"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/moby/moby/api/pkg/stdcopy"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"
)

type dockerRuntimeBackend struct {
	client *dockerclient.Client
	config RuntimeConfig
	err    error
}

var (
	_ DockerHealthBackend       = (*dockerRuntimeBackend)(nil)
	_ ContainerLifecycleBackend = (*dockerRuntimeBackend)(nil)
)

func newDockerRuntimeBackend(config RuntimeConfig) *dockerRuntimeBackend {
	options := []dockerclient.Opt{
		dockerclient.WithTLSClientConfigFromEnv(),
		dockerclient.WithAPIVersionNegotiation(),
	}
	if config.Host == "" {
		options = append(options, dockerclient.WithHostFromEnv())
	} else {
		options = append(options, dockerclient.WithHost(config.Host))
	}
	cli, err := dockerclient.NewClientWithOpts(options...)
	return &dockerRuntimeBackend{client: cli, config: config, err: err}
}

// Ping verifies the Docker daemon and returns public health metadata.
func (b *dockerRuntimeBackend) Ping(ctx context.Context) (*DockerHealthResponse, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	ping, err := cli.Ping(ctx, dockerclient.PingOptions{})
	if err != nil {
		return nil, gerror.Wrap(err, "ping docker daemon")
	}
	return &DockerHealthResponse{
		OK:         true,
		APIVersion: ping.APIVersion,
		OSType:     ping.OSType,
	}, nil
}

// List returns plugin-managed containers owned by userID.
func (b *dockerRuntimeBackend) List(ctx context.Context, userID string) ([]ContainerInfo, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	items, err := b.listVisibleSummaries(ctx, cli, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ContainerInfo, 0, len(items))
	for _, item := range items {
		out = append(out, b.containerInfoFromSummary(item))
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].CreatedAt != out[j].CreatedAt {
			return out[i].CreatedAt > out[j].CreatedAt
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

// Inspect returns one plugin-managed container owned by userID.
func (b *dockerRuntimeBackend) Inspect(ctx context.Context, userID string, containerID string) (*ContainerInfo, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectVisibleContainer(ctx, cli, userID, containerID)
	if err != nil {
		return nil, err
	}
	item := b.containerInfoFromInspect(inspected)
	return &item, nil
}

// Start starts one plugin-managed container owned by userID.
func (b *dockerRuntimeBackend) Start(ctx context.Context, userID string, containerID string) (*ContainerInfo, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectVisibleContainer(ctx, cli, userID, containerID)
	if err != nil {
		return nil, err
	}
	if _, err := cli.ContainerStart(ctx, inspected.ID, dockerclient.ContainerStartOptions{}); err != nil {
		return nil, dockerActionError(err, "start agentbox container")
	}
	return b.Inspect(ctx, userID, inspected.ID)
}

// Stop stops one plugin-managed container owned by userID.
func (b *dockerRuntimeBackend) Stop(ctx context.Context, userID string, containerID string) (*ContainerInfo, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectVisibleContainer(ctx, cli, userID, containerID)
	if err != nil {
		return nil, err
	}
	timeout := b.stopTimeoutSeconds()
	if _, err := cli.ContainerStop(ctx, inspected.ID, dockerclient.ContainerStopOptions{Timeout: &timeout}); err != nil {
		return nil, dockerActionError(err, "stop agentbox container")
	}
	return b.Inspect(ctx, userID, inspected.ID)
}

// Delete removes one plugin-managed container owned by userID.
func (b *dockerRuntimeBackend) Delete(ctx context.Context, userID string, containerID string) (bool, error) {
	cli, err := b.requireClient()
	if err != nil {
		return false, err
	}
	inspected, err := b.inspectVisibleContainer(ctx, cli, userID, containerID)
	if err != nil {
		return false, err
	}
	_, err = cli.ContainerRemove(ctx, inspected.ID, dockerclient.ContainerRemoveOptions{
		Force:         true,
		RemoveVolumes: false,
	})
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return true, nil
		}
		return false, gerror.Wrap(err, "delete agentbox container")
	}
	return true, nil
}

// Logs returns recent logs for one plugin-managed container owned by userID.
func (b *dockerRuntimeBackend) Logs(ctx context.Context, userID string, containerID string) (*LogsResponse, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectVisibleContainer(ctx, cli, userID, containerID)
	if err != nil {
		return nil, err
	}
	reader, err := cli.ContainerLogs(ctx, inspected.ID, dockerclient.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       b.logTail(),
	})
	if err != nil {
		return nil, dockerActionError(err, "read agentbox container logs")
	}
	defer reader.Close()

	var buf bytes.Buffer
	if inspected.Config != nil && inspected.Config.Tty {
		if _, err := io.Copy(&buf, reader); err != nil {
			return nil, gerror.Wrap(err, "copy agentbox container logs")
		}
		return &LogsResponse{Logs: buf.String()}, nil
	}
	if _, err := stdcopy.StdCopy(&buf, &buf, reader); err != nil {
		return nil, gerror.Wrap(err, "copy agentbox container logs")
	}
	return &LogsResponse{Logs: buf.String()}, nil
}

func (b *dockerRuntimeBackend) requireClient() (*dockerclient.Client, error) {
	if b == nil {
		return nil, gerror.New("docker runtime backend is unavailable")
	}
	if b.err != nil {
		return nil, gerror.Wrap(b.err, "create docker client")
	}
	if b.client == nil {
		return nil, gerror.New("docker client is unavailable")
	}
	return b.client, nil
}

func (b *dockerRuntimeBackend) listVisibleSummaries(ctx context.Context, cli *dockerclient.Client, userID string) ([]dockercontainer.Summary, error) {
	filters := make(dockerclient.Filters).
		Add("label", managedContainerLabelFilter()).
		Add("label", userContainerLabelFilter(userID)).
		Add("label", containerLabelContainerID)
	result, err := cli.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, gerror.Wrap(err, "list agentbox containers")
	}
	items := make([]dockercontainer.Summary, 0, len(result.Items))
	for _, item := range result.Items {
		if isIndependentManagedContainer(item.Labels, userID) {
			items = append(items, item)
		}
	}
	return items, nil
}

func (b *dockerRuntimeBackend) inspectVisibleContainer(ctx context.Context, cli *dockerclient.Client, userID string, containerID string) (dockercontainer.InspectResponse, error) {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return dockercontainer.InspectResponse{}, newInvalidInputError()
	}
	inspected, err := cli.ContainerInspect(ctx, containerID, dockerclient.ContainerInspectOptions{})
	if err == nil {
		if inspectVisibleToUser(inspected.Container, userID) {
			return inspected.Container, nil
		}
		return dockercontainer.InspectResponse{}, newContainerNotFoundError()
	}
	if !cerrdefs.IsNotFound(err) {
		return dockercontainer.InspectResponse{}, gerror.Wrap(err, "inspect agentbox container")
	}

	items, err := b.listVisibleSummaries(ctx, cli, userID)
	if err != nil {
		return dockercontainer.InspectResponse{}, err
	}
	for _, item := range items {
		if summaryMatchesContainerID(item, containerID) {
			inspected, err := cli.ContainerInspect(ctx, item.ID, dockerclient.ContainerInspectOptions{})
			if err != nil {
				return dockercontainer.InspectResponse{}, dockerActionError(err, "inspect agentbox container")
			}
			if inspectVisibleToUser(inspected.Container, userID) {
				return inspected.Container, nil
			}
			break
		}
	}
	return dockercontainer.InspectResponse{}, newContainerNotFoundError()
}

func dockerActionError(err error, message string) error {
	if err == nil {
		return nil
	}
	if cerrdefs.IsNotFound(err) {
		return newContainerNotFoundError()
	}
	return gerror.Wrap(err, message)
}

func inspectVisibleToUser(item dockercontainer.InspectResponse, userID string) bool {
	if item.Config == nil {
		return false
	}
	return isIndependentManagedContainer(item.Config.Labels, userID)
}

func summaryMatchesContainerID(item dockercontainer.Summary, containerID string) bool {
	containerID = strings.TrimSpace(containerID)
	if containerID == "" {
		return false
	}
	if item.ID == containerID || labelValue(item.Labels, containerLabelContainerID, "") == containerID {
		return true
	}
	return false
}

func (b *dockerRuntimeBackend) containerInfoFromSummary(item dockercontainer.Summary) ContainerInfo {
	createdAt := int64(0)
	if item.Created > 0 {
		createdAt = time.Unix(item.Created, 0).UnixMilli()
	}
	return ContainerInfo{
		ID:        labelValue(item.Labels, containerLabelContainerID, item.ID),
		Name:      labelValue(item.Labels, containerLabelName, displayName(item.Names)),
		DockerID:  item.ID,
		Image:     item.Image,
		State:     string(item.State),
		Status:    item.Status,
		CreatedAt: createdAt,
		Mounts:    mountInfoFromDocker(item.Mounts),
		Labels:    copyLabels(item.Labels),
		Workspace: b.workspaceMountPath(item.Mounts),
	}
}

func (b *dockerRuntimeBackend) containerInfoFromInspect(item dockercontainer.InspectResponse) ContainerInfo {
	labels := map[string]string{}
	image := item.Image
	if item.Config != nil {
		labels = item.Config.Labels
		image = item.Config.Image
	}
	state := ""
	status := ""
	if item.State != nil {
		state = string(item.State.Status)
		status = string(item.State.Status)
	}
	return ContainerInfo{
		ID:        labelValue(labels, containerLabelContainerID, item.ID),
		Name:      labelValue(labels, containerLabelName, strings.TrimPrefix(item.Name, "/")),
		DockerID:  item.ID,
		Image:     image,
		State:     state,
		Status:    status,
		CreatedAt: dockerCreatedAtMilliseconds(item.Created),
		Mounts:    mountInfoFromDocker(item.Mounts),
		Labels:    copyLabels(labels),
		Workspace: b.workspaceMountPath(item.Mounts),
	}
}

func mountInfoFromDocker(items []dockercontainer.MountPoint) []MountInfo {
	out := make([]MountInfo, 0, len(items))
	for _, item := range items {
		out = append(out, MountInfo{
			Type:        string(item.Type),
			Source:      item.Source,
			Destination: item.Destination,
			Name:        item.Name,
		})
	}
	return out
}

func (b *dockerRuntimeBackend) workspaceMountPath(items []dockercontainer.MountPoint) string {
	workspaceRoot := b.workspaceRootPath()
	for _, item := range items {
		if item.Destination == workspaceRoot {
			return item.Destination
		}
	}
	return ""
}

func (b *dockerRuntimeBackend) workspaceRootPath() string {
	if b == nil || b.config.Workspace.WorkspaceRootPath == "" {
		return DefaultRuntimeConfig().Workspace.WorkspaceRootPath
	}
	return b.config.Workspace.WorkspaceRootPath
}

func (b *dockerRuntimeBackend) sharedRootPath() string {
	if b == nil || b.config.Workspace.SharedRootPath == "" {
		return DefaultRuntimeConfig().Workspace.SharedRootPath
	}
	return b.config.Workspace.SharedRootPath
}

func (b *dockerRuntimeBackend) logTail() string {
	if b == nil || b.config.ContainerLogTail <= 0 {
		return strconv.Itoa(DefaultRuntimeConfig().ContainerLogTail)
	}
	return strconv.Itoa(b.config.ContainerLogTail)
}

func (b *dockerRuntimeBackend) stopTimeoutSeconds() int {
	if b == nil || b.config.StopTimeout <= 0 {
		return int(DefaultRuntimeConfig().StopTimeout / time.Second)
	}
	seconds := int(math.Ceil(b.config.StopTimeout.Seconds()))
	if seconds < 1 {
		return 1
	}
	return seconds
}

func displayName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

func dockerCreatedAtMilliseconds(value string) int64 {
	if strings.TrimSpace(value) == "" {
		return 0
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return 0
	}
	return parsed.UnixMilli()
}
