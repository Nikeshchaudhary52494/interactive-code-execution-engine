package main

import (
	"log"

	"execution-engine/internal/api"
	"execution-engine/internal/engine"
	"execution-engine/internal/executor"
)

func main() {
	// ---- bootstrap docker executor ----
	dockerExec, err := executor.NewDockerExecutor()
	if err != nil {
		panic(err)
	}

	// ---- engine ----
	eng := engine.New(dockerExec)

	// ---- router ----
	r := api.New(eng)

	// ------------------------------------------------
	// Start server
	// ------------------------------------------------
	log.Println("Server started on :8080")
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}