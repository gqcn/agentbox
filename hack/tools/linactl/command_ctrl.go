// This file implements the ctrl command for GoFrame controller generation.

package main

import (
	"context"
	"fmt"

	"linactl/internal/goframecli"
)

// runCtrl runs the embedded GoFrame gen ctrl command in the selected backend
// directory without requiring an external gf binary.
func runCtrl(ctx context.Context, a *app, input commandInput) error {
	if len(input.Args) > 0 {
		return fmt.Errorf("ctrl accepts target parameters only; use dir=<path> or p=<plugin-id>")
	}
	pluginID := input.Get("p")
	pluginIDSet := input.Has("p")
	if input.Has("plugin") {
		pluginID = input.Get("plugin")
		pluginIDSet = true
	}
	targetDir, err := goframecli.ResolveTargetDir(a.root, goframecli.TargetOptions{
		Dir:         input.Get("dir"),
		DirSet:      input.Has("dir"),
		Target:      input.Get("target"),
		TargetSet:   input.Has("target"),
		PluginID:    pluginID,
		PluginIDSet: pluginIDSet,
	})
	if err != nil {
		return err
	}
	return goframecli.Run(ctx, targetDir, a.executable, a.runCommand, "gen", "ctrl")
}
