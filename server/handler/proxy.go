package handler

import (
	"api-gateway/component/constant"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/oneness/httpclient"
	"github.com/ihezebin/oneness/httpserver"
	"github.com/ihezebin/oneness/logger"
)

func Proxy(c *gin.Context) {
	ctx := c.Request.Context()
	body := &httpserver.Body{}

	timeout := c.GetInt(constant.ProxyTimeoutKey)
	newPath := c.GetString(constant.ProxyPathKey)
	newHost := c.GetString(constant.ProxyHostKey)

	hostUrl, err := url.Parse(newHost)
	if err != nil {
		logger.WithError(err).Errorf(ctx, "parse url error, url: %s", newHost)
		body.WithErrorx(httpserver.ErrorWithCode(httpserver.CodeInternalServerError))
		c.AbortWithStatusJSON(http.StatusInternalServerError, body)
		return
	}

	request := c.Request
	oldPath := request.URL.Path

	newUrl := request.URL
	newUrl.Host = hostUrl.Host
	newUrl.Scheme = hostUrl.Scheme
	newUrl.Path = newPath

	request.URL = newUrl
	// 重置为新 request
	request.RequestURI = ""

	logger.Infof(ctx, "%s [%s] -→ [%s][%s][%s]", c.Request.Method, oldPath, newHost, newPath, newUrl.RawQuery)

	response, err := httpclient.Client().SetTimeout(time.Duration(timeout) * time.Second).GetClient().Do(request)
	if err != nil {
		logger.WithError(err).Errorf(ctx, "http client do error")
		body.WithErrorx(httpserver.ErrorWithCode(httpserver.CodeInternalServerError))
		c.AbortWithStatusJSON(http.StatusInternalServerError, body)
		return
	}

	for k, vs := range response.Header {
		for _, v := range vs {
			c.Writer.Header().Add(k, v)
		}
	}
	c.Writer.WriteHeader(response.StatusCode)
	if _, err = io.Copy(c.Writer, response.Body); err != nil {
		logger.WithError(err).Errorf(ctx, "io copy error")
	}
}
