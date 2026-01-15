package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"execution-engine/internal/engine"
	"execution-engine/internal/executor"
	"execution-engine/internal/modules"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // handle auth later
	},
}

func main() {
	// ---- bootstrap docker executor ----
	dockerExec, err := executor.NewDockerExecutor()
	if err != nil {
		panic(err)
	}

	// ---- engine ----
	eng := engine.New(dockerExec)

	r := gin.Default()

	// ------------------------------------------------
	// Create Session (HTTP)
	// ------------------------------------------------
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

	// ------------------------------------------------
	// WebSocket: Interactive Execution
	// ------------------------------------------------
	r.GET("/ws/session/:id", func(c *gin.Context) {
		id := c.Param("id")
		fmt.Println("ws connection for session:", id)

		sess, ok := eng.GetSession(id)
		if !ok {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "session not found",
			})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// -------------------------------
		// Client → stdin
		// -------------------------------
		go func() {
			for {
				var msg struct {
					Type string `json:"type"`
					Data string `json:"data"`
				}

				if err := conn.ReadJSON(&msg); err != nil {
					// client disconnected
					sess.CloseInput()
					return
				}

				switch msg.Type {
				case "input":
					_ = sess.WriteInput(msg.Data)
				case "close":
					sess.CloseInput()
				}
			}
		}()

		// -------------------------------
		// stdout / stderr → client
		// -------------------------------
		lastStdout := 0
		lastStderr := 0

		for {
			select {
			case <-sess.Done():
				// flush remaining output
				_ = sendDiff(conn, "stdout", sess.Stdout.String(), &lastStdout)
				_ = sendDiff(conn, "stderr", sess.Stderr.String(), &lastStderr)

				_ = conn.WriteJSON(gin.H{
					"type":  "state",
					"state": sess.State,
				})
				return

			default:
				if err := sendDiff(conn, "stdout", sess.Stdout.String(), &lastStdout); err != nil {
					return
				}
				if err := sendDiff(conn, "stderr", sess.Stderr.String(), &lastStderr); err != nil {
					return
				}
				time.Sleep(50 * time.Millisecond)
			}
		}
	})

	// ------------------------------------------------
	// Start server
	// ------------------------------------------------
	if err := r.Run(":8080"); err != nil {
		panic(err)
	}
}

func sendDiff(conn *websocket.Conn, t string, data string, last *int) error {
	if len(data) > *last {
		chunk := data[*last:]
		*last = len(data)

		return conn.WriteJSON(gin.H{
			"type": t,
			"data": chunk,
		})
	}
	return nil
}
