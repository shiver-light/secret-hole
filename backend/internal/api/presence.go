package api

import (
	"net/http"
	"strconv"

	"secrethole/backend/internal/db"

	"github.com/gin-gonic/gin"
)

type hbReq struct {
	IsForeground bool `json:"is_foreground"`
}

// POST /v1/presence/heartbeat
func PresenceHeartbeat(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := strconv.ParseInt(c.GetHeader("X-User-ID"), 10, 64)
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID required"})
			return
		}
		var req hbReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		_, err := d.Pool.Exec(c, `
			INSERT INTO user_presence(user_id, is_foreground, last_heartbeat)
			VALUES ($1,$2,now())
			ON CONFLICT (user_id) DO UPDATE
			  SET is_foreground=EXCLUDED.is_foreground,
			      last_heartbeat=EXCLUDED.last_heartbeat
		`, uid, req.IsForeground)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}
