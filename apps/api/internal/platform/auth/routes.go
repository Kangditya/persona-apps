package auth

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func (s *Service) RegisterRoutes(group gin.IRouter) {
    group.GET("/auth/login", s.login)
    group.GET("/auth/callback", s.callback)
    group.GET("/auth/session", s.session)
    group.POST("/auth/logout", s.logout)
    for _, path := range []string{"/auth/login", "/auth/callback", "/auth/session", "/auth/logout"} {
        group.OPTIONS(path, func(c *gin.Context) { c.Status(http.StatusNoContent) })
    }
}
