package engine

import (
	"context"
	"execution-engine/internal/modules"
)

type Executor interface {
	Run(
		ctx context.Context,
		lang string,
		code string,
		inputs []string,
	) (*modules.ExecuteResult, error)
}