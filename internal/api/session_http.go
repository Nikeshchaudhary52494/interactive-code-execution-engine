package api

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"execution-engine/internal/engine"
	"execution-engine/internal/modules"
)

func RegisterSessionHTTP(r *gin.Engine, eng engine.Engine) {
	r.POST("/session", func(c *gin.Context) {
		var req modules.ExecuteRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
			return
		}

		sess, err := eng.StartSession(c.Request.Context(), req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Session %s created", sess.ID)

		c.JSON(http.StatusOK, gin.H{
			"sessionId": sess.ID,
		})
	})
}
