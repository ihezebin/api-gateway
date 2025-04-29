package server

import (
	"context"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ihezebin/olympus/httpserver"
	"github.com/ihezebin/olympus/httpserver/middleware"
	"github.com/ihezebin/olympus/runner"
	"github.com/pkg/errors"

	"api-gateway/config"
	"api-gateway/server/router"
)

type Body struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func NewServer(ctx context.Context, conf *config.Config) (runner.Task, error) {
	matcher, err := RuleMatcher(conf.Endpoints, conf.Rules)
	if err != nil {
		return nil, errors.Wrapf(err, "init rule matcher error")
	}

	server := httpserver.NewServer(
		httpserver.WithPort(conf.Port),
		httpserver.WithServiceName(conf.ServiceName),
		httpserver.WithMiddlewares(
			middleware.Recovery(),
			middleware.LoggingRequestWithoutHeader(),
			middleware.LoggingResponseWithoutHeader(),
			matcher,
			Authentication(),
		),
		httpserver.WithOpenAPInfo(openapi3.Info{
			Version:     "1.0",
			Description: "这是一个使用 Gin 和 OpenAPI 生成 API 文档的示例。",
			Contact: &openapi3.Contact{
				Name:  "ihezebin",
				Email: "ihezebin@gmail.com",
			},
		}),
		httpserver.WithOpenAPIServer(openapi3.Server{
			URL:         fmt.Sprintf("http://localhost:%d", conf.Port),
			Description: "本地开发环境",
		}),
	)

	server.RegisterRoutes()
	server.Engine().NoRoute(router.Proxy)

	err = server.RegisterOpenAPIUI("/openapi", httpserver.StoplightUI)
	if err != nil {
		return nil, err
	}
	_ = server.RegisterOpenAPIUI("/redoc", httpserver.RedocUI)
	_ = server.RegisterOpenAPIUI("/rapidoc", httpserver.RapidocUI)
	_ = server.RegisterOpenAPIUI("/swagger", httpserver.SwaggerUI)

	return server, nil

}
