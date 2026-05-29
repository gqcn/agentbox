// This file executes the embedded GoFrame CLI in the hidden child command.

package goframecli

import (
	"context"
	"fmt"
	"os"

	"github.com/gogf/gf/cmd/gf/v2/gfcmd"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
)

// RunEmbedded executes a whitelisted GoFrame code generation command in the
// current process. Callers should use it only from linactl's hidden child
// command so GoFrame CLI exits cannot terminate the parent linactl process.
func RunEmbedded(ctx context.Context, root string, args ...string) error {
	if err := validateArgs(args); err != nil {
		return err
	}
	cleanup, err := configureGoFrameCLI(args[1] == "dao")
	if err != nil {
		return err
	}
	defer cleanup()
	command, err := gfcmd.GetCommand(ctx)
	if err != nil {
		return fmt.Errorf("initialize embedded GoFrame CLI: %w", err)
	}
	if command == nil {
		return fmt.Errorf("initialize embedded GoFrame CLI: command is nil")
	}
	_, err = command.RunWithSpecificArgs(ctx, append([]string{"gf"}, args...))
	if err != nil {
		return fmt.Errorf("run embedded GoFrame %s %s: %w", args[0], args[1], err)
	}
	return nil
}

// configureGoFrameCLI mirrors the gf CLI's hack/config.yaml lookup without
// calling gfcmd.Command.Run, whose error path can exit the process.
func configureGoFrameCLI(requireConfig bool) (func(), error) {
	adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile)
	if !ok {
		return func() {}, nil
	}
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("resolve embedded GoFrame working directory: %w", err)
	}
	if err := ValidateProjectDir(workingDir); err != nil {
		return nil, err
	}
	if requireConfig {
		if err := ValidateTargetConfig(workingDir); err != nil {
			return nil, err
		}
	}
	configPath := "hack"
	cleanup := func() {}
	if !requireConfig {
		if _, err := os.Stat("hack/config.yaml"); err != nil {
			if os.IsNotExist(err) {
				tempDir, mkErr := os.MkdirTemp("", "linactl-goframe-config-*")
				if mkErr != nil {
					return nil, fmt.Errorf("create temporary embedded GoFrame config directory: %w", mkErr)
				}
				configPath = tempDir
				cleanup = func() {
					if removeErr := os.RemoveAll(tempDir); removeErr != nil {
						fmt.Fprintf(os.Stderr, "warning: remove temporary embedded GoFrame config directory %s: %v\n", tempDir, removeErr)
					}
				}
			} else {
				return nil, fmt.Errorf("check embedded GoFrame CLI config: %w", err)
			}
		}
	}
	if err := adapter.SetPath(configPath); err != nil {
		cleanup()
		return nil, fmt.Errorf("configure embedded GoFrame CLI: %w", err)
	}
	return cleanup, nil
}
