package dynamic

import "go-faster-gateway/pkg/database"

// Message holds configuration information exchanged between parts of gateway
type Message struct {
	ProviderName  string
	Configuration *Configuration
}

// Configurations is for currentConfigurations Map.
type Configurations map[string]*Configuration

// Configuration is the root of the dynamic configuration.
type Configuration struct {
	//数据库
	Databases *database.Database `description:"gateway db settings." json:"databases,omitempty" toml:"databases,omitempty" yaml:"databases,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`

	//负载均衡策略
	BalanceMode BalanceMode `json:"balanceMode"  yaml:"balanceMode" `
	//全局中间件
	GlobalMiddleware []string `json:"globalMiddleware" yaml:"globalMiddleware"`
	//api对应的路由配置
	HTTP *HTTPConfiguration `json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty"`
}

// 代理的负载均衡策略名
type BalanceMode struct {
	Balance string `json:"balance,omitempty" toml:"balance,omitempty" yaml:"balance,omitempty"`
}

// HTTPConfiguration 代理的服务信息
type HTTPConfiguration struct {
	//服务名-配置信息
	Services map[string]*Service `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty" export:"true"`
}

// Service 服务配置
type Service struct {
	//上游得服务名称
	Service string `json:"service,omitempty" toml:"service,omitempty" yaml:"service,omitempty"`
	//代理的路由配置信息
	Routers Router `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty"`
	//代理的目标服务信息
	Servers []Server `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty"`
	//对应的中间件
	Middleware []string `json:"middleware,omitempty" toml:"middleware,omitempty" yaml:"middleware,omitempty"`
}

// 代理的路由信息
type Router struct {
	// api_path 请求的api的url配置
	Path string `json:"path,omitempty"  toml:"path,omitempty" yaml:"path,omitempty"`
	// 请求方法(*,GET,POST,DELETE
	Method string `json:"method,omitempty"  toml:"method,omitempty" yaml:"method,omitempty"`
	// proxy_path 目标代理的api的url配置
	ProxyPath string `json:"proxyPath,omitempty"  toml:"proxyPath,omitempty" yaml:"proxyPath,omitempty"`
	// 规则
	Rule string `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
}

// Server 目标服务配置
type Server struct {
	//host地址
	Host string `json:"host,omitempty" toml:"host,omitempty" yaml:"host,omitempty"`
	//端口
	Port uint64 `json:"port,omitempty" toml:"port,omitempty" yaml:"port,omitempty"`
	//权重
	Weight int `json:"weight,omitempty" toml:"weight,omitempty" yaml:"weight,omitempty"`
	//是否健康
	Healthy bool `json:"healthy,omitempty" toml:"healthy,omitempty" yaml:"healthy,omitempty"`
}
