package dynamic

import (
	"fmt"
	"go-faster-gateway/pkg/ip"
)

// ForwardAuthDefaultMaxBodySize is the ForwardAuth.MaxBodySize option default value.
const ForwardAuthDefaultMaxBodySize int64 = -1

// Middleware holds the Middleware configuration.
type Middleware struct {
	AddPrefix        *AddPrefix        `json:"addPrefix,omitempty" toml:"addPrefix,omitempty" yaml:"addPrefix,omitempty" export:"true"`
	StripPrefix      *StripPrefix      `json:"stripPrefix,omitempty" toml:"stripPrefix,omitempty" yaml:"stripPrefix,omitempty" export:"true"`
	StripPrefixRegex *StripPrefixRegex `json:"stripPrefixRegex,omitempty" toml:"stripPrefixRegex,omitempty" yaml:"stripPrefixRegex,omitempty" export:"true"`

	Chain *Chain `json:"chain,omitempty" toml:"chain,omitempty" yaml:"chain,omitempty" export:"true"`
	// Deprecated: please use IPAllowList instead.
	IPWhiteList *IPWhiteList `json:"ipWhiteList,omitempty" toml:"ipWhiteList,omitempty" yaml:"ipWhiteList,omitempty" export:"true"`
	IPAllowList *IPAllowList `json:"ipAllowList,omitempty" toml:"ipAllowList,omitempty" yaml:"ipAllowList,omitempty" export:"true"`
	BasicAuth   *BasicAuth   `json:"basicAuth,omitempty" toml:"basicAuth,omitempty" yaml:"basicAuth,omitempty" export:"true"`
	Buffering   *Buffering   `json:"buffering,omitempty" toml:"buffering,omitempty" yaml:"buffering,omitempty" export:"true"`
	// Gateway API filter middlewares.
	RequestHeaderModifier  *HeaderModifier `json:"requestHeaderModifier,omitempty" toml:"-" yaml:"-" label:"-" file:"-" kv:"-" export:"true"`
	ResponseHeaderModifier *HeaderModifier `json:"responseHeaderModifier,omitempty" toml:"-" yaml:"-" label:"-" file:"-" kv:"-" export:"true"`
	URLRewrite             *URLRewrite     `json:"URLRewrite,omitempty" toml:"-" yaml:"-" label:"-" file:"-" kv:"-" export:"true"`

	Headers  *Headers  `json:"headers,omitempty" toml:"headers,omitempty" yaml:"headers,omitempty" export:"true"`
	Compress *Compress `json:"compress,omitempty" toml:"compress,omitempty" yaml:"compress,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// AddPrefix holds the add prefix middleware configuration.
// This middleware updates the path of a request before forwarding it.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/addprefix/
type AddPrefix struct {
	// Prefix is the string to add before the current path in the requested URL.
	// It should include a leading slash (/).
	Prefix string `json:"prefix,omitempty" toml:"prefix,omitempty" yaml:"prefix,omitempty" export:"true"`
}

// BasicAuth holds the basic auth middleware configuration.
// This middleware restricts access to your services to known users.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/basicauth/
type BasicAuth struct {
	// Users is an array of authorized users.
	// Each user must be declared using the name:hashed-password format.
	// Tip: Use htpasswd to generate the passwords.
	Users Users `json:"users,omitempty" toml:"users,omitempty" yaml:"users,omitempty" loggable:"false"`
	// UsersFile is the path to an external file that contains the authorized users.
	UsersFile string `json:"usersFile,omitempty" toml:"usersFile,omitempty" yaml:"usersFile,omitempty"`
	// Realm allows the protected resources on a server to be partitioned into a set of protection spaces, each with its own authentication scheme.
	// Default: traefik.
	Realm string `json:"realm,omitempty" toml:"realm,omitempty" yaml:"realm,omitempty"`
	// RemoveHeader sets the removeHeader option to true to remove the authorization header before forwarding the request to your service.
	// Default: false.
	RemoveHeader bool `json:"removeHeader,omitempty" toml:"removeHeader,omitempty" yaml:"removeHeader,omitempty" export:"true"`
	// HeaderField defines a header field to store the authenticated user.
	// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/basicauth/#headerfield
	HeaderField string `json:"headerField,omitempty" toml:"headerField,omitempty" yaml:"headerField,omitempty" export:"true"`
}

// Buffering holds the buffering middleware configuration.
// This middleware retries or limits the size of requests that can be forwarded to backends.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/buffering/#maxrequestbodybytes
type Buffering struct {
	// MaxRequestBodyBytes defines the maximum allowed body size for the request (in bytes).
	// If the request exceeds the allowed size, it is not forwarded to the service, and the client gets a 413 (Request Entity Too Large) response.
	// Default: 0 (no maximum).
	MaxRequestBodyBytes int64 `json:"maxRequestBodyBytes,omitempty" toml:"maxRequestBodyBytes,omitempty" yaml:"maxRequestBodyBytes,omitempty" export:"true"`
	// MemRequestBodyBytes defines the threshold (in bytes) from which the request will be buffered on disk instead of in memory.
	// Default: 1048576 (1Mi).
	MemRequestBodyBytes int64 `json:"memRequestBodyBytes,omitempty" toml:"memRequestBodyBytes,omitempty" yaml:"memRequestBodyBytes,omitempty" export:"true"`
	// MaxResponseBodyBytes defines the maximum allowed response size from the service (in bytes).
	// If the response exceeds the allowed size, it is not forwarded to the client. The client gets a 500 (Internal Server Error) response instead.
	// Default: 0 (no maximum).
	MaxResponseBodyBytes int64 `json:"maxResponseBodyBytes,omitempty" toml:"maxResponseBodyBytes,omitempty" yaml:"maxResponseBodyBytes,omitempty" export:"true"`
	// MemResponseBodyBytes defines the threshold (in bytes) from which the response will be buffered on disk instead of in memory.
	// Default: 1048576 (1Mi).
	MemResponseBodyBytes int64 `json:"memResponseBodyBytes,omitempty" toml:"memResponseBodyBytes,omitempty" yaml:"memResponseBodyBytes,omitempty" export:"true"`
	// RetryExpression defines the retry conditions.
	// It is a logical combination of functions with operators AND (&&) and OR (||).
	// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/buffering/#retryexpression
	RetryExpression string `json:"retryExpression,omitempty" toml:"retryExpression,omitempty" yaml:"retryExpression,omitempty" export:"true"`
}

// Chain holds the chain middleware configuration.
// This middleware enables to define reusable combinations of other pieces of middleware.
type Chain struct {
	// Middlewares is the list of middleware names which composes the chain.
	Middlewares []string `json:"middlewares,omitempty" toml:"middlewares,omitempty" yaml:"middlewares,omitempty" export:"true"`
}

// IPStrategy holds the IP strategy configuration used by Traefik to determine the client IP.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/ipallowlist/#ipstrategy
type IPStrategy struct {
	// Depth tells Traefik to use the X-Forwarded-For header and take the IP located at the depth position (starting from the right).
	Depth int `json:"depth,omitempty" toml:"depth,omitempty" yaml:"depth,omitempty" export:"true"`
	// ExcludedIPs configures Traefik to scan the X-Forwarded-For header and select the first IP not in the list.
	ExcludedIPs []string `json:"excludedIPs,omitempty" toml:"excludedIPs,omitempty" yaml:"excludedIPs,omitempty"`
	// IPv6Subnet configures Traefik to consider all IPv6 addresses from the defined subnet as originating from the same IP. Applies to RemoteAddrStrategy and DepthStrategy.
	IPv6Subnet *int `json:"ipv6Subnet,omitempty" toml:"ipv6Subnet,omitempty" yaml:"ipv6Subnet,omitempty"`
	// TODO(mpl): I think we should make RemoteAddr an explicit field. For one thing, it would yield better documentation.
}

// Get an IP selection strategy.
// If nil return the RemoteAddr strategy
// else return a strategy based on the configuration using the X-Forwarded-For Header.
// Depth override the ExcludedIPs.
func (s *IPStrategy) Get() (ip.Strategy, error) {
	if s == nil {
		return &ip.RemoteAddrStrategy{}, nil
	}

	if s.Depth > 0 {
		if s.IPv6Subnet != nil && (*s.IPv6Subnet <= 0 || *s.IPv6Subnet > 128) {
			return nil, fmt.Errorf("invalid IPv6 subnet %d value, should be greater to 0 and lower or equal to 128", *s.IPv6Subnet)
		}

		return &ip.DepthStrategy{
			Depth:      s.Depth,
			IPv6Subnet: s.IPv6Subnet,
		}, nil
	}

	if len(s.ExcludedIPs) > 0 {
		checker, err := ip.NewChecker(s.ExcludedIPs)
		if err != nil {
			return nil, err
		}
		return &ip.PoolStrategy{
			Checker: checker,
		}, nil
	}

	if s.IPv6Subnet != nil && (*s.IPv6Subnet <= 0 || *s.IPv6Subnet > 128) {
		return nil, fmt.Errorf("invalid IPv6 subnet %d value, should be greater to 0 and lower or equal to 128", *s.IPv6Subnet)
	}

	return &ip.RemoteAddrStrategy{
		IPv6Subnet: s.IPv6Subnet,
	}, nil
}

// +k8s:deepcopy-gen=true

// IPWhiteList holds the IP whitelist middleware configuration.
// This middleware limits allowed requests based on the client IP.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/ipwhitelist/
// Deprecated: please use IPAllowList instead.
type IPWhiteList struct {
	// SourceRange defines the set of allowed IPs (or ranges of allowed IPs by using CIDR notation). Required.
	SourceRange []string    `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
	IPStrategy  *IPStrategy `json:"ipStrategy,omitempty" toml:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// +k8s:deepcopy-gen=true

// IPAllowList holds the IP allowlist middleware configuration.
// This middleware limits allowed requests based on the client IP.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/ipallowlist/
type IPAllowList struct {
	// SourceRange defines the set of allowed IPs (or ranges of allowed IPs by using CIDR notation).
	SourceRange []string    `json:"sourceRange,omitempty" toml:"sourceRange,omitempty" yaml:"sourceRange,omitempty"`
	IPStrategy  *IPStrategy `json:"ipStrategy,omitempty" toml:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
	// RejectStatusCode defines the EasyServiceRoute status code used for refused requests.
	// If not set, the default is 403 (Forbidden).
	RejectStatusCode int `json:"rejectStatusCode,omitempty" toml:"rejectStatusCode,omitempty" yaml:"rejectStatusCode,omitempty" label:"allowEmpty" file:"allowEmpty" kv:"allowEmpty" export:"true"`
}

