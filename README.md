# api-gateway

一个带有认证功能的 API 代理服务，可以基于 domain、header 匹配后再进行 uri 匹配、重写、代理。

## endpoint
能代理到的服务列表

## rule
匹配规则，可基于域名domain、header 或 uri 进行匹配。

domain 和 header 都可选，当两者同时存在时，取且判断；当同时为空时，视为全局的 uri 匹配。

当出现重复的 rule 时，其下的 uris 会合并

### uri
priority 越大优先级越高，auth 为 true 时会验证 token，验证通过后会将 token 中的 id 放到 header 中。

uri 指向的 endpoint 表示该路径具体代理到的上述服务。

rewrite 可以重写路径，比如 /api/v1/* 可以重写为 /api/*。

## 示例
代理服务配置 `config.toml`
```toml
service_name = "api-gateway"
port = 80

[logger]
    level = "debug"
    filename = "api-gateway.log"

[redis]
    addr = "127.0.0.1:6379"
    password = "root"

[[endpoints]]
    name = "go-template-ddd"
    host = "http://127.0.0.1:8080"

[[endpoints]]
    name = "user-service"
    host = "http://127.0.0.1:9000"

[[endpoints]]
    name = "blog"
    host = "http://127.0.0.1:9001"
```

全局规则配置 `global.toml`, 无需
```toml
[[uris]]
    paths = ["/global/*"]
    endpoint = "go-template-ddd"
    auth = true
    desc = "测试"
    priority = 0
    [uris.rewrite]
        "/global" = "/"
```

目标服务和地址为：`http://127.0.0.1:8080/health?a=1&b=2`

实际访问的代理地址为：
```shell
# hosts: hezebin.com 127.0.0.1
curl --location 'http://hezebin.com/global/health?a=1&b=2' \
--header 'Token: eyJlbmNvZGUiOiJiYXNlNjRyYXd1cmwiLCJ0eXAiOiJqd3QiLCJhbGciOiJIU0EyNTYifQ.eyJpc3N1ZXIiOiJnaXRodWIuY29tL2loZXplYmluL2p3dCIsIm93bmVyIjoiaGV6ZWJpbiIsInB1cnBvc2UiOiJhdXRoZW50aWNhdGlvbiIsImlzc3VlZF9hdCI6IjIwMjQtMDQtMTZUMTU6MTA6MjUuMzgzNjExKzA4OjAwIiwiZXhwaXJlIjozMDAwMDAwMDAwMCwiZXh0ZXJuYWwiOnsia2V5IjoidmFsdWUifX0.8c7YT36gDAcNESdOE318AkzTsvsRxXGcMEOYoq2OzmQ'
```
> Token 生成：https://github.com/ihezebin/jwt

日志内容为：
```json
{"body":"","file":"logging.go:28","func":"github.com/ihezebin/oneness/httpserver/middleware.LoggingRequest.func1","level":"info","method":"GET","msg":"incoming http request","remote":"127.0.0.1:62604","service":"api-gateway","time":"2024-04-16 15:33:54","timestamp":1713252834,"uri":"/global/health?a=1\u0026b=2"}

{"file":"proxy.go:44","func":"api-gateway/server/handler.Proxy","level":"info","msg":"GET [/global/health] => [/health]","service":"api-gateway","time":"2024-04-16 15:33:54","timestamp":1713252834}

{"body":"ok","file":"logging.go:66","func":"github.com/ihezebin/oneness/httpserver/middleware.LoggingResponse.func1","level":"info","msg":"outgoing http response","service":"api-gateway","status":"200 OK","time":"2024-04-16 15:33:54","timestamp":1713252834}
```

## Todo
- 可根据实际情况实现网关服务的热更新，将配置持久化，可参考 DDD 项目目录：[https://github.com/ihezebin/ddd](https://github.com/ihezebin/ddd)
- 高并发控制和限流