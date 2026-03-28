package router

import (
	"github.com/gin-gonic/gin"

	"bitbucket.org/appmax-space/go-boilerplate/internal/infrastructure/web/handler"
)

// RegisterUserRoutes registra todas as rotas relacionadas a Entity
func RegisterUserRoutes(rg *gin.RouterGroup, h *handler.UserHandler) {
	rg.POST("/entities", h.Create)
	rg.GET("/entities", h.List)
	rg.GET("/entities/:id", h.GetByID)
	rg.PUT("/entities/:id", h.Update)
	rg.DELETE("/entities/:id", h.Delete)
}