// SourceCriterion defines what criterion is used to group requests as originating from a common source.
// If none are set, the default is to use the request's remote address field.
// All fields are mutually exclusive.
type SourceCriterion struct {
	IPStrategy *IPStrategy `json:"ipStrategy,omitempty" toml:"ipStrategy,omitempty" yaml:"ipStrategy,omitempty" export:"true"`
	// RequestHeaderName defines the name of the header used to group incoming requests.
	RequestHeaderName string `json:"requestHeaderName,omitempty" toml:"requestHeaderName,omitempty" yaml:"requestHeaderName,omitempty" export:"true"`
	// RequestHost defines whether to consider the request Host as the source.
	RequestHost bool `json:"requestHost,omitempty" toml:"requestHost,omitempty" yaml:"requestHost,omitempty" export:"true"`
}

// Compress holds the compress middleware configuration.
// This middleware compresses responses before sending them to the client, using gzip, brotli, or zstd compression.
type Compress struct {
	// ExcludedContentTypes defines the list of content types to compare the Content-Type header of the incoming requests and responses before compressing.
	// `application/grpc` is always excluded.
	ExcludedContentTypes []string `json:"excludedContentTypes,omitempty" toml:"excludedContentTypes,omitempty" yaml:"excludedContentTypes,omitempty" export:"true"`
	// IncludedContentTypes defines the list of content types to compare the Content-Type header of the responses before compressing.
	IncludedContentTypes []string `json:"includedContentTypes,omitempty" toml:"includedContentTypes,omitempty" yaml:"includedContentTypes,omitempty" export:"true"`
	// MinResponseBodyBytes defines the minimum amount of bytes a response body must have to be compressed.
	// Default: 1024.
	MinResponseBodyBytes int `json:"minResponseBodyBytes,omitempty" toml:"minResponseBodyBytes,omitempty" yaml:"minResponseBodyBytes,omitempty" export:"true"`
	// Encodings defines the list of supported compression algorithms.
	Encodings []string `json:"encodings,omitempty" toml:"encodings,omitempty" yaml:"encodings,omitempty" export:"true"`
	// DefaultEncoding specifies the default encoding if the `Accept-Encoding` header is not in the request or contains a wildcard (`*`).
	DefaultEncoding string `json:"defaultEncoding,omitempty" toml:"defaultEncoding,omitempty" yaml:"defaultEncoding,omitempty" export:"true"`
}

