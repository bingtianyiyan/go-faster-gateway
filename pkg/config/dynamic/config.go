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
	BalanceMode BalanceMode `json:"balanceMode" toml:"balanceMode,omitempty" yaml:"balanceMode" `
	//全局中间件
	GlobalMiddleware []string `json:"globalMiddleware" toml:"globalMiddleware,omitempty" yaml:"globalMiddleware"`
	//api对应的路由配置
	EasyServiceRoute *ServiceRouteConfiguration `json:"easyServiceRoute,omitempty" toml:"easyServiceRoute,omitempty" yaml:"easyServiceRoute,omitempty"`
}

// 代理的负载均衡策略名
type BalanceMode struct {
	Balance string `json:"balance,omitempty" toml:"balance,omitempty" yaml:"balance,omitempty"`
}

// ServiceRouteConfiguration 代理的服务信息
type ServiceRouteConfiguration struct {
	//服务名-配置信息,一个服务下面可能有多种比如http或者ws
	//第一个key是某个上游的服务的名称 比如userSystem,第二个key是对应区分不同协议下的http类型/websocket类型服务定义唯一的名称，比如userSystem_http,userSystem_websocket
	Services map[string]map[string]*ServiceRoute `json:"services,omitempty" toml:"services,omitempty" yaml:"services,omitempty"`
}

// ServiceRoute 服务配置
type ServiceRoute struct {
	//上游的服务名称
	ServiceName string `json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty"`
	//负载均衡策略
	BalanceMode string `json:"balanceMode" toml:"balanceMode,omitempty" yaml:"balanceMode,omitempty" `
	//协议(http,https,websocket,tcp,udp)
	Handler string `json:"handler,omitempty" toml:"handler,omitempty" yaml:"handler,omitempty" `
	//代理的路由配置信息 路由列表
	Routers []Router `json:"routers,omitempty" toml:"routers,omitempty" yaml:"routers,omitempty"`
	//代理的目标服务信息
	Servers []Server `json:"servers,omitempty" toml:"servers,omitempty" yaml:"servers,omitempty"`
	//对应的中间件
	Middlewares []string `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty"`
}

// 代理的路由信息
type Router struct {
	// api_path 请求的api的url配置
	Path string `json:"path,omitempty"  toml:"path,omitempty" yaml:"path,omitempty"`
	// 请求方法(*,GET,POST,DELETE
	Methods []string `json:"methods,omitempty"  toml:"methods,omitempty" yaml:"methods,omitempty"`
	//路由类型  static/param/wildcard/subrouter
	Type string `json:"type,omitempty" toml:"type,omitempty" yaml:"type,omitempty"`
	//路由前缀
	Prefix string `json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty"`
	// proxy_path 目标代理的api的url配置
	ProxyPath string `json:"proxyPath,omitempty"  toml:"proxyPath,omitempty" yaml:"proxyPath,omitempty"`
	// 规则
	Rule   string            `json:"rule,omitempty" toml:"rule,omitempty" yaml:"rule,omitempty"`
	Params map[string]string `yaml:"params,omitempty" json:"params"` // 参数约束
}

// RouteParamConfig 路由参数约束
type RouteParamConfig struct {
	Pattern string ` json:"pattern" yaml:"pattern"` // 参数正则约束
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
