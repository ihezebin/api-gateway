package middleware

import (
	"api-gateway/domain/entity"
	"fmt"

	"github.com/gin-gonic/gin"
)

type Matcher struct {
	endpointName2HostM map[string]string
	header2RuleM       map[string]*entity.Rule
	domain2RuleM       map[string]*entity.Rule
	globalRule         *entity.Rule
}

var matcher *Matcher

func RuleMatcher(endpoints []*entity.Endpoint, rules []*entity.Rule) gin.HandlerFunc {
	endpointName2HostM := make(map[string]string)
	header2RuleM := make(map[string]*entity.Rule)
	domain2RuleM := make(map[string]*entity.Rule)
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

		if rule.Headers != nil {
			for headerK, headerV := range rule.Headers {
				header2RuleM[fmt.Sprintf("%s=%s", headerK, headerV)] = rule
			}
		}
		if rule.Domains != nil {
			for _, domain := range rule.Domains {
				domain2RuleM[domain] = rule
			}
		}
	}

	if globalRule == nil {
		globalRule = new(entity.Rule)
	}

	matcher = &Matcher{
		endpointName2HostM: endpointName2HostM,
		header2RuleM:       header2RuleM,
		domain2RuleM:       domain2RuleM,
		globalRule:         globalRule,
	}

	return func(c *gin.Context) {
		c.Next()
	}
}
