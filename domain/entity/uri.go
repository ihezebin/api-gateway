package entity

type Uri struct {
	// 代理路由的path
	Paths []string `json:"paths"`
	// 路径重写, 两个参数  如：{"/test": "/"} ,则：http://www.test.com/test/ping ->  http://www.test.com/ping
	Rewrite map[string]string `json:"rewrite" mapstructure:"rewrite"`
	// 路径代理到的具体服务
	Endpoint string `json:"endpoint"`
	// 代理描述
	Desc string `json:"desc"`
	// 优先级，值越大，优先级越高
	Priority int `json:"priority"`
	// 是否启用认证中间件
	Auth bool `json:"auth"`
}
