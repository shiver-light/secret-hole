package util

import (
	"context"
	"errors"
	"time"

	"secrethole/backend/internal/db"

	"github.com/jackc/pgx/v5"
)

// 定义“同时在线”= is_foreground 且 20 秒内有心跳
func IsActiveForeground(ctx context.Context, d *db.DB, uid int64) (bool, error) {
	var isFg bool
	var ts time.Time
	err := d.Pool.QueryRow(ctx, `
		SELECT is_foreground, last_heartbeat
		  FROM user_presence WHERE user_id=$1
	`, uid).Scan(&isFg, &ts)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	if !isFg {
		return false, nil
	}
	return time.Since(ts) <= 20*time.Second, nil
}
