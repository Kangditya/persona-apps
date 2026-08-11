package auth

import "github.com/gin-gonic/gin"

func (s *Service) RegisterRoutes(group *gin.RouterGroup) {
    group.GET("/auth/login", s.login)
    group.GET("/auth/callback", s.callback)
    group.GET("/auth/session", s.session)
    group.POST("/auth/logout", s.logout)
}
