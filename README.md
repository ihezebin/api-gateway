# api-gateway

## endpoint
能代理到的服务列表

## rule
匹配规则，可基于域名domain、header 或 uri 进行匹配。

domain 和 header 都可选，当两者同时存在时，取且判断；当同时为空时，直接做 uri 匹配。

### uri
priority 越大优先级越高，auth 为 true 时会验证 token，验证通过后会将 token 中的 id 放到 header 中。

uri 指向的 endpoint 表示该路径具体代理到的上述服务。