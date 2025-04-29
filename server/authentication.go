package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/jwt"
	"github.com/ihezebin/olympus/httpserver"
	"github.com/ihezebin/olympus/logger"

	"api-gateway/component/cache"
	"api-gateway/component/constant"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		needAuth := c.GetBool(constant.AuthenticationKey)

		body := &httpserver.Body[any]{}
		tokenStr := c.GetHeader(constant.HeaderKeyToken)
		if tokenStr == "" {
			if needAuth {
				body.WithErr(httpserver.ErrWithUnAuthorized())
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
			body.WithErr(httpserver.ErrorWithAuthorizationFailed(err.Error()))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		faked, err := token.Faked()
		if err != nil {
			logger.WithError(err).Errorf(ctx, "faked token error, token: %s", tokenStr)
			body.WithErr(httpserver.ErrorWithInternalServer())
			c.AbortWithStatusJSON(http.StatusInternalServerError, body)
			return
		}

		if faked {
			body.WithErr(httpserver.ErrorWithAuthorizationFailed("伪造的令牌"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		if token.Expired() {
			body.WithErr(httpserver.ErrorWithAuthorizationFailed("令牌已过期"))
			c.AbortWithStatusJSON(http.StatusUnauthorized, body)
			return
		}

		tokenKey := fmt.Sprintf(constant.TokenRedisKeyFormat, token.Payload().Owner)
		tokenVal, err := cache.RedisCacheClient().Get(ctx, tokenKey).Result()
		if err != nil {
			logger.WithError(err).Errorf(ctx, "get token from redis error, token: %s", tokenKey)
			body.WithErr(httpserver.ErrorWithInternalServer())
			c.AbortWithStatusJSON(http.StatusInternalServerError, body)
			return
		}

		if tokenVal != tokenStr {
			body.WithErr(httpserver.ErrorWithAuthorizationFailed("令牌被重置"))
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
