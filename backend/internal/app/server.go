package app

import (
	"context"
	"net/http"

	"secrethole/backend/internal/api"
	"secrethole/backend/internal/config"
	"secrethole/backend/internal/db"

	"github.com/gin-gonic/gin"
)

type Server struct {
	Cfg config.Config
	DB  *db.DB
}

func NewServer(cfg config.Config, d *db.DB) *Server { return &Server{Cfg: cfg, DB: d} }

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(CORSMiddleware(s.Cfg.CorsOrigins))

	v1 := r.Group("/v1") {
    // 不需要登录
    v1.POST("/auth/register", api.RegisterHandler(s.DB))
    v1.POST("/auth/login",    api.LoginHandler(s.DB))

    // 需要登录
    v1.Use(api.AuthMiddleware(s.DB))
    {
        v1.GET("/me/profile",  api.MeProfileGet(s.DB))
        v1.PUT("/me/profile",  api.MeProfilePut(s.DB))
        v1.POST("/auth/logout", api.LogoutHandler(s.DB))

        v1.GET("/me/quota",            api.MeQuota(s.DB))
        v1.POST("/messages",           api.PostMessage(s.DB))
        v1.POST("/messages/claim",     api.ClaimMessage(s.DB))
        v1.POST("/messages/:id/ack-delete", api.AckDelete())
        v1.POST("/presence/heartbeat", api.PresenceHeartbeat(s.DB))
        v1.POST("/messages/:id/reply", api.ReplyMessage(s.DB))
        v1.POST("/dm/send",            api.SendDM(s.DB))
    }
}
	return r
}

func (s *Server) Start(addr string) error {
	return http.ListenAndServe(addr, s.Router())
}

func (s *Server) Migrate(ctx context.Context) error { return db.Migrate(ctx, s.DB) }