func (c *Compress) SetDefaults() {
	c.Encodings = []string{"zstd", "br", "gzip"}
}

// StripPrefix holds the strip prefix middleware configuration.
// This middleware removes the specified prefixes from the URL path.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/stripprefix/
type StripPrefix struct {
	// Prefixes defines the prefixes to strip from the request URL.
	Prefixes []string `json:"prefixes,omitempty" toml:"prefixes,omitempty" yaml:"prefixes,omitempty" export:"true"`

	// Deprecated: ForceSlash option is deprecated, please remove any usage of this option.
	// ForceSlash ensures that the resulting stripped path is not the empty string, by replacing it with / when necessary.
	// Default: true.
	ForceSlash *bool `json:"forceSlash,omitempty" toml:"forceSlash,omitempty" yaml:"forceSlash,omitempty" export:"true"`
}

// +k8s:deepcopy-gen=true

// StripPrefixRegex holds the strip prefix regex middleware configuration.
// This middleware removes the matching prefixes from the URL path.
// More info: https://doc.traefik.io/traefik/v3.3/middlewares/http/stripprefixregex/
type StripPrefixRegex struct {
	// Regex defines the regular expression to match the path prefix from the request URL.
	Regex []string `json:"regex,omitempty" toml:"regex,omitempty" yaml:"regex,omitempty" export:"true"`
}

