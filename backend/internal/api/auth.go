package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/argon2"
)

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

func RegisterHandler(c *gin.Context) {
	return func(c *gin.Context) {
		db := mustDB(c)
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
		// 简单校验“最多7个汉字”：应用层按 rune 计数（这里不强卡字形，前端也会限制）
		if runeLen(req.Name) > 7 {
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

		var userID int64
		err = db.QueryRow(`
		INSERT INTO users(name, avatar_base64, gender_color, password_hash)
		VALUES ($1,$2,$3,$4) RETURNING id
	`, req.Name, nullIfEmpty(req.AvatarBase64), nullIfEmpty(req.GenderColor), hash).Scan(&userID)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "name exists?"})
			return
		}
		c.JSON(http.StatusOK, RegisterResp{UserID: userID})
	}
}

func LoginHandler(c *gin.Context) {
	return func(c *gin.Context) {
		db := mustDB(c)
		var req LoginReq
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad json"})
			return
		}
		var hash string
		err := db.QueryRow(`SELECT password_hash FROM users WHERE id=$1`, req.UserID).Scan(&hash)
		if errors.Is(err, sql.ErrNoRows) {
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
		_, err = db.Exec(`INSERT INTO auth_tokens(token, user_id, expires_at) VALUES ($1,$2,$3)`,
			token, req.UserID, time.Now().Add(30*24*time.Hour))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "token issue"})
			return
		}
		c.JSON(http.StatusOK, LoginResp{UserID: req.UserID, Token: token})
	}
}

// 中间件：优先用 X-Auth-Token 解 user_id；兼容 X-User-ID（老接口）
func AuthMiddleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if tok := c.GetHeader("X-Auth-Token"); tok != "" {
			var uid int64
			var exp *time.Time
			err := db.QueryRow(`SELECT user_id, expires_at FROM auth_tokens WHERE token=$1`, tok).Scan(&uid, &exp)
			if err == nil {
				if exp == nil || exp.After(time.Now()) {
					c.Set("user_id", uid)
				}
			}
		}
		// 兼容旧逻辑
		if _, exists := c.Get("user_id"); !exists {
			if v := c.GetHeader("X-User-ID"); v != "" {
				// 你已有的解析逻辑：转为 int64；此处略
			}
		}
		c.Next()
	}
}

// utils
func mustDB(c *gin.Context) *sql.DB {
	dbi, ok := c.MustGet("db").(*sql.DB)
	if !ok {
		panic("db not in context")
	}
	return dbi
}
func runeLen(s string) int { return len([]rune(s)) }
func nullIfEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return &s
}

// ========== 密码哈希（Argon2id 简易实现） ==========
func hashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	// 参数：可以按你的安全/性能调整
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	return "argon2id$1$65536$4$32$" +
		base64.RawStdEncoding.EncodeToString(salt) + "$" +
		base64.RawStdEncoding.EncodeToString(hash), nil
}

func verifyPassword(stored, password string) bool {
	// 简化解析（生产建议用标准库封装/第三方库）
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
