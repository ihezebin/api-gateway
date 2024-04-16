package middleware

import (
	"api-gateway/component/constant"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/jwt"
	"github.com/ihezebin/oneness/httpserver"
	"github.com/ihezebin/oneness/logger"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		needAuth := c.GetBool(constant.AuthenticationKey)
		if !needAuth {
			c.Next()
			return
		}

		body := &httpserver.Body{}
		tokenStr := c.GetHeader(constant.HeaderKeyToken)
		if tokenStr == "" {
			body.WithErrorx(httpserver.ErrorWithCode(httpserver.CodeUnauthorized))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		ctx := c.Request.Context()
		token, err := jwt.Parse(tokenStr, constant.TokenSecret)
		if err != nil {
			logger.WithError(err).Errorf(ctx, "parse token error, token: %s", tokenStr)
			body.WithErrorx(httpserver.ErrorWithCode(httpserver.CodeUnauthorized))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		faked, err := token.Faked()
		if err != nil {
			logger.WithError(err).Errorf(ctx, "faked token error, token: %s", tokenStr)
			body.WithErrorx(httpserver.ErrorWithCode(httpserver.CodeInternalServerError))
			c.AbortWithStatusJSON(http.StatusInternalServerError, body)
			return
		}

		if faked {
			body.WithErrorx(httpserver.ErrorWithCode(httpserver.CodeUnauthorized))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		if token.Expired() {
			body.WithErrorx(httpserver.ErrorWithCode(httpserver.CodeUnauthorized))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		c.Request.Header.Set(constant.HeaderKeyUid, token.Payload().Owner)
		c.Next()
	}
}
