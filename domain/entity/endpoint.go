package entity

type Endpoint struct {
	// 服务名
	Name string `json:"name"`
	// 服务地址
	Host string `json:"host"`
}
