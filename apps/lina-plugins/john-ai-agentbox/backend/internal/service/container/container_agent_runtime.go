// This file implements the minimal trusted Agent runtime Docker lifecycle.
// It creates only plugin-labelled long-lived Agent containers from persisted
// Agent image data; arbitrary Docker create options remain unavailable.

package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/moby/moby/api/pkg/stdcopy"
	dockercontainer "github.com/moby/moby/api/types/container"
	dockerclient "github.com/moby/moby/client"

	catalogsvc "john-ai-agentbox/backend/internal/service/catalog"
)

const (
	agentRuntimeStatusRunning = "running"
	agentRuntimeStatusStopped = "stopped"
)

var nonContainerNameChar = regexp.MustCompile(`[^a-z0-9-]+`)

var _ catalogsvc.AgentRuntimeBackend = (*dockerRuntimeBackend)(nil)

// CreateAgentRuntime creates and starts one plugin-managed Agent runtime container.
func (b *dockerRuntimeBackend) CreateAgentRuntime(ctx context.Context, input catalogsvc.AgentRuntimeContainerInput) (*catalogsvc.AgentRuntimeContainerInfo, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	input, err = normalizeAgentRuntimeInput(input)
	if err != nil {
		return nil, err
	}
	if err := b.removeExistingAgentContainers(ctx, cli, input.UserID, input.AgentID); err != nil {
		return nil, err
	}
	labels := agentRuntimeLabels(input)
	resp, err := cli.ContainerCreate(ctx, dockerclient.ContainerCreateOptions{
		Config: &dockercontainer.Config{
			Image:      input.ImageRef,
			Labels:     labels,
			WorkingDir: b.workspaceRootPath(),
			Tty:        true,
			OpenStdin:  true,
			Cmd:        []string{"sleep", "infinity"},
		},
		HostConfig: &dockercontainer.HostConfig{},
		Name:       agentRuntimeContainerName(input.AgentID),
	})
	if err != nil {
		return nil, dockerActionError(err, "create agentbox agent runtime container")
	}
	if _, err := cli.ContainerStart(ctx, resp.ID, dockerclient.ContainerStartOptions{}); err != nil {
		if _, removeErr := cli.ContainerRemove(ctx, resp.ID, dockerclient.ContainerRemoveOptions{Force: true, RemoveVolumes: false}); removeErr != nil && !cerrdefs.IsNotFound(removeErr) {
			return nil, gerror.Wrapf(err, "remove failed agent runtime container after start failure: %v", removeErr)
		}
		return nil, dockerActionError(err, "start agentbox agent runtime container")
	}
	return b.agentRuntimeInfo(ctx, cli, input.UserID, input.AgentID, resp.ID)
}

// StartAgentRuntime starts an existing plugin-managed Agent runtime container.
func (b *dockerRuntimeBackend) StartAgentRuntime(ctx context.Context, userID string, agentID string, containerID string) (*catalogsvc.AgentRuntimeContainerInfo, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectVisibleAgentContainer(ctx, cli, userID, agentID, containerID)
	if err != nil {
		return nil, err
	}
	if inspected.State == nil || !inspected.State.Running {
		if _, err := cli.ContainerStart(ctx, inspected.ID, dockerclient.ContainerStartOptions{}); err != nil {
			return nil, dockerActionError(err, "start agentbox agent runtime container")
		}
	}
	return b.agentRuntimeInfo(ctx, cli, userID, agentID, inspected.ID)
}

// StopAgentRuntime stops an existing plugin-managed Agent runtime container.
func (b *dockerRuntimeBackend) StopAgentRuntime(ctx context.Context, userID string, agentID string, containerID string) (*catalogsvc.AgentRuntimeContainerInfo, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectVisibleAgentContainer(ctx, cli, userID, agentID, containerID)
	if err != nil {
		return nil, err
	}
	if inspected.State != nil && inspected.State.Running {
		timeout := b.stopTimeoutSeconds()
		if _, err := cli.ContainerStop(ctx, inspected.ID, dockerclient.ContainerStopOptions{Timeout: &timeout}); err != nil {
			return nil, dockerActionError(err, "stop agentbox agent runtime container")
		}
	}
	return b.agentRuntimeInfo(ctx, cli, userID, agentID, inspected.ID)
}

