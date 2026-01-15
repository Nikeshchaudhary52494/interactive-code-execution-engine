package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"execution-engine/internal/engine"
	"execution-engine/internal/executor"
	"execution-engine/internal/modules"
)

func main() {
	// ---- bootstrap docker executor ----
	dockerExec, err := executor.NewDockerExecutor()
	if err != nil {
		panic(err)
	}

	// ---- engine ----
	eng := engine.New(dockerExec)

	r := gin.Default()

	// -------------------------------
	// Create Session
	// -------------------------------
	r.POST("/session", func(c *gin.Context) {
		var req modules.ExecuteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid json",
			})
			return
		}

		sess, err := eng.StartSession(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"sessionId": sess.ID,
		})
	})

	// -------------------------------
	// Send Input
	// -------------------------------
	r.POST("/session/:id/input", func(c *gin.Context) {
		id := c.Param("id")

		sess, ok := eng.GetSession(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "session not found",
			})
			return
		}

		var body struct {
			Data string `json:"data"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid json",
			})
			return
		}

		if err := sess.WriteInput(body.Data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "input written",
			"state":   sess.State,
		})
	})

	// -------------------------------
	// Get Output (temporary polling)
	// -------------------------------
	r.GET("/session/:id/output", func(c *gin.Context) {
		id := c.Param("id")

		sess, ok := eng.GetSession(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "session not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"stdout": sess.Stdout.String(),
			"stderr": sess.Stderr.String(),
			"state":  sess.State,
		})
	})

	// -------------------------------
	// Close Input
	// -------------------------------
	r.POST("/session/:id/close", func(c *gin.Context) {
		id := c.Param("id")

		sess, ok := eng.GetSession(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "session not found",
			})
			return
		}

		// Close stdin (send EOF)
		sess.CloseInput()

		// OPTIONAL: wait until process actually exits
		// sess.Wait()

		c.JSON(http.StatusOK, gin.H{
			"message": "session closed",
			"state":   sess.State,
			"stdout":  sess.Stdout.String(),
			"stderr":  sess.Stderr.String(),
		})
	})

	// -------------------------------
	// Start server
	// -------------------------------
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}
