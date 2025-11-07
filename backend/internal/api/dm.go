package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"secrethole/backend/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

// POST /v1/dm/send  { to, body }  —— 有效临时好友窗口内不限条数
func SendDM(d *db.DB) gin.HandlerFunc {
	type req struct {
		To   int64  `json:"to"`
		Body string `json:"body"`
	}
	return func(c *gin.Context) {
		uid, _ := strconv.ParseInt(c.GetHeader("X-User-ID"), 10, 64)
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID required"})
			return
		}
		var r req
		if err := c.BindJSON(&r); err != nil || strings.TrimSpace(r.Body) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		minID, maxID := normPair(uid, r.To)
		var until *string
		err := d.Pool.QueryRow(c, `
			SELECT to_char(active_until, 'YYYY-MM-DD"T"HH24:MI:SSOF')
			  FROM temp_friendships
			 WHERE user_min=$1 AND user_max=$2
			   AND active_until IS NOT NULL
			   AND active_until > now()
		`, minID, maxID).Scan(&until)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				c.JSON(http.StatusForbidden, gin.H{"error": "no active window"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if _, err := d.Pool.Exec(c, `
			INSERT INTO direct_messages(sender_id, recv_id, body)
			VALUES($1,$2,$3)
		`, uid, r.To, r.Body); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true, "active_until": until})
	}
}