// AgentRuntimeLogs returns recent logs for an existing plugin-managed Agent runtime container.
func (b *dockerRuntimeBackend) AgentRuntimeLogs(ctx context.Context, userID string, agentID string, containerID string) (*catalogsvc.AgentRuntimeLogsOutput, error) {
	cli, err := b.requireClient()
	if err != nil {
		return nil, err
	}
	inspected, err := b.inspectVisibleAgentContainer(ctx, cli, userID, agentID, containerID)
	if err != nil {
		return nil, err
	}
	reader, err := cli.ContainerLogs(ctx, inspected.ID, dockerclient.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       b.logTail(),
	})
	if err != nil {
		return nil, dockerActionError(err, "read agentbox agent runtime logs")
	}
	defer reader.Close()

	var buf bytes.Buffer
	if inspected.Config != nil && inspected.Config.Tty {
		if _, err := io.Copy(&buf, reader); err != nil {
			return nil, gerror.Wrap(err, "copy agentbox agent runtime logs")
		}
		return &catalogsvc.AgentRuntimeLogsOutput{Logs: buf.String()}, nil
	}
	if _, err := stdcopy.StdCopy(&buf, &buf, reader); err != nil {
		return nil, gerror.Wrap(err, "copy agentbox agent runtime logs")
	}
	return &catalogsvc.AgentRuntimeLogsOutput{Logs: buf.String()}, nil
}

func (b *dockerRuntimeBackend) removeExistingAgentContainers(ctx context.Context, cli *dockerclient.Client, userID string, agentID string) error {
	items, err := b.listAgentContainerSummaries(ctx, cli, userID, agentID)
	if err != nil {
		return err
	}
	for _, item := range items {
		if _, err := cli.ContainerRemove(ctx, item.ID, dockerclient.ContainerRemoveOptions{Force: true, RemoveVolumes: false}); err != nil && !cerrdefs.IsNotFound(err) {
			return dockerActionError(err, "delete previous agentbox agent runtime container")
		}
	}
	return nil
}

func (b *dockerRuntimeBackend) agentRuntimeInfo(ctx context.Context, cli *dockerclient.Client, userID string, agentID string, containerID string) (*catalogsvc.AgentRuntimeContainerInfo, error) {
	inspected, err := b.inspectVisibleAgentContainer(ctx, cli, userID, agentID, containerID)
	if err != nil {
		return nil, err
	}
	status := agentRuntimeStatusStopped
	if inspected.State != nil && inspected.State.Running {
		status = agentRuntimeStatusRunning
	}
	labels := map[string]string{}
	if inspected.Config != nil {
		labels = inspected.Config.Labels
	}
	return &catalogsvc.AgentRuntimeContainerInfo{
		ContainerID: labelValue(labels, containerLabelContainerID, agentID),
		DockerID:    inspected.ID,
		Status:      status,
	}, nil
}

func (b *dockerRuntimeBackend) listAgentContainerSummaries(ctx context.Context, cli *dockerclient.Client, userID string, agentID string) ([]dockercontainer.Summary, error) {
	filters := make(dockerclient.Filters).
		Add("label", managedContainerLabelFilter()).
		Add("label", userContainerLabelFilter(userID)).
		Add("label", containerLabelAgentID+"="+strings.TrimSpace(agentID))
	result, err := cli.ContainerList(ctx, dockerclient.ContainerListOptions{
		All:     true,
		Filters: filters,
	})
	if err != nil {
		return nil, dockerActionError(err, "list agentbox agent runtime containers")
	}
	items := make([]dockercontainer.Summary, 0, len(result.Items))
	for _, item := range result.Items {
		if isAgentManagedContainer(item.Labels, userID, agentID) {
			items = append(items, item)
		}
	}
	return items, nil
}

