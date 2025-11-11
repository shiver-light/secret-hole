package api

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"secrethole/backend/internal/db"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v5" // 用于 ErrNoRows
	"golang.org/x/crypto/argon2"
)

// ====== 请求/响应模型 ======

type RegisterReq struct {
	Name         string `json:"name"`
	Password     string `json:"password"`
	AvatarBase64 string `json:"avatar_base64,omitempty"`
	GenderColor  string `json:"gender_color,omitempty"`
}
type RegisterResp struct {
	UserID int64 `json:"user_id"`
}

type LoginReq struct {
	UserID   int64  `json:"user_id"`
	Password string `json:"password"`
}
type LoginResp struct {
	UserID int64  `json:"user_id"`
	Token  string `json:"token"`
}

// ====== Handler（使用 d.Pool，别再从 Context 取 DB）======

func RegisterHandler(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		req.Name = strings.TrimSpace(req.Name)
		if req.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
			return
		}
		if len([]rune(req.Name)) > 7 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name too long (<=7 characters)"})
			return
		}
		if len(req.Password) < 6 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "password too short (>=6)"})
			return
		}
		if req.GenderColor != "" && !strings.HasPrefix(req.GenderColor, "#") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "gender_color must be like #RRGGBB"})
			return
		}

		hash, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "hash error"})
			return
		}

		ctx := c.Request.Context()
		var userID int64
		err = d.Pool.QueryRow(
			ctx,
			`INSERT INTO users(name, avatar_base64, gender_color, password_hash)
			 VALUES ($1,$2,$3,$4) RETURNING id`,
			req.Name, nullIfEmpty(req.AvatarBase64), nullIfEmpty(req.GenderColor), hash,
		).Scan(&userID)
		if err != nil {
			if isUniqueViolation(err) {
				c.JSON(http.StatusConflict, gin.H{"error": "name exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		c.JSON(http.StatusOK, RegisterResp{UserID: userID})
	}
}

func LoginHandler(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}

		ctx := c.Request.Context()

		var hash string
		err := d.Pool.QueryRow(ctx, `SELECT password_hash FROM users WHERE id=$1`, req.UserID).Scan(&hash)
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id or password"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
			return
		}
		if !verifyPassword(hash, req.Password) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user_id or password"})
			return
		}

		token := randomToken(32)
		_, err = d.Pool.Exec(
			ctx,
			`INSERT INTO auth_tokens(token, user_id, expires_at) VALUES ($1,$2,$3)`,
			token, req.UserID, time.Now().Add(30*24*time.Hour),
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token issue"})
			return
		}
		c.JSON(http.StatusOK, LoginResp{UserID: req.UserID, Token: token})
	}
}

// 可选：基于 Token 的鉴权中间件（统一用 d.Pool）
func AuthMiddleware(d *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		if tok := c.GetHeader("X-Auth-Token"); tok != "" {
			var uid int64
			var exp *time.Time
			if err := d.Pool.QueryRow(
				ctx, `SELECT user_id, expires_at FROM auth_tokens WHERE token=$1`, tok,
			).Scan(&uid, &exp); err == nil {
				if exp == nil || exp.After(time.Now()) {
					c.Set("user_id", uid)
				}
			}
		}
		// 兼容老的 X-User-ID（如需）
		if _, exists := c.Get("user_id"); !exists {
			if v := c.GetHeader("X-User-ID"); v != "" {
				// 这里按你的老逻辑解析字符串为 int64，然后 c.Set("user_id", uid)
			}
		}
		c.Next()
	}
}

// ====== 工具函数（保留你原来的实现也可）======

func nullIfEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" // unique_violation
	}
	return false
}

// ---- 密码哈希/校验（和你原来一致）----

func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return "argon2id$1$65536$4$32$" +
		base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(hash), nil
}

func verifyPassword(stored, password string) bool {
	parts := strings.Split(stored, "$")
	if len(parts) < 7 {
		return false
	}
	salt, err1 := base64.RawStdEncoding.DecodeString(parts[5])
	sum, err2 := base64.RawStdEncoding.DecodeString(parts[6])
	if err1 != nil || err2 != nil {
		return false
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, uint32(len(sum)))
	return subtleConstantCompare(hash, sum)
}

func subtleConstantCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func randomToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}
