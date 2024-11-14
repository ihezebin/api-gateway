package middleware

import (
	"api-gateway/component/constant"
	"api-gateway/domain/entity"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/oneness/httpserver"
	"github.com/ihezebin/oneness/logger"
	"github.com/pkg/errors"
)

func RuleMatcher(endpoints []*entity.Endpoint, rules []*entity.Rule) (gin.HandlerFunc, error) {
	m, err := initialize(endpoints, rules)
	if err != nil {
		return nil, err
	}

	return func(c *gin.Context) {
		domain := c.Request.Host
		header := c.Request.Header
		path := c.Request.URL.Path

		ctx := c.Request.Context()

		rule := m.FindRule(domain, header)
		if rule == nil {

			logger.Infof(ctx, "not match rule. domain: %s, header: %s", domain, header)
			body := &httpserver.Body{}
			body.WithErrorx(httpserver.NewError(httpserver.CodeNotFound, fmt.Sprintf("%s + %s 未注册域名到网关", domain, path)))
			c.AbortWithStatusJSON(http.StatusNotFound, body)
			return
		}

		timeout := rule.Timeout
		if timeout == 0 {
			timeout = 10
		}
		uri, newPath := rule.FindPath(path)

		if uri == nil {
			logger.Infof(ctx, "not match path. domain: %s, header: %s, path: %s", domain, header, path)
			body := &httpserver.Body{}
			body.WithErrorx(httpserver.NewError(httpserver.CodeNotFound, fmt.Sprintf("%s + %s 未注册路由到网关", domain, path)))
			c.AbortWithStatusJSON(http.StatusNotFound, body)
			return
		}
		host := m.FindEndpointHost(uri.Endpoint)

		c.Set(constant.ProxyPathKey, newPath)
		c.Set(constant.ProxyHostKey, host)
		c.Set(constant.ProxyTimeoutKey, timeout)
		c.Set(constant.AuthenticationKey, uri.Auth)

		c.Next()
	}, nil
}

type matcher struct {
	endpointName2HostM map[string]string
	domain2RuleM       map[string]*entity.Rule
	globalRules        []*entity.Rule
}

func headerKey(headerK string, headerV ...string) string {
	return fmt.Sprintf("%s=%s", headerK, strings.Join(headerV, ","))
}

func domainAndHeaderKey(domain string, headerKey string) string {
	return fmt.Sprintf("%s&%s", domain, headerKey)
}

func (m *matcher) FindRule(domain string, header http.Header) *entity.Rule {
	// 匹配 domain
	rule, ok := m.domain2RuleM[domain]
	if ok {
		// 匹配 header
		for k, v := range rule.Headers {
			if header.Get(k) != v {
				return nil
			}
		}
		return rule
	}

	// 全局规则
	for _, globalRule := range m.globalRules {
		// 匹配 header
		for k, v := range globalRule.Headers {
			if header.Get(k) != v {
				return nil
			}
		}
		return globalRule
	}

	return nil
}

func (m *matcher) FindEndpointHost(name string) string {
	return m.endpointName2HostM[name]
}

func initialize(endpoints []*entity.Endpoint, rules []*entity.Rule) (*matcher, error) {
	endpointName2HostM := make(map[string]string)
	domain2RuleM := make(map[string]*entity.Rule)
	globalRules := make([]*entity.Rule, 0)

	endpointNames := make([]string, 0)
	for _, endpoint := range endpoints {
		endpointName2HostM[endpoint.Name] = endpoint.Host
		endpointNames = append(endpointNames, endpoint.Name)
	}

	domains := make([]string, 0)
	for _, rule := range rules {
		// 全局规则
		if len(rule.Domains) == 0 {
			globalRules = append(globalRules, rule)
			continue
		}

		for _, domain := range rule.Domains {
			domains = append(domains, domain)
			_, ok := domain2RuleM[domain]
			if ok {
				return nil, errors.Errorf("domain duplicate: %s", domain)
			}
			domain2RuleM[domain] = rule
		}
	}

	// 排序
	for key, rule := range domain2RuleM {
		domain2RuleM[key] = rule.SortUris()
	}
	for _, rule := range globalRules {
		rule.SortUris()
	}

	logger.Infof(context.Background(), "init rule matcher success, register domains: %+v, global rules len: %d, endpoints: %+v", domains, len(globalRules), endpointNames)

	return &matcher{
		endpointName2HostM: endpointName2HostM,
		domain2RuleM:       domain2RuleM,
		globalRules:        globalRules,
	}, nil
}