// Users holds a list of users.
type Users []string

// +k8s:deepcopy-gen=true

// HeaderModifier holds the request/response header modifier configuration.
type HeaderModifier struct {
	Set    map[string]string `json:"set,omitempty"`
	Add    map[string]string `json:"add,omitempty"`
	Remove []string          `json:"remove,omitempty"`
}

// URLRewrite holds the URL rewrite middleware configuration.
type URLRewrite struct {
	Hostname   *string `json:"hostname,omitempty"`
	Path       *string `json:"path,omitempty"`
	PathPrefix *string `json:"pathPrefix,omitempty"`
}

// Headers holds the headers middleware configuration.
// This middleware manages the requests and responses headers.
type Headers struct {
	// CustomRequestHeaders defines the header names and values to apply to the request.
	CustomRequestHeaders map[string]string `json:"customRequestHeaders,omitempty" toml:"customRequestHeaders,omitempty" yaml:"customRequestHeaders,omitempty" export:"true"`
	// CustomResponseHeaders defines the header names and values to apply to the response.
	CustomResponseHeaders map[string]string `json:"customResponseHeaders,omitempty" toml:"customResponseHeaders,omitempty" yaml:"customResponseHeaders,omitempty" export:"true"`

	// AccessControlAllowCredentials defines whether the request can include user credentials.
	AccessControlAllowCredentials bool `json:"accessControlAllowCredentials,omitempty" toml:"accessControlAllowCredentials,omitempty" yaml:"accessControlAllowCredentials,omitempty" export:"true"`
	// AccessControlAllowHeaders defines the Access-Control-Request-Headers values sent in preflight response.
	AccessControlAllowHeaders []string `json:"accessControlAllowHeaders,omitempty" toml:"accessControlAllowHeaders,omitempty" yaml:"accessControlAllowHeaders,omitempty" export:"true"`
	// AccessControlAllowMethods defines the Access-Control-Request-Method values sent in preflight response.
	AccessControlAllowMethods []string `json:"accessControlAllowMethods,omitempty" toml:"accessControlAllowMethods,omitempty" yaml:"accessControlAllowMethods,omitempty" export:"true"`
	// AccessControlAllowOriginList is a list of allowable origins. Can also be a wildcard origin "*".
	AccessControlAllowOriginList []string `json:"accessControlAllowOriginList,omitempty" toml:"accessControlAllowOriginList,omitempty" yaml:"accessControlAllowOriginList,omitempty"`
	// AccessControlAllowOriginListRegex is a list of allowable origins written following the Regular Expression syntax (https://golang.org/pkg/regexp/).
	AccessControlAllowOriginListRegex []string `json:"accessControlAllowOriginListRegex,omitempty" toml:"accessControlAllowOriginListRegex,omitempty" yaml:"accessControlAllowOriginListRegex,omitempty"`
	// AccessControlExposeHeaders defines the Access-Control-Expose-Headers values sent in preflight response.
	AccessControlExposeHeaders []string `json:"accessControlExposeHeaders,omitempty" toml:"accessControlExposeHeaders,omitempty" yaml:"accessControlExposeHeaders,omitempty" export:"true"`
	// AccessControlMaxAge defines the time that a preflight request may be cached.
	AccessControlMaxAge int64 `json:"accessControlMaxAge,omitempty" toml:"accessControlMaxAge,omitempty" yaml:"accessControlMaxAge,omitempty" export:"true"`
	// AddVaryHeader defines whether the Vary header is automatically added/updated when the AccessControlAllowOriginList is set.
	AddVaryHeader bool `json:"addVaryHeader,omitempty" toml:"addVaryHeader,omitempty" yaml:"addVaryHeader,omitempty" export:"true"`
	// AllowedHosts defines the fully qualified list of allowed domain names.
	AllowedHosts []string `json:"allowedHosts,omitempty" toml:"allowedHosts,omitempty" yaml:"allowedHosts,omitempty"`
	// HostsProxyHeaders defines the header keys that may hold a proxied hostname value for the request.
	HostsProxyHeaders []string `json:"hostsProxyHeaders,omitempty" toml:"hostsProxyHeaders,omitempty" yaml:"hostsProxyHeaders,omitempty" export:"true"`
	// SSLProxyHeaders defines the header keys with associated values that would indicate a valid HTTPS request.
	// It can be useful when using other proxies (example: "X-Forwarded-Proto": "https").
	SSLProxyHeaders map[string]string `json:"sslProxyHeaders,omitempty" toml:"sslProxyHeaders,omitempty" yaml:"sslProxyHeaders,omitempty"`
	// STSSeconds defines the max-age of the Strict-Transport-Security header.
	// If set to 0, the header is not set.
	STSSeconds int64 `json:"stsSeconds,omitempty" toml:"stsSeconds,omitempty" yaml:"stsSeconds,omitempty" export:"true"`
	// STSIncludeSubdomains defines whether the includeSubDomains directive is appended to the Strict-Transport-Security header.
	STSIncludeSubdomains bool `json:"stsIncludeSubdomains,omitempty" toml:"stsIncludeSubdomains,omitempty" yaml:"stsIncludeSubdomains,omitempty" export:"true"`
	// STSPreload defines whether the preload flag is appended to the Strict-Transport-Security header.
	STSPreload bool `json:"stsPreload,omitempty" toml:"stsPreload,omitempty" yaml:"stsPreload,omitempty" export:"true"`
	// ForceSTSHeader defines whether to add the STS header even when the connection is EasyServiceRoute.
	ForceSTSHeader bool `json:"forceSTSHeader,omitempty" toml:"forceSTSHeader,omitempty" yaml:"forceSTSHeader,omitempty" export:"true"`
	// FrameDeny defines whether to add the X-Frame-Options header with the DENY value.
	FrameDeny bool `json:"frameDeny,omitempty" toml:"frameDeny,omitempty" yaml:"frameDeny,omitempty" export:"true"`
	// CustomFrameOptionsValue defines the X-Frame-Options header value.
	// This overrides the FrameDeny option.
	CustomFrameOptionsValue string `json:"customFrameOptionsValue,omitempty" toml:"customFrameOptionsValue,omitempty" yaml:"customFrameOptionsValue,omitempty"`
	// ContentTypeNosniff defines whether to add the X-Content-Type-Options header with the nosniff value.
	ContentTypeNosniff bool `json:"contentTypeNosniff,omitempty" toml:"contentTypeNosniff,omitempty" yaml:"contentTypeNosniff,omitempty" export:"true"`
	// BrowserXSSFilter defines whether to add the X-XSS-Protection header with the value 1; mode=block.
	BrowserXSSFilter bool `json:"browserXssFilter,omitempty" toml:"browserXssFilter,omitempty" yaml:"browserXssFilter,omitempty" export:"true"`
	// CustomBrowserXSSValue defines the X-XSS-Protection header value.
	// This overrides the BrowserXssFilter option.
	CustomBrowserXSSValue string `json:"customBrowserXSSValue,omitempty" toml:"customBrowserXSSValue,omitempty" yaml:"customBrowserXSSValue,omitempty"`
	// ContentSecurityPolicy defines the Content-Security-Policy header value.
	ContentSecurityPolicy string `json:"contentSecurityPolicy,omitempty" toml:"contentSecurityPolicy,omitempty" yaml:"contentSecurityPolicy,omitempty"`
	// ContentSecurityPolicyReportOnly defines the Content-Security-Policy-Report-Only header value.
	ContentSecurityPolicyReportOnly string `json:"contentSecurityPolicyReportOnly,omitempty" toml:"contentSecurityPolicyReportOnly,omitempty" yaml:"contentSecurityPolicyReportOnly,omitempty"`
	// PublicKey is the public key that implements HPKP to prevent MITM attacks with forged certificates.
	PublicKey string `json:"publicKey,omitempty" toml:"publicKey,omitempty" yaml:"publicKey,omitempty"`
	// ReferrerPolicy defines the Referrer-Policy header value.
	// This allows sites to control whether browsers forward the Referer header to other sites.
	ReferrerPolicy string `json:"referrerPolicy,omitempty" toml:"referrerPolicy,omitempty" yaml:"referrerPolicy,omitempty" export:"true"`
	// PermissionsPolicy defines the Permissions-Policy header value.
	// This allows sites to control browser features.
	PermissionsPolicy string `json:"permissionsPolicy,omitempty" toml:"permissionsPolicy,omitempty" yaml:"permissionsPolicy,omitempty" export:"true"`
	// IsDevelopment defines whether to mitigate the unwanted effects of the AllowedHosts, SSL, and STS options when developing.
	// Usually testing takes place using EasyServiceRoute, not HTTPS, and on localhost, not your production domain.
	// If you would like your development environment to mimic production with complete Host blocking, SSL redirects,
	// and STS headers, leave this as false.
	IsDevelopment bool `json:"isDevelopment,omitempty" toml:"isDevelopment,omitempty" yaml:"isDevelopment,omitempty" export:"true"`

	// Deprecated: FeaturePolicy option is deprecated, please use PermissionsPolicy instead.
	FeaturePolicy *string `json:"featurePolicy,omitempty" toml:"featurePolicy,omitempty" yaml:"featurePolicy,omitempty" export:"true"`
	// Deprecated: SSLRedirect option is deprecated, please use EntryPoint redirection or RedirectScheme instead.
	SSLRedirect *bool `json:"sslRedirect,omitempty" toml:"sslRedirect,omitempty" yaml:"sslRedirect,omitempty" export:"true"`
	// Deprecated: SSLTemporaryRedirect option is deprecated, please use EntryPoint redirection or RedirectScheme instead.
	SSLTemporaryRedirect *bool `json:"sslTemporaryRedirect,omitempty" toml:"sslTemporaryRedirect,omitempty" yaml:"sslTemporaryRedirect,omitempty" export:"true"`
	// Deprecated: SSLHost option is deprecated, please use RedirectRegex instead.
	SSLHost *string `json:"sslHost,omitempty" toml:"sslHost,omitempty" yaml:"sslHost,omitempty"`
	// Deprecated: SSLForceHost option is deprecated, please use RedirectRegex instead.
	SSLForceHost *bool `json:"sslForceHost,omitempty" toml:"sslForceHost,omitempty" yaml:"sslForceHost,omitempty" export:"true"`
}

