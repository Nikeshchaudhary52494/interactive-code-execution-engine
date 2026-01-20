package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

// copyCodeToContainer copies source code into /workspace inside the container.
// Works with tmpfs-backed workspaces (REQUIRED for sandboxing).
func copyCodeToContainer(
	ctx context.Context,
	cli *client.Client,
	containerID string,
	filename string,
	code string,
) error {
	if filename == "" {
		return fmt.Errorf("filename cannot be empty")
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	header := &tar.Header{
		Name:    filename,
		Mode:    0644,
		Size:    int64(len(code)),
		ModTime: time.Now(),
	}

	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("tar header error: %w", err)
	}

	if _, err := tw.Write([]byte(code)); err != nil {
		return fmt.Errorf("tar write error: %w", err)
	}

	if err := tw.Close(); err != nil {
		return fmt.Errorf("tar close error: %w", err)
	}

	if err := cli.CopyToContainer(
		ctx,
		containerID,
		"/workspace",
		&buf,
		container.CopyToContainerOptions{
			AllowOverwriteDirWithFile: true,
		},
	); err != nil {
		return fmt.Errorf("docker copy failed: %w", err)
	}

	return nil
}
