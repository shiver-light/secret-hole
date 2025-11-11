package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"secrethole/backend/internal/db"
	"secrethole/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type anonReq struct {
	DeviceHash string `json:"device_hash"`
}

type anonResp struct {
	UserID   int64  `json:"user_id"`
	Date     string `json:"date"`
	SentLeft int    `json:"send_left"`
	RecvLeft int    `json:"claim_left"`
}

func AuthAnon(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req anonReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		req.DeviceHash = strings.TrimSpace(req.DeviceHash)
		if req.DeviceHash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "device_hash required"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()
		var userID int64
		// upsert user by device_hash
		err := d.Pool.QueryRow(ctx, `
INSERT INTO users(device_hash) VALUES ($1)
ON CONFLICT(device_hash) DO UPDATE SET device_hash = EXCLUDED.device_hash
RETURNING id;
`, req.DeviceHash).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		date := util.ShanghaiDate(time.Now())
		// ensure quota row exists
		_, _ = d.Pool.Exec(ctx, `
INSERT INTO daily_quota(user_id, quota_date) VALUES ($1, $2)
ON CONFLICT (user_id, quota_date) DO NOTHING;
`, userID, date)

		// read counts
		var sent, recv int
		err = d.Pool.QueryRow(ctx, `SELECT sent_count, recv_count FROM daily_quota WHERE user_id=$1 AND quota_date=$2`, userID, date).Scan(&sent, &recv)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, anonResp{UserID: userID, Date: date, SentLeft: 3 - sent, RecvLeft: 3 - recv})
	}
}
