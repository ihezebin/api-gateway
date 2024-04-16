package entity

import (
	"regexp"
	"sort"
	"strings"
)

type Rule struct {
	Domains []string          `json:"domains" mapstructure:"domains"`
	Headers map[string]string `json:"headers" mapstructure:"headers"`
	// Timeout seconds
	Timeout int   `json:"timeout" mapstructure:"timeout"`
	Uris    []Uri `json:"uris" mapstructure:"uris"`
}

func (r *Rule) SortUris() *Rule {
	sort.SliceStable(r.Uris, func(i, j int) bool {
		return r.Uris[i].Priority > r.Uris[j].Priority
	})
	return r
}

func (r *Rule) FindPath(path string) (*Uri, string) {
	for _, uri := range r.Uris {
		for _, pattern := range uri.Paths {
			matched, _ := regexp.MatchString(pattern, path)
			if matched {
				// 路径重写
				for k, v := range uri.Rewrite {
					path = strings.ReplaceAll(path, k, v)
				}

				// 修复url路径
				if !strings.HasPrefix(path, "/") {
					path = "/" + path
				}

				path = strings.ReplaceAll(path, `//`, `/`)

				return &uri, path
			}
		}

	}
	return nil, ""
}
