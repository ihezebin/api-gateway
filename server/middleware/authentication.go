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

		body := &httpserver.Body{}
		tokenStr := c.GetHeader(constant.HeaderKeyToken)
		if tokenStr == "" {
			if needAuth {
				body.WithErrorx(httpserver.ErrWithUnAuthorized())
				c.AbortWithStatusJSON(http.StatusUnauthorized, body)
				return
			}

			c.Next()
			return
		}

		ctx := c.Request.Context()
		token, err := jwt.Parse(tokenStr, constant.TokenSecret)
		if err != nil {
			logger.WithError(err).Errorf(ctx, "parse token error, token: %s", tokenStr)
			body.WithErrorx(httpserver.ErrorWithAuthorizationFailed(err.Error()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		faked, err := token.Faked()
		if err != nil {
			logger.WithError(err).Errorf(ctx, "faked token error, token: %s", tokenStr)
			body.WithErrorx(httpserver.ErrorWithInternalServer())
			c.AbortWithStatusJSON(http.StatusInternalServerError, body)
			return
		}

		if faked {
			body.WithErrorx(httpserver.ErrorWithAuthorizationFailed("伪造的令牌"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		if token.Expired() {
			body.WithErrorx(httpserver.ErrorWithAuthorizationFailed("令牌已过期"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		// 传递账号ID
		c.Request.Header.Set(constant.HeaderKeyUid, token.Payload().Owner)

		//query := c.Request.URL.Query()
		//query.Set(constant.QueryKeyUid, token.Payload().Owner)
		//c.Request.URL.RawQuery = query.Encode()

		c.Next()
	}
}
