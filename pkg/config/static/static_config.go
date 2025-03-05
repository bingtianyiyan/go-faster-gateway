package static

import (
	"fmt"
	"go-faster-gateway/pkg/log"
	"go-faster-gateway/pkg/provider/file"
	ptypes "go-faster-gateway/pkg/types"
	"time"
)

const (
	// DefaultInternalEntryPointName the name of the default internal entry point.
	DefaultInternalEntryPointName = "fastgateway"

	// DefaultGraceTimeout controls how long Traefik serves pending requests
	// prior to shutting down.
	DefaultGraceTimeout = 10 * time.Second

	// DefaultIdleTimeout before closing an idle connection.
	DefaultIdleTimeout = 180 * time.Second

	// DefaultReadTimeout defines the default maximum duration for reading the entire request, including the body.
	DefaultReadTimeout = 60 * time.Second

	// DefaultUDPTimeout defines how long to wait by default on an idle session,
	// before releasing all resources related to that session.
	DefaultUDPTimeout = 3 * time.Second
)

// Configuration is the static configuration.
type Configuration struct {
	ServersTransport    *ServersTransport          `description:"Servers default transport." json:"serversTransport,omitempty" toml:"serversTransport,omitempty" yaml:"serversTransport,omitempty" export:"true"`
	TCPServersTransport *TCPServersTransport       `description:"TCP servers default transport." json:"tcpServersTransport,omitempty" toml:"tcpServersTransport,omitempty" yaml:"tcpServersTransport,omitempty" export:"true"`
	EntryPoints         EntryPoints                `description:"Entry points definition." json:"entryPoints,omitempty" toml:"entryPoints,omitempty" yaml:"entryPoints,omitempty" export:"true"`
	Providers           *Providers                 `description:"Providers configuration." json:"providers,omitempty" toml:"providers,omitempty" yaml:"providers,omitempty" export:"true"`
	HostResolver        *ptypes.HostResolverConfig `description:"Enable CNAME Flattening." json:"hostResolver,omitempty" toml:"hostResolver,omitempty" yaml:"hostResolver,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	Core                *Core                      `description:"Core controls." json:"core,omitempty" toml:"core,omitempty" yaml:"core,omitempty" export:"true"`
	Spiffe              *SpiffeClientConfig        `description:"SPIFFE integration configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" export:"true"`
	//日志
	Logger *log.Logger `description:"gateway log settings." json:"log,omitempty" toml:"logger,omitempty" yaml:"logger,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// Core configures Traefik core behavior.
type Core struct {
	DefaultRuleSyntax string `description:"Defines the rule parser default syntax (v2 or v3)" json:"defaultRuleSyntax,omitempty" toml:"defaultRuleSyntax,omitempty" yaml:"defaultRuleSyntax,omitempty"`
}

// SetDefaults sets the default values.
func (c *Core) SetDefaults() {
	c.DefaultRuleSyntax = "v3"
}

// SpiffeClientConfig defines the SPIFFE client configuration.
type SpiffeClientConfig struct {
	WorkloadAPIAddr string `description:"Defines the workload API address." json:"workloadAPIAddr,omitempty" toml:"workloadAPIAddr,omitempty" yaml:"workloadAPIAddr,omitempty"`
}

// CertificateResolver contains the configuration for the different types of certificates resolver.
type CertificateResolver struct {
	//ACME      *acmeprovider.Configuration `description:"Enables ACME (Let's Encrypt) automatic SSL." json:"acme,omitempty" toml:"acme,omitempty" yaml:"acme,omitempty" export:"true"`
	Tailscale *struct{} `description:"Enables Tailscale certificate resolution." json:"tailscale,omitempty" toml:"tailscale,omitempty" yaml:"tailscale,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// ServersTransport options to configure communication between Traefik and the servers.
type ServersTransport struct {
	InsecureSkipVerify  bool                   `description:"Disable SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAs             []ptypes.FileOrContent `description:"Add cert file for self-signed certificate." json:"rootCAs,omitempty" toml:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	MaxIdleConnsPerHost int                    `description:"If non-zero, controls the maximum idle (keep-alive) to keep per-host. If zero, DefaultMaxIdleConnsPerHost is used" json:"maxIdleConnsPerHost,omitempty" toml:"maxIdleConnsPerHost,omitempty" yaml:"maxIdleConnsPerHost,omitempty" export:"true"`
	ForwardingTimeouts  *ForwardingTimeouts    `description:"Timeouts for requests forwarded to the backend servers." json:"forwardingTimeouts,omitempty" toml:"forwardingTimeouts,omitempty" yaml:"forwardingTimeouts,omitempty" export:"true"`
	Spiffe              *Spiffe                `description:"Defines the SPIFFE configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// Spiffe holds the SPIFFE configuration.
type Spiffe struct {
	IDs         []string `description:"Defines the allowed SPIFFE IDs (takes precedence over the SPIFFE TrustDomain)." json:"ids,omitempty" toml:"ids,omitempty" yaml:"ids,omitempty"`
	TrustDomain string   `description:"Defines the allowed SPIFFE trust domain." json:"trustDomain,omitempty" toml:"trustDomain,omitempty" yaml:"trustDomain,omitempty"`
}

// TCPServersTransport options to configure communication between Traefik and the servers.
type TCPServersTransport struct {
	DialKeepAlive ptypes.Duration `description:"Defines the interval between keep-alive probes for an active network connection. If zero, keep-alive probes are sent with a default value (currently 15 seconds), if supported by the protocol and operating system. Network protocols or operating systems that do not support keep-alives ignore this field. If negative, keep-alive probes are disabled" json:"dialKeepAlive,omitempty" toml:"dialKeepAlive,omitempty" yaml:"dialKeepAlive,omitempty" export:"true"`
	DialTimeout   ptypes.Duration `description:"Defines the amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	// TerminationDelay, corresponds to the deadline that the proxy sets, after one
	// of its connected peers indicates it has closed the writing capability of its
	// connection, to close the reading capability as well, hence fully terminating the
	// connection. It is a duration in milliseconds, defaulting to 100. A negative value
	// means an infinite deadline (i.e. the reading capability is never closed).
	TerminationDelay ptypes.Duration  `description:"Defines the delay to wait before fully terminating the connection, after one connected peer has closed its writing capability." json:"terminationDelay,omitempty" toml:"terminationDelay,omitempty" yaml:"terminationDelay,omitempty" export:"true"`
	TLS              *TLSClientConfig `description:"Defines the TLS configuration." json:"tls,omitempty" toml:"tls,omitempty" yaml:"tls,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// TLSClientConfig options to configure TLS communication between Traefik and the servers.
type TLSClientConfig struct {
	InsecureSkipVerify bool                   `description:"Disables SSL certificate verification." json:"insecureSkipVerify,omitempty" toml:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty" export:"true"`
	RootCAs            []ptypes.FileOrContent `description:"Defines a list of CA secret used to validate self-signed certificate" json:"rootCAs,omitempty" toml:"rootCAs,omitempty" yaml:"rootCAs,omitempty"`
	Spiffe             *Spiffe                `description:"Defines the SPIFFE TLS configuration." json:"spiffe,omitempty" toml:"spiffe,omitempty" yaml:"spiffe,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// RespondingTimeouts contains timeout configurations for incoming requests to the Traefik instance.
type RespondingTimeouts struct {
	ReadTimeout  ptypes.Duration `description:"ReadTimeout is the maximum duration for reading the entire request, including the body. If zero, no timeout is set." json:"readTimeout,omitempty" toml:"readTimeout,omitempty" yaml:"readTimeout,omitempty" export:"true"`
	WriteTimeout ptypes.Duration `description:"WriteTimeout is the maximum duration before timing out writes of the response. If zero, no timeout is set." json:"writeTimeout,omitempty" toml:"writeTimeout,omitempty" yaml:"writeTimeout,omitempty" export:"true"`
	IdleTimeout  ptypes.Duration `description:"IdleTimeout is the maximum amount duration an idle (keep-alive) connection will remain idle before closing itself. If zero, no timeout is set." json:"idleTimeout,omitempty" toml:"idleTimeout,omitempty" yaml:"idleTimeout,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (a *RespondingTimeouts) SetDefaults() {
	a.ReadTimeout = ptypes.Duration(DefaultReadTimeout)
	a.IdleTimeout = ptypes.Duration(DefaultIdleTimeout)
}

// ForwardingTimeouts contains timeout configurations for forwarding requests to the backend servers.
type ForwardingTimeouts struct {
	DialTimeout           ptypes.Duration `description:"The amount of time to wait until a connection to a backend server can be established. If zero, no timeout exists." json:"dialTimeout,omitempty" toml:"dialTimeout,omitempty" yaml:"dialTimeout,omitempty" export:"true"`
	ResponseHeaderTimeout ptypes.Duration `description:"The amount of time to wait for a server's response headers after fully writing the request (including its body, if any). If zero, no timeout exists." json:"responseHeaderTimeout,omitempty" toml:"responseHeaderTimeout,omitempty" yaml:"responseHeaderTimeout,omitempty" export:"true"`
	IdleConnTimeout       ptypes.Duration `description:"The maximum period for which an idle HTTP keep-alive connection will remain open before closing itself" json:"idleConnTimeout,omitempty" toml:"idleConnTimeout,omitempty" yaml:"idleConnTimeout,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (f *ForwardingTimeouts) SetDefaults() {
	f.DialTimeout = ptypes.Duration(30 * time.Second)
	f.IdleConnTimeout = ptypes.Duration(90 * time.Second)
}

// LifeCycle contains configurations relevant to the lifecycle (such as the shutdown phase) of Traefik.
type LifeCycle struct {
	RequestAcceptGraceTimeout ptypes.Duration `description:"Duration to keep accepting requests before Traefik initiates the graceful shutdown procedure." json:"requestAcceptGraceTimeout,omitempty" toml:"requestAcceptGraceTimeout,omitempty" yaml:"requestAcceptGraceTimeout,omitempty" export:"true"`
	GraceTimeOut              ptypes.Duration `description:"Duration to give active requests a chance to finish before Traefik stops." json:"graceTimeOut,omitempty" toml:"graceTimeOut,omitempty" yaml:"graceTimeOut,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (a *LifeCycle) SetDefaults() {
	a.GraceTimeOut = ptypes.Duration(DefaultGraceTimeout)
}

// Tracing holds the tracing configuration.
type Tracing struct {
	ServiceName             string            `description:"Sets the name for this service." json:"serviceName,omitempty" toml:"serviceName,omitempty" yaml:"serviceName,omitempty" export:"true"`
	ResourceAttributes      map[string]string `description:"Defines additional resource attributes (key:value)." json:"resourceAttributes,omitempty" toml:"resourceAttributes,omitempty" yaml:"resourceAttributes,omitempty" export:"true"`
	CapturedRequestHeaders  []string          `description:"Request headers to add as attributes for server and client spans." json:"capturedRequestHeaders,omitempty" toml:"capturedRequestHeaders,omitempty" yaml:"capturedRequestHeaders,omitempty" export:"true"`
	CapturedResponseHeaders []string          `description:"Response headers to add as attributes for server and client spans." json:"capturedResponseHeaders,omitempty" toml:"capturedResponseHeaders,omitempty" yaml:"capturedResponseHeaders,omitempty" export:"true"`
	SafeQueryParams         []string          `description:"Query params to not redact." json:"safeQueryParams,omitempty" toml:"safeQueryParams,omitempty" yaml:"safeQueryParams,omitempty" export:"true"`
	SampleRate              float64           `description:"Sets the rate between 0.0 and 1.0 of requests to trace." json:"sampleRate,omitempty" toml:"sampleRate,omitempty" yaml:"sampleRate,omitempty" export:"true"`
	AddInternals            bool              `description:"Enables tracing for internal services (ping, dashboard, etc...)." json:"addInternals,omitempty" toml:"addInternals,omitempty" yaml:"addInternals,omitempty" export:"true"`
}

// SetDefaults sets the default values.
func (t *Tracing) SetDefaults() {
	t.ServiceName = "gateway"
	t.SampleRate = 1.0
}

// Providers contains providers configuration.
type Providers struct {
	ProvidersThrottleDuration ptypes.Duration `description:"Backends throttle duration: minimum duration between 2 events from providers before applying a new configuration. It avoids unnecessary reloads if multiples events are sent in a short amount of time." json:"providersThrottleDuration,omitempty" toml:"providersThrottleDuration,omitempty" yaml:"providersThrottleDuration,omitempty" export:"true"`

	File *file.Provider `description:"Enable File backend with default settings." json:"file,omitempty" toml:"file,omitempty" yaml:"file,omitempty" export:"true"`

	//Rest              *rest.Provider    `description:"Enable Rest backend with default settings." json:"rest,omitempty" toml:"rest,omitempty" yaml:"rest,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	//consulcatalog.ProviderBuilder：这是专门用于构建和配置 Consul Catalog Provider 的构建器。
	//ConsulCatalog Provider 允许 Traefik 从 Consul 的服务发现目录中动态获取服务信息，并据此更新其路由规则。
	//它主要用于与 Consul 的服务注册与发现功能进行集成，实现服务的自动发现和动态路由管理‌
	//ConsulCatalog *consulcatalog.ProviderBuilder `description:"Enable ConsulCatalog backend with default settings." json:"consulCatalog,omitempty" toml:"consulCatalog,omitempty" yaml:"consulCatalog,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	//Nomad         *nomad.ProviderBuilder         `description:"Enable Nomad backend with default settings." json:"nomad,omitempty" toml:"nomad,omitempty" yaml:"nomad,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	//Ecs           *ecs.Provider                  `description:"Enable AWS ECS backend with default settings." json:"ecs,omitempty" toml:"ecs,omitempty" yaml:"ecs,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	//Consul        *consul.ProviderBuilder        `description:"Enable Consul backend with default settings." json:"consul,omitempty" toml:"consul,omitempty" yaml:"consul,omitempty" label:"allowEmpty" file:"allowEmpty"  export:"true"`
	//Etcd          *etcd.Provider                 `description:"Enable Etcd backend with default settings." json:"etcd,omitempty" toml:"etcd,omitempty" yaml:"etcd,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	//ZooKeeper     *zk.Provider                   `description:"Enable ZooKeeper backend with default settings." json:"zooKeeper,omitempty" toml:"zooKeeper,omitempty" yaml:"zooKeeper,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	//Redis         *redis.Provider                `description:"Enable Redis backend with default settings." json:"redis,omitempty" toml:"redis,omitempty" yaml:"redis,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
	//HTTP          *http.Provider                 `description:"Enable HTTP backend with default settings." json:"http,omitempty" toml:"http,omitempty" yaml:"http,omitempty" label:"allowEmpty" file:"allowEmpty" export:"true"`
}

// SetEffectiveConfiguration adds missing configuration parameters derived from existing ones.
// It also takes care of maintaining backwards compatibility.
func (c *Configuration) SetEffectiveConfiguration() {
	// Creates the default entry point if needed
	if !c.hasUserDefinedEntrypoint() {
		ep := &EntryPoint{Address: ":80"}
		ep.SetDefaults()
		// TODO: double check this tomorrow
		if c.EntryPoints == nil {
			c.EntryPoints = make(EntryPoints)
		}
		c.EntryPoints["http"] = ep
	}

	// Creates the internal gateway entry point if needed
	//if c.API != nil && c.API.Insecure {
	//
	//	if _, ok := c.EntryPoints[DefaultInternalEntryPointName]; !ok {
	//		ep := &EntryPoint{Address: ":8080"}
	//		ep.SetDefaults()
	//		c.EntryPoints[DefaultInternalEntryPointName] = ep
	//	}
	//}

}

func (c *Configuration) hasUserDefinedEntrypoint() bool {
	return len(c.EntryPoints) != 0
}

// ValidateConfiguration validate that configuration is coherent.
func (c *Configuration) ValidateConfiguration() error {

	if c.Core != nil {
		switch c.Core.DefaultRuleSyntax {
		case "v3": // NOOP
		case "v2":
			// TODO: point to migration guide.
			log.Log.Warn("v2 rules syntax is now deprecated, please use v3 instead...")
		default:
			return fmt.Errorf("unsupported default rule syntax configuration: %q", c.Core.DefaultRuleSyntax)
		}
	}

	return nil
}
