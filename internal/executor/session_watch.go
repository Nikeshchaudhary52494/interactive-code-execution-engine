package executor

import (
	"context"
	"os"

	"github.com/docker/docker/api/types/container"

	"execution-engine/internal/session"
)

func (d *DockerExecutor) watchSession(
	s *session.Session,
	tempDir string,
) {
	defer os.RemoveAll(tempDir)

	// ---------------- stream stdout ----------------
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := s.Output.Read(buf)
			if n > 0 {
				s.AppendOutput(buf[:n]) // 🔥 safe wrapper
			}
			if err != nil {
				return
			}
		}
	}()

	waitCh, _ := d.cli.ContainerWait(
		context.Background(),
		s.ContainerID,
		container.WaitConditionNotRunning,
	)

	select {
	case <-waitCh:
		s.MarkFinished()

	case <-s.Context().Done(): // 🔥 session cancelled
		_ = d.cli.ContainerKill(
			context.Background(),
			s.ContainerID,
			"KILL",
		)
		s.MarkTerminated()
	}

	// ALWAYS remove container
	_ = d.cli.ContainerRemove(
		context.Background(),
		s.ContainerID,
		container.RemoveOptions{Force: true},
	)
}
