package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"secrethole/backend/internal/db"
	"secrethole/backend/internal/util"

	"github.com/gin-gonic/gin"
)

func MeQuota(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		id, _ := strconv.ParseInt(userIDStr, 10, 64)
		if id == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		var sent, recv int
		date := util.ShanghaiDate(time.Now())
		_, _ = d.Pool.Exec(ctx, `INSERT INTO daily_quota(user_id, quota_date) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, date)
		_ = d.Pool.QueryRow(ctx, `SELECT sent_count, recv_count FROM daily_quota WHERE user_id=$1 AND quota_date=$2`, id, date).Scan(&sent, &recv)
		c.JSON(http.StatusOK, gin.H{"date": date, "send_left": 3 - sent, "claim_left": 3 - recv})
	}
}
