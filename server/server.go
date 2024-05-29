package server

import (
	"api-gateway/config"
	"api-gateway/server/handler"
	"api-gateway/server/middleware"
	"context"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/ihezebin/oneness/httpserver"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// Run server
// @title Go Template DDD 示例 API 文档
// @version 1.0
// @description 这是一个使用 Gin 和 Swagger 生成 API 文档的示例。
// @host localhost:8080
// @BasePath /
func Run(ctx context.Context, conf *config.Config) error {
	serverHandler := httpserver.NewServerHandlerWithOptions(
		httpserver.WithLoggingRequest(true),
		httpserver.WithLoggingResponse(false),
		httpserver.WithMiddlewares(
			middleware.Cors(),
			middleware.RuleMatcher(conf.Endpoints, conf.Rules),
			middleware.Authentication(),
		),
	)

	pprof.Register(serverHandler)
	serverHandler.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	serverHandler.GET("/metrics", gin.WrapH(promhttp.Handler()))
	serverHandler.NoRoute(handler.Proxy)

	httpserver.ResetServerHandler(serverHandler)

	return httpserver.Run(ctx, conf.Port)
}
