package engine

import (
	"context"
	"log"
	"sync"
	"time"

	"execution-engine/internal/executor"
	"execution-engine/internal/modules"
	"execution-engine/internal/session"
)

type engineImpl struct {
	executor *executor.DockerExecutor
	sessions *session.Manager
	sem      chan struct{} // concurrency limiter
	wg       sync.WaitGroup
}

func New(exec *executor.DockerExecutor) Engine {
	return &engineImpl{
		executor: exec,
		sessions: session.NewManager(),
		sem:      make(chan struct{}, 10), // 🔥 MAX 10 containers
	}
}

func (e *engineImpl) StartSession(
	ctx context.Context,
	req modules.ExecuteRequest,
) (*session.Session, error) {

	// 1️⃣ Create LOGICAL session (WAITING)
	sess := session.NewPending(
		session.NewID(),
		req.Language,
		req.Code,
	)

	e.sessions.Add(sess)

	log.Printf("Engine: session %s created (WAITING)", sess.ID)

	e.wg.Add(1)
	// 2️⃣ Background goroutine tries to run it
	go func() {
		defer e.wg.Done()
		select {
		case e.sem <- struct{}{}:
			// slot acquired
			log.Printf("Engine: slot acquired for session %s", sess.ID)

			// start actual docker execution
			err := e.executor.StartSession(
				context.Background(),
				sess,
			)
			if err != nil {
				log.Printf("Engine: failed to start session %s: %v", sess.ID, err)
				sess.MarkTerminated()
				<-e.sem
				e.sessions.Remove(sess.ID)
				return
			}

			sess.MarkRunning()

			// wait until execution finishes AND resources are cleaned up
			<-sess.CleanupDone()

			log.Printf(
				"Engine: session %s finished (state=%s)",
				sess.ID,
				sess.State,
			)

			e.sessions.Remove(sess.ID)
			<-e.sem // 🔥 release slot

		case <-time.After(2 * time.Minute):
			// optional: waiting timeout
			log.Printf("Engine: session %s timed out while waiting", sess.ID)
			sess.MarkTerminated()
			e.sessions.Remove(sess.ID)
		}
	}()

	return sess, nil
}

func (e *engineImpl) GetSession(id string) (*session.Session, bool) {
	return e.sessions.Get(id)
}

func (e *engineImpl) Shutdown(ctx context.Context) error {
	log.Println("Engine: shutting down, waiting for active sessions...")
	
	done := make(chan struct{})
	go func() {
		e.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Engine: all sessions finished.")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
