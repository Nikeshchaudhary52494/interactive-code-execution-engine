package session

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type Session struct {
	ID        string
	State     State
	StartedAt time.Time

	ContainerID string

	Stdin  io.WriteCloser
	Output io.Reader

	Stdout strings.Builder
	Stderr strings.Builder

	done     chan struct{}
	doneOnce sync.Once
	mu       sync.Mutex
}

func New(id, containerID string, stdin io.WriteCloser, output io.Reader) *Session {
	return &Session{
		ID:          id,
		ContainerID: containerID,
		State:       StateRunning,
		StartedAt:   time.Now(),
		Stdin:       stdin,
		Output:      output,
		done:        make(chan struct{}),
	}
}

// --------------------
// Input handling
// --------------------

func (s *Session) WriteInput(data string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State != StateRunning && s.State != StateWaitingInput {
		return fmt.Errorf("session not accepting input (state=%s)", s.State)
	}

	_, err := s.Stdin.Write([]byte(data))
	return err
}

func (s *Session) CloseInput() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Mark intent: no more input
	if s.State == StateRunning {
		s.State = StateWaitingInput
	}

	_ = s.Stdin.Close()
}

// --------------------
// Lifecycle handling
// --------------------

func (s *Session) MarkFinished() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == StateFinished || s.State == StateClosed {
		return
	}

	s.State = StateFinished
	s.signalDone()
}

func (s *Session) MarkTerminated() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == StateFinished || s.State == StateClosed {
		return
	}

	s.State = StateTerminated
	s.signalDone()
}

func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.State == StateClosed {
		return
	}

	s.State = StateClosed
	s.signalDone()
}

// --------------------
// Synchronization
// --------------------

func (s *Session) signalDone() {
	s.doneOnce.Do(func() {
		close(s.done)
	})
}

func (s *Session) Done() <-chan struct{} {
	return s.done
}

// Wait blocks until the session finishes or terminates
func (s *Session) Wait() {
	<-s.done
}
