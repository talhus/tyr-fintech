package handlers

import (
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/iamtbay/tyr-fintech/internal/notifications"
	"github.com/iamtbay/tyr-fintech/pkg/apperrors"
)

type NotificationHandler struct {
	hub *notifications.Hub
}

// NEW HANDLER
func NewNotificationHandler(hub *notifications.Hub) *NotificationHandler {
	return &NotificationHandler{
		hub: hub,
	}
}

// STREAM
func (h *NotificationHandler) Stream(c *gin.Context) {
	userID := c.GetString("userID")
	if userID == "" {
		apperrors.New(http.StatusUnauthorized, "Unauthorized")
		return
	}

	origin := c.GetHeader("Origin")
	if origin == "" {
		origin = "http://localhost:3000"
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", origin)
	c.Header("Access-Control-Allow-Credentials", "true")
	c.Header("X-Accel-Buffering", "no")

	messageChan := h.hub.Register(userID)
	defer h.hub.Unregister(userID)

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	//start stream
	c.Stream(func(w io.Writer) bool {
		select {
		case msg, ok := <-messageChan:
			if !ok {
				return false
			}
			c.SSEvent("message", msg)
			return true
		case <-ticker.C:
			c.Writer.WriteString(": heartbeat\n\n")
			c.Writer.Flush()
			return true
		case <-c.Request.Context().Done():
			return false
		}
	})

}

