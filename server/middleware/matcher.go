package middleware

import (
	"api-gateway/component/constant"
	"api-gateway/domain/entity"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ihezebin/oneness/logger"
)

func RuleMatcher(endpoints []*entity.Endpoint, rules []*entity.Rule) gin.HandlerFunc {
	initMatcher(endpoints, rules)

	return func(c *gin.Context) {
		domain := c.Request.Host
		header := c.Request.Header
		path := c.Request.URL.Path

		ctx := c.Request.Context()

		rule := matcher.FindRule(domain, header)

		timeout := rule.Timeout
		if timeout == 0 {
			timeout = 10
		}
		uri, newPath := rule.FindPath(path)

		if uri == nil {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		host := matcher.FindEndpointHost(uri.Endpoint)
		logger.Infof(ctx, "domain: %s, header: %s, path: %s, endpoint: %s, uri: %s", domain, header, path, uri.Endpoint, uri.Path)

		c.Set(constant.ProxyPathKey, newPath)
		c.Set(constant.ProxyHostKey, host)
		c.Set(constant.ProxyTimeoutKey, timeout)
		c.Set(constant.AuthenticationKey, uri.Auth)

		c.Next()
	}
}

type Matcher struct {
	endpointName2HostM map[string]string
	header2RuleM       map[string]*entity.Rule
	domain2RuleM       map[string]*entity.Rule
	domainAndHeader2M  map[string]*entity.Rule
	globalRule         *entity.Rule
}

var matcher *Matcher

func headerKey(headerK string, headerV ...string) string {
	return fmt.Sprintf("%s=%s", headerK, strings.Join(headerV, ","))
}

func domainAndHeaderKey(domain string, headerKey string) string {
	return fmt.Sprintf("%s&%s", domain, headerKey)
}

func (m *Matcher) FindRule(domain string, header http.Header) *entity.Rule {
	var headerRule *entity.Rule
	// 优先匹配 domain&header
	for headerK, headerV := range header {
		rule, ok := m.domain2RuleM[domainAndHeaderKey(domain, headerKey(headerK, headerV...))]
		if ok {
			return rule
		}

		headerRule = m.header2RuleM[headerKey(headerK, headerV...)]
	}

	// 仅 domain 匹配
	for k, rule := range m.domain2RuleM {
		if k == domain {
			return rule
		}
	}

	// 仅 header 匹配
	if headerRule != nil {
		return headerRule
	}

	// 全局规则
	return m.globalRule
}

func (m *Matcher) FindEndpointHost(name string) string {
	return m.endpointName2HostM[name]
}

func initMatcher(endpoints []*entity.Endpoint, rules []*entity.Rule) {
	endpointName2HostM := make(map[string]string)
	header2RuleM := make(map[string]*entity.Rule)
	domain2RuleM := make(map[string]*entity.Rule)
	domainAndHeader2M := make(map[string]*entity.Rule)
	var globalRule *entity.Rule

	for _, endpoint := range endpoints {
		endpointName2HostM[endpoint.Name] = endpoint.Host
	}

	for _, rule := range rules {
		if len(rule.Headers) == 0 && len(rule.Domains) == 0 {
			if globalRule == nil {
				globalRule = rule
			} else { // 合并规则
				globalRule.Uris = append(globalRule.Uris, rule.Uris...)
			}
			continue
		}

		if len(rule.Headers) > 0 && len(rule.Domains) > 0 {
			for _, domain := range rule.Domains {
				for headerK, headerV := range rule.Headers {
					key := domainAndHeaderKey(domain, headerKey(headerK, headerV))
					r, ok := domainAndHeader2M[key]
					if ok {
						r.Uris = append(r.Uris, rule.Uris...)
					} else {
						r = rule
					}

					domainAndHeader2M[key] = r
				}
			}
			continue
		}

		for headerK, headerV := range rule.Headers {
			key := fmt.Sprintf("%s=%s", headerK, headerV)
			r, ok := header2RuleM[key]
			if ok {
				r.Uris = append(r.Uris, rule.Uris...)
			} else {
				r = rule
			}
			header2RuleM[key] = r
		}

		for _, domain := range rule.Domains {
			r, ok := domain2RuleM[domain]
			if ok {
				r.Uris = append(r.Uris, rule.Uris...)
			} else {
				r = rule
			}
			domain2RuleM[domain] = r
		}
	}

	if globalRule == nil {
		globalRule = new(entity.Rule)
	}

	// 排序
	for key, rule := range header2RuleM {
		header2RuleM[key] = rule.SortUris()
	}

	for key, rule := range domain2RuleM {
		domain2RuleM[key] = rule.SortUris()
	}

	for key, rule := range domainAndHeader2M {
		domainAndHeader2M[key] = rule.SortUris()
	}

	globalRule = globalRule.SortUris()

	matcher = &Matcher{
		endpointName2HostM: endpointName2HostM,
		header2RuleM:       header2RuleM,
		domain2RuleM:       domain2RuleM,
		globalRule:         globalRule,
	}
}
