package executor

import (
	"context"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/pkg/stdcopy"

	"execution-engine/internal/session"
)

func (d *DockerExecutor) watchSession(
	ctx context.Context,
	s *session.Session,
	tempDir string,
) {
	defer os.RemoveAll(tempDir)

	// pump stdout/stderr until container exits
	stdcopy.StdCopy(&s.Stdout, &s.Stderr, s.Output)

	statusCh, _ := d.cli.ContainerWait(
		ctx,
		s.ContainerID,
		container.WaitConditionNotRunning,
	)

	<-statusCh

	s.MarkFinished()

	_ = d.cli.ContainerRemove(
		context.Background(),
		s.ContainerID,
		container.RemoveOptions{Force: true},
	)
}
