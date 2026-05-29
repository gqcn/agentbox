// This file contains target-directory validation for embedded GoFrame code generation.

package goframecli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TargetOptions describes one requested GoFrame code generation target.
type TargetOptions struct {
	Dir           string
	DirSet        bool
	Target        string
	TargetSet     bool
	PluginID      string
	PluginIDSet   bool
	RequireConfig bool
}

// ResolveTargetDir resolves the backend directory whose hack/config.yaml should
// drive GoFrame code generation.
func ResolveTargetDir(root string, options TargetOptions) (string, error) {
	if options.DirSet && strings.TrimSpace(options.Dir) == "" {
		return "", fmt.Errorf("GoFrame code generation parameter dir is empty")
	}
	if options.TargetSet && strings.TrimSpace(options.Target) == "" {
		return "", fmt.Errorf("GoFrame code generation parameter target is empty")
	}
	if options.PluginIDSet && strings.TrimSpace(options.PluginID) == "" {
		return "", fmt.Errorf("GoFrame code generation parameter p is empty")
	}

	var selected []string
	if options.DirSet {
		selected = append(selected, "dir")
	}
	if options.TargetSet {
		selected = append(selected, "target")
	}
	if options.PluginIDSet {
		selected = append(selected, "p")
	}
	if len(selected) > 1 {
		return "", fmt.Errorf("GoFrame code generation accepts one target selector, got %s", strings.Join(selected, ", "))
	}

	target := filepath.Join("apps", "lina-core")
	switch {
	case options.DirSet:
		target = options.Dir
	case options.TargetSet:
		target = options.Target
	case options.PluginIDSet:
		target = filepath.Join("apps", "lina-plugins", options.PluginID, "backend")
	}

	targetDir, err := normalizeTargetDir(root, target)
	if err != nil {
		return "", err
	}
	if err := ValidateProjectDir(targetDir); err != nil {
		return "", err
	}
	if options.RequireConfig {
		if err := ValidateTargetConfig(targetDir); err != nil {
			return "", err
		}
	}
	return targetDir, nil
}

// ValidateProjectDir checks that a code generation target is a directory.
func ValidateProjectDir(targetDir string) error {
	if targetDir == "" {
		return fmt.Errorf("GoFrame code generation target directory is empty")
	}
	info, err := os.Stat(targetDir)
	if err != nil {
		return fmt.Errorf("check GoFrame code generation target %s: %w", targetDir, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("GoFrame code generation target %s is not a directory", targetDir)
	}
	return nil
}

// ValidateTargetConfig checks that a target directory has GoFrame CLI config.
func ValidateTargetConfig(targetDir string) error {
	if err := ValidateProjectDir(targetDir); err != nil {
		return err
	}
	configPath := filepath.Join(targetDir, "hack", "config.yaml")
	configInfo, err := os.Stat(configPath)
	if err != nil {
		return fmt.Errorf("GoFrame code generation target %s is missing hack/config.yaml: %w", targetDir, err)
	}
	if configInfo.IsDir() {
		return fmt.Errorf("GoFrame code generation target %s has hack/config.yaml as a directory", targetDir)
	}
	return nil
}

func normalizeTargetDir(root string, target string) (string, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return "", fmt.Errorf("GoFrame code generation target directory is empty")
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(root, filepath.FromSlash(target))
	}
	absolute, err := filepath.Abs(target)
	if err != nil {
		return "", fmt.Errorf("resolve GoFrame code generation target %s: %w", target, err)
	}
	return absolute, nil
}
