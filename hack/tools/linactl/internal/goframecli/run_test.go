// This file verifies linactl's embedded GoFrame CLI dispatch boundary.

package goframecli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"linactl/internal/toolrun"
)

func TestRunDispatchesHiddenCommandInTargetDir(t *testing.T) {
	root := t.TempDir()
	binary := filepath.Join(root, "bin", "linactl")
	targetDir := filepath.Join(root, "apps", "lina-plugins", "demo", "backend")

	for _, tc := range []struct {
		name string
		args []string
	}{
		{name: "ctrl", args: []string{"gen", "ctrl"}},
		{name: "dao", args: []string{"gen", "dao"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var (
				gotOptions toolrun.Options
				gotName    string
				gotArgs    []string
				calls      int
			)
			runner := func(_ context.Context, options toolrun.Options, name string, args ...string) error {
				calls++
				gotOptions = options
				gotName = name
				gotArgs = append([]string(nil), args...)
				return nil
			}

			err := Run(context.Background(), targetDir, func() (string, error) {
				return binary, nil
			}, runner, tc.args...)
			if err != nil {
				t.Fatalf("Run returned error: %v", err)
			}
			if calls != 1 {
				t.Fatalf("expected one runner call, got %d", calls)
			}
			if gotName != binary {
				t.Fatalf("runner name mismatch: got %q want %q", gotName, binary)
			}
			expectedArgs := append([]string{hiddenCommand}, tc.args...)
			if !reflect.DeepEqual(gotArgs, expectedArgs) {
				t.Fatalf("runner args mismatch: got %#v want %#v", gotArgs, expectedArgs)
			}
			if gotOptions.Dir != targetDir {
				t.Fatalf("runner dir mismatch: got %q want %q", gotOptions.Dir, targetDir)
			}
		})
	}
}

func TestResolveTargetDirAllowsControllerTargetWithoutHackConfig(t *testing.T) {
	root := t.TempDir()
	targetDir := filepath.Join(root, "apps", "lina-plugins", "demo", "backend")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}

	resolved, err := ResolveTargetDir(root, TargetOptions{PluginID: "demo", PluginIDSet: true})
	if err != nil {
		t.Fatalf("ResolveTargetDir returned error: %v", err)
	}
	if resolved != targetDir {
		t.Fatalf("target mismatch: got %q want %q", resolved, targetDir)
	}

	_, err = ResolveTargetDir(root, TargetOptions{PluginID: "demo", PluginIDSet: true, RequireConfig: true})
	if err == nil || !strings.Contains(err.Error(), "missing hack/config.yaml") {
		t.Fatalf("expected missing config error, got %v", err)
	}

	configPath := filepath.Join(targetDir, "hack", "config.yaml")
	if err = os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		t.Fatalf("mkdir config dir: %v", err)
	}
	if err = os.WriteFile(configPath, []byte("gfcli: {}\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err = ValidateTargetConfig(targetDir); err != nil {
		t.Fatalf("ValidateTargetConfig returned error: %v", err)
	}
}

func TestValidateArgsAllowsOnlyCodeGenerationCommands(t *testing.T) {
	for _, args := range [][]string{
		{"gen", "ctrl"},
		{"gen", "dao"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			if err := validateArgs(args); err != nil {
				t.Fatalf("validateArgs should allow %v: %v", args, err)
			}
		})
	}

	for _, args := range [][]string{
		{"install"},
		{"build"},
		{"gen", "service"},
		{"gen", "ctrl", "extra"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			err := validateArgs(args)
			if err == nil || !strings.Contains(err.Error(), "embedded GoFrame only supports gen ctrl or gen dao") {
				t.Fatalf("expected whitelist error for %v, got %v", args, err)
			}
		})
	}
}

func TestRunRejectsUnsupportedCommandsBeforeExecution(t *testing.T) {
	for _, args := range [][]string{
		{"install"},
		{"build"},
		{"gen", "service"},
		{"gen", "ctrl", "extra"},
		{"dao"},
	} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			err := Run(context.Background(), t.TempDir(), func() (string, error) {
				t.Fatalf("executable resolver should not run for invalid args")
				return "", nil
			}, func(context.Context, toolrun.Options, string, ...string) error {
				t.Fatalf("runner should not run for invalid args")
				return nil
			}, args...)
			if err == nil || !strings.Contains(err.Error(), "embedded GoFrame only supports gen ctrl or gen dao") {
				t.Fatalf("expected whitelist error, got %v", err)
			}
		})
	}
}

func TestRunReportsExecutableResolutionFailure(t *testing.T) {
	expectedErr := errors.New("no executable")
	err := Run(context.Background(), t.TempDir(), func() (string, error) {
		return "", expectedErr
	}, func(context.Context, toolrun.Options, string, ...string) error {
		t.Fatalf("runner should not run when executable resolution fails")
		return nil
	}, "gen", "ctrl")
	if err == nil || !strings.Contains(err.Error(), "resolve linactl executable") {
		t.Fatalf("expected executable resolution error, got %v", err)
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected wrapped executable error, got %v", err)
	}
}

func TestRunEmbeddedRejectsUnsupportedCommandsBeforeChangingDirectory(t *testing.T) {
	original, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve current directory: %v", err)
	}
	err = RunEmbedded(context.Background(), filepath.Join(t.TempDir(), "missing-root"), "install")
	if err == nil || !strings.Contains(err.Error(), "embedded GoFrame only supports gen ctrl or gen dao") {
		t.Fatalf("expected whitelist error, got %v", err)
	}
	current, err := filepath.Abs(".")
	if err != nil {
		t.Fatalf("resolve current directory after RunEmbedded: %v", err)
	}
	if current != original {
		t.Fatalf("RunEmbedded changed directory before validating args: got %q want %q", current, original)
	}
}

func TestConfigureGoFrameCLIAllowsCtrlWithoutHackConfig(t *testing.T) {
	original, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	targetDir := t.TempDir()
	if err = os.Chdir(targetDir); err != nil {
		t.Fatalf("chdir target: %v", err)
	}
	t.Cleanup(func() {
		if chdirErr := os.Chdir(original); chdirErr != nil {
			t.Fatalf("restore working directory: %v", chdirErr)
		}
	})

	cleanup, err := configureGoFrameCLI(false)
	if err != nil {
		t.Fatalf("configureGoFrameCLI for ctrl returned error: %v", err)
	}
	cleanup()

	_, err = configureGoFrameCLI(true)
	if err == nil || !strings.Contains(err.Error(), "missing hack/config.yaml") {
		t.Fatalf("expected dao config error, got %v", err)
	}
}
