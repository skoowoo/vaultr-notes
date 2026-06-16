package agent

import (
	"context"
	"os/exec"
	"time"
)

func fetchDevinModels(resolved string, env []string) ([]ModelOption, error) {
	return DetectACPModels(context.Background(), ACPDetectOpts{
		Bin: resolved,
		Args: []string{
			"--permission-mode", "dangerous",
			"--respect-workspace-trust", "false",
			"acp",
		},
		Env:           env,
		Timeout:       15 * time.Second,
		DefaultOption: DefaultModelOption,
	})
}

func fetchHermesModels(resolved string, env []string) ([]ModelOption, error) {
	return DetectACPModels(context.Background(), ACPDetectOpts{
		Bin:           resolved,
		Args:          []string{"acp", "--accept-hooks"},
		Env:           env,
		Timeout:       15 * time.Second,
		DefaultOption: DefaultModelOption,
	})
}

func fetchKimiModels(resolved string, env []string) ([]ModelOption, error) {
	return DetectACPModels(context.Background(), ACPDetectOpts{
		Bin:           resolved,
		Args:          []string{"acp"},
		Env:           env,
		Timeout:       15 * time.Second,
		DefaultOption: DefaultModelOption,
	})
}

func fetchKiroModels(resolved string, env []string) ([]ModelOption, error) {
	return DetectACPModels(context.Background(), ACPDetectOpts{
		Bin:           resolved,
		Args:          []string{"acp"},
		Env:           env,
		Timeout:       15 * time.Second,
		DefaultOption: DefaultModelOption,
	})
}

func fetchKiloModels(resolved string, env []string) ([]ModelOption, error) {
	return fetchKiroModels(resolved, env)
}

func fetchVibeModels(resolved string, env []string) ([]ModelOption, error) {
	return DetectACPModels(context.Background(), ACPDetectOpts{
		Bin:           resolved,
		Args:          nil,
		Env:           env,
		Timeout:       15 * time.Second,
		DefaultOption: DefaultModelOption,
	})
}

func fetchPiModels(resolved string, env []string) ([]ModelOption, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, resolved, "--list-models")
	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	m := ParsePiModels(string(out))
	if len(m) == 0 {
		return nil, nil
	}
	return m, nil
}
