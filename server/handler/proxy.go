package handler

import (
	"api-gateway/component/constant"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/oneness/logger"
)

func Proxy(c *gin.Context) {
	ctx := c.Request.Context()

	newPath, err := url.Parse(c.GetString(constant.ProxyPathKey))
	if err != nil {
		c.AbortWithStatus(http.StatusBadGateway)
		return
	}
	// 代理请求
	oldPath := c.Request.URL.Path
	req := new(http.Request)
	*req = *(c.Request)
	req.URL = newPath
	req = req.WithContext(ctx)
	req.RequestURI = ""

	logger.Infof(ctx, "%s [%s] => [%s]", c.Request.Method, oldPath, req.URL.String())

	//resp, err := httpc.GetGlobalClient().Kernel().GetClient().Do(req)
	//if err != nil {
	//	c.JSON(http.StatusBadGateway, result.Failed(err))
	//	return
	//}
	//defer func() {
	//	_ = resp.Body.Close()
	//}()

	//for key, values := range resp.Header {
	//	for _, value := range values {
	//		c.Writer.Header().Add(key, value)
	//	}
	//}
	//c.Writer.WriteHeader(resp.StatusCode)
	//_, _ = io.Copy(c.Writer, resp.Body)
}
