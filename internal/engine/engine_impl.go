package engine

import (
	"context"
	"execution-engine/internal/executor"
	"execution-engine/internal/modules"
	"execution-engine/internal/session"
	"log"
)

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

	// 1️⃣ Start execution (container + session)
	sess, err := e.executor.StartSession(ctx, req.Language, req.Code)
	if err != nil {
		return nil, err
	}

	// 2️⃣ Register session
	e.sessions.Add(sess)

	log.Printf("Engine: session %s started", sess.ID)

	// 3️⃣ Auto-cleanup when session finishes or is terminated
	go func() {
		<-sess.Done()

		log.Printf(
			"Engine: session %s ended (state=%s), cleaning up",
			sess.ID,
			sess.State,
		)

		e.sessions.Remove(sess.ID)
	}()

	return sess, nil
}

func (e *engineImpl) GetSession(id string) (*session.Session, bool) {
	return e.sessions.Get(id)
}
