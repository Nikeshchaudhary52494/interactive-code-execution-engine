package modules

type ExecuteRequest struct {
	Language    string
	Code        string
	TimeLimitMs int64
	Inputs      []string
}

type ExecuteResult struct {
	ExitCode   int
	Stdout     string
	Stderr     string
	DurationMs int64
	TimedOut   bool
}
