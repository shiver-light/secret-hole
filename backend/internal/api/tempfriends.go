package api

import (
	"context"
	"time"

	"secrethole/backend/internal/db"

	"github.com/jackc/pgx/v5"
)

func normPair(a, b int64) (int64, int64) {
	if a < b {
		return a, b
	}
	return b, a
}

const (
	baseWindow = time.Minute
	// 近似 1000 年（time.Duration 为纳秒 int64，安全）
	capWindow = time.Hour * 24 * 365 * 100
)

// 开启或续期临时好友窗口；保留 streak；累计历史时长
func openOrExtendTempFriendship(ctx context.Context, d *db.DB, a, b int64) (until time.Time, streak int, err error) {
	minID, maxID := normPair(a, b)
	tx, err := d.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return
	}
	defer tx.Rollback(ctx)

	var lastStarted, activeUntil *time.Time
	err = tx.QueryRow(ctx, `
		SELECT streak, last_started_at, active_until
		  FROM temp_friendships
		 WHERE user_min=$1 AND user_max=$2
		 FOR UPDATE
	`, minID, maxID).Scan(&streak, &lastStarted, &activeUntil)
	if err != nil && err != pgx.ErrNoRows {
		return
	}
	if err == pgx.ErrNoRows {
		streak = 1
	} else {
		if streak <= 0 {
			streak = 1
		} else {
			streak++
		}
	}

	dur := baseWindow * time.Duration(1<<(streak-1))
	if dur > capWindow {
		dur = capWindow
	}
	now := time.Now()
	until = now.Add(dur)

	var addSec int64
	if activeUntil != nil && lastStarted != nil {
		end := *activeUntil
		if end.After(now) {
			end = now
		}
		if end.After(*lastStarted) {
			addSec = int64(end.Sub(*lastStarted).Seconds())
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO temp_friendships(user_min,user_max,streak,last_started_at,active_until,total_active_seconds)
		VALUES($1,$2,$3,$4,$5,$6)
		ON CONFLICT(user_min,user_max) DO UPDATE
		  SET streak=$3,
		      last_started_at=$4,
		      active_until=$5,
		      total_active_seconds = temp_friendships.total_active_seconds + $6
	`, minID, maxID, streak, now, until, addSec)
	if err != nil {
		return
	}
	err = tx.Commit(ctx)
	return
}