// HasCustomHeadersDefined checks to see if any of the custom header elements have been set.
func (h *Headers) HasCustomHeadersDefined() bool {
	return h != nil && (len(h.CustomResponseHeaders) != 0 ||
		len(h.CustomRequestHeaders) != 0)
}

// HasCorsHeadersDefined checks to see if any of the cors header elements have been set.
func (h *Headers) HasCorsHeadersDefined() bool {
	return h != nil && (h.AccessControlAllowCredentials ||
		len(h.AccessControlAllowHeaders) != 0 ||
		len(h.AccessControlAllowMethods) != 0 ||
		len(h.AccessControlAllowOriginList) != 0 ||
		len(h.AccessControlAllowOriginListRegex) != 0 ||
		len(h.AccessControlExposeHeaders) != 0 ||
		h.AccessControlMaxAge != 0 ||
		h.AddVaryHeader)
}

// HasSecureHeadersDefined checks to see if any of the secure header elements have been set.
func (h *Headers) HasSecureHeadersDefined() bool {
	return h != nil && (len(h.AllowedHosts) != 0 ||
		len(h.HostsProxyHeaders) != 0 ||
		(h.SSLRedirect != nil && *h.SSLRedirect) ||
		(h.SSLTemporaryRedirect != nil && *h.SSLTemporaryRedirect) ||
		(h.SSLForceHost != nil && *h.SSLForceHost) ||
		(h.SSLHost != nil && *h.SSLHost != "") ||
		len(h.SSLProxyHeaders) != 0 ||
		h.STSSeconds != 0 ||
		h.STSIncludeSubdomains ||
		h.STSPreload ||
		h.ForceSTSHeader ||
		h.FrameDeny ||
		h.CustomFrameOptionsValue != "" ||
		h.ContentTypeNosniff ||
		h.BrowserXSSFilter ||
		h.CustomBrowserXSSValue != "" ||
		h.ContentSecurityPolicy != "" ||
		h.ContentSecurityPolicyReportOnly != "" ||
		h.PublicKey != "" ||
		h.ReferrerPolicy != "" ||
		(h.FeaturePolicy != nil && *h.FeaturePolicy != "") ||
		h.PermissionsPolicy != "" ||
		h.IsDevelopment)
}
