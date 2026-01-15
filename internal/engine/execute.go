package engine

import "context"

type Executor interface {
	Run(
		ctx context.Context,
		lang string,
		code string,
		inputs []string,
	) (*ExecuteResult, error)
}