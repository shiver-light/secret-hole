package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"secrethole/backend/internal/db"
	"secrethole/backend/internal/util"

	"github.com/gin-gonic/gin"
)

type postReq struct {
	Body         string `json:"body"`
	ReadDuration int    `json:"read_duration"`
}

type postResp struct {
	MessageID int64 `json:"message_id"`
}

type claimResp struct {
	ID           int64     `json:"message_id"`
	Body         string    `json:"body"`
	ReadDuration int       `json:"read_duration"`
	ReadAt       time.Time `json:"read_at"`
	ExpiresAt    time.Time `json:"expires_at"`
}

func PostMessage(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := strconv.ParseInt(c.GetHeader("X-User-ID"), 10, 64)
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID required"})
			return
		}
		var req postReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		if req.ReadDuration < 1 || req.ReadDuration > 120 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "read_duration 1~120"})
			return
		}
		if l := len([]rune(req.Body)); l == 0 || l > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "body 1~500 chars"})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		// check & incr quota atomically
		date := util.ShanghaiDate(time.Now())
		cmd := `UPDATE daily_quota SET sent_count = sent_count + 1 WHERE user_id=$1 AND quota_date=$2 AND sent_count < 3`
		ct, err := d.Pool.Exec(ctx, cmd, uid, date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if ct.RowsAffected() == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "daily send limit reached"})
			return
		}

		var id int64
		err = d.Pool.QueryRow(ctx, `INSERT INTO messages(author_id, body_cipher, read_duration) VALUES ($1,$2,$3) RETURNING id`, uid, req.Body, req.ReadDuration).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, postResp{MessageID: id})
	}
}

func ClaimMessage(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, _ := strconv.ParseInt(c.GetHeader("X-User-ID"), 10, 64)
		if uid == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "X-User-ID required"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 4*time.Second)
		defer cancel()

		date := util.ShanghaiDate(time.Now())
		// 先扣领取配额（失败则返回429）
		cmd := `UPDATE daily_quota SET recv_count = recv_count + 1 WHERE user_id=$1 AND quota_date=$2 AND recv_count < 3`
		if ct, err := d.Pool.Exec(ctx, cmd, uid, date); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		} else if ct.RowsAffected() == 0 {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "daily receive limit reached"})
			return
		}

		// 开启事务：挑一条未读且非自己发的，标记为已读并计算过期
		tx, err := d.Pool.Begin(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer tx.Rollback(ctx)

		row := tx.QueryRow(ctx, `
WITH pick AS (
SELECT id FROM messages WHERE status = 0 AND (author_id IS NULL OR author_id <> $1)
ORDER BY random() LIMIT 1 FOR UPDATE SKIP LOCKED
)
UPDATE messages m
SET status=1, read_at=now(), expires_at = now() + (m.read_duration || ' seconds')::interval
FROM pick
WHERE m.id = pick.id
RETURNING m.id, m.body_cipher, m.read_duration, m.read_at, m.expires_at;
`, uid)

		var resp claimResp
		if err := row.Scan(&resp.ID, &resp.Body, &resp.ReadDuration, &resp.ReadAt, &resp.ExpiresAt); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.JSON(http.StatusNotFound, gin.H{"error": "no message available"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := tx.Commit(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

func AckDelete() gin.HandlerFunc {
	return func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) }
}
