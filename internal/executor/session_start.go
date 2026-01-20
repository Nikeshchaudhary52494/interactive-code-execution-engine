package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"

	"execution-engine/internal/language"
	"execution-engine/internal/session"
)

const (
	workspaceDir = "/workspace"
)

func (d *DockerExecutor) StartSession(
	ctx context.Context,
	lang string,
	code string,
) (*session.Session, error) {

	spec, err := language.Resolve(lang)
	if err != nil {
		return nil, err
	}

	tempDir, err := os.MkdirTemp("", "exec-*")
	if err != nil {
		return nil, err
	}

	codePath := filepath.Join(tempDir, spec.FileName)
	if err := os.WriteFile(codePath, []byte(code), 0644); err != nil {
		return nil, err
	}

	createResp, err := d.cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:           spec.Image,
			Cmd:             spec.RunCommand,
			WorkingDir:      workspaceDir,
			OpenStdin:       true,
			AttachStdin:     true,
			StdinOnce:       false,
			AttachStdout:    true,
			AttachStderr:    true,
			NetworkDisabled: true,
		},
		&container.HostConfig{
			Resources: container.Resources{
				Memory:    200 * 1024 * 1024,
				NanoCPUs:  500_000_000,
				PidsLimit: ptr(int64(32)),
			},
			ReadonlyRootfs: true,
			CapDrop:        []string{"ALL"},
			SecurityOpt:    []string{"no-new-privileges"},
			Tmpfs: map[string]string{
				// "/workspace": "rw,size=64m,noexec,nosuid",
				"/tmp": "rw,size=32m,noexec,nosuid",
			},
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   tempDir,
					Target:   workspaceDir,
					ReadOnly: false,
				},
			},
		},
		nil, nil, "",
	)
	if err != nil {
		return nil, fmt.Errorf("container create: %w", err)
	}

	// containerID := createResp.ID
	// defer func() {
	// 	_ = d.cli.ContainerRemove(
	// 		context.Background(),
	// 		containerID,
	// 		container.RemoveOptions{Force: true},
	// 	)
	// }()

	// if err := copyCodeToContainer(
	// 	ctx,
	// 	d.cli,
	// 	createResp.ID,
	// 	spec.FileName, // e.g. main.py
	// 	code,
	// ); err != nil {
	// 	// cleanup on failure
	// 	_ = d.cli.ContainerRemove(context.Background(), createResp.ID, container.RemoveOptions{Force: true})
	// 	return nil, err
	// }

	attach, err := d.cli.ContainerAttach(
		ctx,
		createResp.ID,
		container.AttachOptions{
			Stream: true,
			Stdin:  true,
			Stdout: true,
			Stderr: true,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("container attach: %w", err)
	}

	if err := d.cli.ContainerStart(ctx, createResp.ID, container.StartOptions{}); err != nil {
		return nil, err
	}

	sessCtx, cancel := context.WithCancel(context.Background())

	sess := session.New(
		session.NewID(),
		createResp.ID,
		attach.Conn,
		attach.Reader,
		sessCtx,
		cancel,
	)

	go d.watchSession(sess, tempDir)

	return sess, nil
}
