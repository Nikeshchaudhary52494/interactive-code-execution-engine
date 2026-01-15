package engine

import (
	"context"
	"log"

	"execution-engine/internal/executor"
	"execution-engine/internal/modules"
	"execution-engine/internal/session"
)

type Engine interface {
	StartSession(ctx context.Context, req modules.ExecuteRequest) (*session.Session, error)
	GetSession(id string) (*session.Session, bool)
}

type engineImpl struct {
	executor *executor.DockerExecutor
	sessions *session.Manager
}

func New(exec *executor.DockerExecutor) Engine {
	return &engineImpl{
		executor: exec,
		sessions: session.NewManager(),
	}
}

func (e *engineImpl) StartSession(
	ctx context.Context,
	req modules.ExecuteRequest,
) (*session.Session, error) {

	sess, err := e.executor.StartSession(ctx, req.Language, req.Code)
	if err != nil {
		return nil, err
	}

	e.sessions.Add(sess)
	sess.Start() // Start session lifecycle (e.g., inactivity timer)

	// 🔥 AUTO-REMOVE when done
	go func() {
		<-sess.Done()
		log.Printf("Engine: Session %s is done, removing from manager", sess.ID)
		e.sessions.Remove(sess.ID)
	}()

	return sess, nil
}

func (e *engineImpl) GetSession(id string) (*session.Session, bool) {
	return e.sessions.Get(id)
}
