package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"secrethole/backend/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type OkResp struct {
	Ok bool `json:"ok"`
}

type MeProfileResp struct {
	UserID       int64   `json:"user_id"`
	Name         string  `json:"name"`
	AvatarBase64 *string `json:"avatar_base64,omitempty"`
	GenderColor  *string `json:"gender_color,omitempty"`
}

type MeProfileUpdateReq struct {
	Name         string  `json:"name"`
	AvatarBase64 *string `json:"avatar_base64,omitempty"`
	GenderColor  *string `json:"gender_color,omitempty"`
}

// GET /v1/me/profile
func MeProfileGet(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var name string
		var avatar sql.NullString
		var color sql.NullString

		err := d.Pool.QueryRow(
			c.Request.Context(),
			`SELECT name, avatar_base64, gender_color FROM users WHERE id=$1`,
			uid,
		).Scan(&name, &avatar, &color)
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error", "detail": err.Error()})
			return
		}

		resp := MeProfileResp{
			UserID: uid,
			Name:   name,
		}
		if avatar.Valid {
			resp.AvatarBase64 = &avatar.String
		}
		if color.Valid {
			resp.GenderColor = &color.String
		}
		c.JSON(http.StatusOK, resp)
	}
}

// PUT /v1/me/profile
func MeProfilePut(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		var req MeProfileUpdateReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" || runeLen(req.Name) > 7 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid name (1~7 chars)"})
			return
		}
		if req.GenderColor != nil && *req.GenderColor != "" && !isHexColor(*req.GenderColor) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gender_color must be #RRGGBB"})
			return
		}

		// Null 处理
		nAvatar := toNullString(req.AvatarBase64)
		nColor := toNullString(req.GenderColor)

		ct := c.Request.Context()
		cmd, err := d.Pool.Exec(ct,
			`UPDATE users
			  SET name=$1, avatar_base64=$2, gender_color=$3
			  WHERE id=$4`,
			req.Name, nAvatar, nColor, uid,
		)
		if err != nil {
			// 唯一冲突（昵称重复）
			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "name exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error", "detail": err.Error()})
			return
		}
		if cmd.RowsAffected() == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusOK, OkResp{Ok: true})
	}
}

// POST /v1/auth/logout
// 使当前 token 失效：删除或立即过期
func LogoutHandler(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, ok := currentUserID(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		tok := c.GetHeader("X-Auth-Token")
		if tok == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing token"})
			return
		}

		ct := c.Request.Context()
		// 方案1：物理删除
		_, err := d.Pool.Exec(ct,
			`DELETE FROM auth_tokens WHERE token=$1 AND user_id=$2`,
			tok, uid,
		)
		// 方案2（可选）：逻辑过期
		// _, err := d.Pool.Exec(ct, `UPDATE auth_tokens SET expires_at=$1 WHERE token=$2 AND user_id=$3`, time.Now().Add(-time.Hour), tok, uid)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error", "detail": err.Error()})
			return
		}
		c.JSON(http.StatusOK, OkResp{Ok: true})
	}
}

// ========== 辅助 ==========
func currentUserID(c *gin.Context) (int64, bool) {
	// 优先中间件注入的 user_id
	if v, ok := c.Get("user_id"); ok {
		if id, ok2 := v.(int64); ok2 {
			return id, true
		}
	}
	// 兼容旧：X-User-ID（不安全，仅开发/过渡期）
	if s := strings.TrimSpace(c.GetHeader("X-User-ID")); s != "" {
		if id, err := parseInt64(s); err == nil {
			return id, true
		}
	}
	return 0, false
}
func parseInt64(s string) (int64, error) {
	var id int64
	_, err := fmt.Sscan(s, &id)
	return id, err
}
func runeLen(s string) int { return len([]rune(s)) }

var hexRe = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

func isHexColor(s string) bool { return hexRe.MatchString(s) }
func toNullString(p *string) sql.NullString {
	if p == nil || strings.TrimSpace(*p) == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: strings.TrimSpace(*p), Valid: true}
}