func (b *dockerRuntimeBackend) inspectVisibleAgentContainer(ctx context.Context, cli *dockerclient.Client, userID string, agentID string, containerID string) (dockercontainer.InspectResponse, error) {
	userID = strings.TrimSpace(userID)
	agentID = strings.TrimSpace(agentID)
	containerID = strings.TrimSpace(containerID)
	if userID == "" || agentID == "" {
		return dockercontainer.InspectResponse{}, newInvalidInputError()
	}
	if containerID != "" {
		inspected, err := cli.ContainerInspect(ctx, containerID, dockerclient.ContainerInspectOptions{})
		if err == nil {
			if inspectAgentVisibleToUser(inspected.Container, userID, agentID) {
				return inspected.Container, nil
			}
			return dockercontainer.InspectResponse{}, newContainerNotFoundError()
		}
		if !cerrdefs.IsNotFound(err) {
			return dockercontainer.InspectResponse{}, dockerActionError(err, "inspect agentbox agent runtime container")
		}
	}
	items, err := b.listAgentContainerSummaries(ctx, cli, userID, agentID)
	if err != nil {
		return dockercontainer.InspectResponse{}, err
	}
	for _, item := range items {
		if containerID == "" || summaryMatchesAgentContainerID(item, agentID, containerID) {
			inspected, err := cli.ContainerInspect(ctx, item.ID, dockerclient.ContainerInspectOptions{})
			if err != nil {
				return dockercontainer.InspectResponse{}, dockerActionError(err, "inspect agentbox agent runtime container")
			}
			if inspectAgentVisibleToUser(inspected.Container, userID, agentID) {
				return inspected.Container, nil
			}
		}
	}
	return dockercontainer.InspectResponse{}, newContainerNotFoundError()
}

func normalizeAgentRuntimeInput(input catalogsvc.AgentRuntimeContainerInput) (catalogsvc.AgentRuntimeContainerInput, error) {
	input.AgentID = strings.TrimSpace(input.AgentID)
	input.UserID = strings.TrimSpace(input.UserID)
	input.Name = strings.TrimSpace(input.Name)
	input.ImageRef = strings.TrimSpace(input.ImageRef)
	input.AgentType = strings.TrimSpace(input.AgentType)
	if input.AgentID == "" || input.UserID == "" || input.ImageRef == "" || input.ImageID <= 0 {
		return catalogsvc.AgentRuntimeContainerInput{}, newInvalidInputError()
	}
	if input.Name == "" {
		input.Name = input.AgentID
	}
	return input, nil
}

func agentRuntimeLabels(input catalogsvc.AgentRuntimeContainerInput) map[string]string {
	return map[string]string{
		containerLabelManaged:     containerManagedValue,
		containerLabelUser:        input.UserID,
		containerLabelContainerID: input.AgentID,
		containerLabelName:        input.Name,
		containerLabelAgentID:     input.AgentID,
		containerLabelAgentType:   input.AgentType,
		containerLabelImageID:     fmt.Sprintf("%d", input.ImageID),
	}
}

func inspectAgentVisibleToUser(item dockercontainer.InspectResponse, userID string, agentID string) bool {
	return item.Config != nil && isAgentManagedContainer(item.Config.Labels, userID, agentID)
}

func summaryMatchesAgentContainerID(item dockercontainer.Summary, agentID string, containerID string) bool {
	if containerID == "" {
		return true
	}
	if item.ID == containerID {
		return true
	}
	if labelValue(item.Labels, containerLabelContainerID, "") == containerID {
		return true
	}
	return labelValue(item.Labels, containerLabelAgentID, "") == agentID && containerID == agentID
}

func agentRuntimeContainerName(agentID string) string {
	return "agentbox-" + slugContainerName(agentID)
}

func slugContainerName(input string) string {
	value := strings.ToLower(strings.TrimSpace(input))
	value = strings.ReplaceAll(value, "_", "-")
	value = nonContainerNameChar.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-")
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	if value == "" {
		return "agent"
	}
	if len(value) > 42 {
		value = strings.Trim(value[:42], "-")
	}
	return value
}
