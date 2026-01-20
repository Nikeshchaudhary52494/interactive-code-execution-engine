package engine

import (
	"context"

	"execution-engine/internal/modules"
	"execution-engine/internal/session"
)

type Engine interface {
	StartSession(ctx context.Context, req modules.ExecuteRequest) (*session.Session, error)
	GetSession(id string) (*session.Session, bool)
}
