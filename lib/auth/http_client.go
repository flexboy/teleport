/*
Copyright 2023 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package auth

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gravitational/roundtrip"
	"github.com/gravitational/trace"
	"github.com/jonboulle/clockwork"

	"github.com/gravitational/teleport"
	"github.com/gravitational/teleport/api/breaker"
	"github.com/gravitational/teleport/api/client"
	"github.com/gravitational/teleport/api/client/proto"
	"github.com/gravitational/teleport/api/constants"
	apidefaults "github.com/gravitational/teleport/api/defaults"
	tracehttp "github.com/gravitational/teleport/api/observability/tracing/http"
	"github.com/gravitational/teleport/api/types"
	"github.com/gravitational/teleport/lib/defaults"
	"github.com/gravitational/teleport/lib/events"
	"github.com/gravitational/teleport/lib/httplib"
	"github.com/gravitational/teleport/lib/services"
	"github.com/gravitational/teleport/lib/session"
	"github.com/gravitational/teleport/lib/utils"
)

// HTTPClientConfig contains configuration for an HTTP client.
type HTTPClientConfig struct {
	// TLS holds the TLS config for the http client.
	TLS *tls.Config
	// MaxIdleConns controls the maximum number of idle (keep-alive) connections across all hosts.
	MaxIdleConns int
	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle (keep-alive) connections to keep per-host.
	MaxIdleConnsPerHost int
	// MaxConnsPerHost limits the total number of connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	MaxConnsPerHost int
	// RequestTimeout specifies a time limit for requests made by this Client.
	RequestTimeout time.Duration
	// IdleConnTimeout defines the maximum amount of time before idle connections are closed.
	IdleConnTimeout time.Duration
	// ResponseHeaderTimeout specifies the amount of time to wait for a server's
	// response headers after fully writing the request (including its body, if any).
	// This time does not include the time to read the response body.
	ResponseHeaderTimeout time.Duration
	// Dialer is a custom dialer used to dial a server. The Dialer should
	// have custom logic to provide an address to the dialer. If set, Dialer
	// takes precedence over all other connection options.
	Dialer client.ContextDialer
	// ALPNSNIAuthDialClusterName if present the client will include ALPN SNI routing information in TLS Hello message
	// allowing to dial auth service through Teleport Proxy directly without using SSH Tunnels.
	ALPNSNIAuthDialClusterName string
	// CircuitBreakerConfig defines how the circuit breaker should behave.
	CircuitBreakerConfig breaker.Config
}

// CheckAndSetDefaults validates and sets defaults for HTTP configuration.
func (c *HTTPClientConfig) CheckAndSetDefaults() error {
	if c.TLS == nil {
		return trace.BadParameter("missing TLS config")
	}

	if c.Dialer == nil {
		return trace.BadParameter("missing dialer")
	}

	// Set the next protocol. This is needed due to the Auth Server using a
	// multiplexer for protocol detection. Unless next protocol is specified
	// it will attempt to upgrade to HTTP2 and at that point there is no way
	// to distinguish between HTTP2/JSON or gRPC.
	c.TLS.NextProtos = []string{teleport.HTTPNextProtoTLS}

	// Configure ALPN SNI direct dial TLS routing information used by ALPN SNI proxy in order to
	// dial auth service without using SSH tunnels.
	c.TLS = client.ConfigureALPN(c.TLS, c.ALPNSNIAuthDialClusterName)

	if c.CircuitBreakerConfig.Trip == nil || c.CircuitBreakerConfig.IsSuccessful == nil {
		c.CircuitBreakerConfig = breaker.DefaultBreakerConfig(clockwork.NewRealClock())
	}

	// One or both of these timeouts should be set to ensure there is a timeout in place.
	if c.RequestTimeout == 0 && c.ResponseHeaderTimeout == 0 {
		c.RequestTimeout = defaults.HTTPRequestTimeout
		c.ResponseHeaderTimeout = apidefaults.DefaultIOTimeout
	}

	// Leaving this unset will lead to connections open forever and will cause memory leaks in a long running process.
	if c.IdleConnTimeout == 0 {
		c.IdleConnTimeout = defaults.HTTPIdleTimeout
	}

	// Increase the size of the connection pool. This substantially improves the
	// performance of Teleport under load as it reduces the number of TLS
	// handshakes performed.
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = defaults.HTTPMaxIdleConns
	}
	if c.MaxIdleConnsPerHost == 0 {
		c.MaxIdleConnsPerHost = defaults.HTTPMaxIdleConnsPerHost
	}

	// Limit the total number of connections to the Auth Server. Some hosts allow a low
	// number of connections per process (ulimit) to a host. This is a problem for
	// enhanced session recording auditing which emits so many events to the
	// Audit Log (using the Auth Client) that the connection pool often does not
	// have a free connection to return, so just opens a new one. This quickly
	// leads to hitting the OS limit and the client returning out of file
	// descriptors error.
	if c.MaxConnsPerHost == 0 {
		c.MaxConnsPerHost = defaults.HTTPMaxConnsPerHost
	}

	return nil
}

// Clone creates a new client with the same configuration.
func (c *HTTPClientConfig) Clone() *HTTPClientConfig {
	return &HTTPClientConfig{
		TLS:                        c.TLS.Clone(),
		MaxIdleConns:               c.MaxIdleConns,
		MaxIdleConnsPerHost:        c.MaxIdleConnsPerHost,
		MaxConnsPerHost:            c.MaxConnsPerHost,
		RequestTimeout:             c.RequestTimeout,
		IdleConnTimeout:            c.IdleConnTimeout,
		ResponseHeaderTimeout:      c.ResponseHeaderTimeout,
		Dialer:                     c.Dialer,
		ALPNSNIAuthDialClusterName: c.ALPNSNIAuthDialClusterName,
		CircuitBreakerConfig:       c.CircuitBreakerConfig,
	}
}

// HTTPClient is a teleport HTTP API client.
type HTTPClient struct {
	*roundtrip.Client
	// cfg is the http client configuration.
	cfg *HTTPClientConfig
}

// NewHTTPClient creates a new HTTP client with TLS authentication and the given dialer.
func NewHTTPClient(cfg *HTTPClientConfig, params ...roundtrip.ClientParam) (*HTTPClient, error) {
	if err := cfg.CheckAndSetDefaults(); err != nil {
		return nil, err
	}

	transport := &http.Transport{
		DialContext:           cfg.Dialer.DialContext,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		TLSClientConfig:       cfg.TLS,
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		MaxConnsPerHost:       cfg.MaxConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
	}

	roundtripClient, err := newRoundtripClient(cfg, transport)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &HTTPClient{
		cfg:    cfg,
		Client: roundtripClient,
	}, nil
}

func newRoundtripClient(cfg *HTTPClientConfig, transport *http.Transport, params ...roundtrip.ClientParam) (*roundtrip.Client, error) {
	cb, err := breaker.New(cfg.CircuitBreakerConfig)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	clientParams := append(
		[]roundtrip.ClientParam{
			roundtrip.HTTPClient(&http.Client{
				Timeout:   cfg.RequestTimeout,
				Transport: tracehttp.NewTransport(breaker.NewRoundTripper(cb, transport)),
			}),
			roundtrip.SanitizerEnabled(true),
		},
		params...,
	)

	// Since the client uses a custom dialer and SNI is used for TLS handshake, the address
	// used here is arbitrary as it just needs to be set to pass http request validation.
	roundtripClient, err := roundtrip.NewClient("https://"+constants.APIDomain, CurrentVersion, clientParams...)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return roundtripClient, nil
}

// CloneHTTPClient creates a new HTTP client with the same configuration.
func (c *HTTPClient) CloneHTTPClient(params ...roundtrip.ClientParam) (*HTTPClient, error) {
	cfg := c.cfg.Clone()

	// We copy the transport which may have had roundtrip.ClientParams applied on initial creation.
	transport, err := c.getTransport()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	roundtripClient, err := newRoundtripClient(c.cfg, transport, params...)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return &HTTPClient{
		Client: roundtripClient,
		cfg:    cfg,
	}, nil
}

// ClientParamRequestTimeout sets request timeout of the HTTP transport used by the client.
func ClientParamTimeout(timeout time.Duration) roundtrip.ClientParam {
	return func(c *roundtrip.Client) error {
		c.HTTPClient().Timeout = timeout
		return nil
	}
}

// ClientParamResponseHeaderTimeout sets response header timeout of the HTTP transport used by the client.
func ClientParamResponseHeaderTimeout(timeout time.Duration) roundtrip.ClientParam {
	return func(c *roundtrip.Client) error {
		if t, err := getHTTPTransport(c); err == nil {
			t.ResponseHeaderTimeout = timeout
		}
		return nil
	}
}

// ClientParamIdleConnTimeout sets idle connection header timeout of the HTTP transport used by the client.
func ClientParamIdleConnTimeout(timeout time.Duration) roundtrip.ClientParam {
	return func(c *roundtrip.Client) error {
		if t, err := getHTTPTransport(c); err == nil {
			t.IdleConnTimeout = timeout
		}
		return nil
	}
}

// Close closes the HTTP client connection to the auth server.
func (c *HTTPClient) Close() {
	c.Client.HTTPClient().CloseIdleConnections()
}

// TLSConfig returns the HTTP client's TLS config.
func (c *HTTPClient) TLSConfig() *tls.Config {
	return c.cfg.TLS
}

// GetTransport returns the HTTP client's transport.
func (c *HTTPClient) getTransport() (*http.Transport, error) {
	return getHTTPTransport(c.Client)
}

func getHTTPTransport(c *roundtrip.Client) (*http.Transport, error) {
	type wrapper interface {
		Unwrap() http.RoundTripper
	}

	transport := c.HTTPClient().Transport
	for {
		switch t := transport.(type) {
		case wrapper:
			transport = t.Unwrap()
		case *http.Transport:
			return t, nil
		default:
			return nil, trace.BadParameter("unexpected transport type %T", t)
		}
	}
}

// PostJSON is a generic method that issues http POST request to the server
func (c *HTTPClient) PostJSON(ctx context.Context, endpoint string, val interface{}) (*roundtrip.Response, error) {
	return httplib.ConvertResponse(c.Client.PostJSON(ctx, endpoint, val))
}

// PutJSON is a generic method that issues http PUT request to the server
func (c *HTTPClient) PutJSON(ctx context.Context, endpoint string, val interface{}) (*roundtrip.Response, error) {
	return httplib.ConvertResponse(c.Client.PutJSON(ctx, endpoint, val))
}

// PostForm is a generic method that issues http POST request to the server
func (c *HTTPClient) PostForm(ctx context.Context, endpoint string, vals url.Values, files ...roundtrip.File) (*roundtrip.Response, error) {
	return httplib.ConvertResponse(c.Client.PostForm(ctx, endpoint, vals, files...))
}

// Get issues http GET request to the server
func (c *HTTPClient) Get(ctx context.Context, u string, params url.Values) (*roundtrip.Response, error) {
	return httplib.ConvertResponse(c.Client.Get(ctx, u, params))
}

// Delete issues http Delete Request to the server
func (c *HTTPClient) Delete(ctx context.Context, u string) (*roundtrip.Response, error) {
	return httplib.ConvertResponse(c.Client.Delete(ctx, u))
}

// ProcessKubeCSR processes CSR request against Kubernetes CA, returns
// signed certificate if successful.
func (c *HTTPClient) ProcessKubeCSR(req KubeCSR) (*KubeCSRResponse, error) {
	if err := req.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	out, err := c.PostJSON(context.TODO(), c.Endpoint("kube", "csr"), req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var re KubeCSRResponse
	if err := json.Unmarshal(out.Bytes(), &re); err != nil {
		return nil, trace.Wrap(err)
	}
	return &re, nil
}

// RotateCertAuthority starts or restarts certificate authority rotation process.
func (c *HTTPClient) RotateCertAuthority(ctx context.Context, req RotateRequest) error {
	_, err := c.PostJSON(ctx, c.Endpoint("authorities", string(req.Type), "rotate"), req)
	return trace.Wrap(err)
}

// RotateExternalCertAuthority rotates external certificate authority,
// this method is used to update only public keys and certificates of the
// the certificate authorities of trusted clusters.
func (c *HTTPClient) RotateExternalCertAuthority(ctx context.Context, ca types.CertAuthority) error {
	if err := services.ValidateCertAuthority(ca); err != nil {
		return trace.Wrap(err)
	}
	data, err := services.MarshalCertAuthority(ca)
	if err != nil {
		return trace.Wrap(err)
	}
	_, err = c.PostJSON(ctx, c.Endpoint("authorities", string(ca.GetType()), "rotate", "external"),
		&rotateExternalCertAuthorityRawReq{CA: data})
	return trace.Wrap(err)
}

// UpsertCertAuthority updates or inserts new cert authority
func (c *HTTPClient) UpsertCertAuthority(ca types.CertAuthority) error {
	if err := services.ValidateCertAuthority(ca); err != nil {
		return trace.Wrap(err)
	}

	data, err := services.MarshalCertAuthority(ca)
	if err != nil {
		return trace.Wrap(err)
	}
	_, err = c.PostJSON(context.TODO(), c.Endpoint("authorities", string(ca.GetType())),
		&upsertCertAuthorityRawReq{CA: data})
	return trace.Wrap(err)
}

// GetCertAuthorities returns a list of certificate authorities
func (c *HTTPClient) GetCertAuthorities(ctx context.Context, caType types.CertAuthType, loadKeys bool, opts ...services.MarshalOption) ([]types.CertAuthority, error) {
	if err := caType.Check(); err != nil {
		return nil, trace.Wrap(err)
	}

	resp, err := c.Get(ctx, c.Endpoint("authorities", string(caType)), url.Values{
		"load_keys": []string{fmt.Sprintf("%t", loadKeys)},
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(resp.Bytes(), &items); err != nil {
		return nil, err
	}
	cas := make([]types.CertAuthority, 0, len(items))
	for _, raw := range items {
		ca, err := services.UnmarshalCertAuthority(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}

		cas = append(cas, ca)
	}

	return cas, nil
}

// GetCertAuthority returns certificate authority by given id. Parameter loadSigningKeys
// controls if signing keys are loaded
func (c *HTTPClient) GetCertAuthority(ctx context.Context, id types.CertAuthID, loadSigningKeys bool, opts ...services.MarshalOption) (types.CertAuthority, error) {
	if err := id.Check(); err != nil {
		return nil, trace.Wrap(err)
	}

	out, err := c.Get(ctx, c.Endpoint("authorities", string(id.Type), id.DomainName), url.Values{
		"load_keys": []string{fmt.Sprintf("%t", loadSigningKeys)},
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	ca, err := services.UnmarshalCertAuthority(out.Bytes(), opts...)
	return ca, trace.Wrap(err)
}

// DeleteCertAuthority deletes cert authority by ID
func (c *HTTPClient) DeleteCertAuthority(id types.CertAuthID) error {
	if err := id.Check(); err != nil {
		return trace.Wrap(err)
	}
	_, err := c.Delete(context.TODO(), c.Endpoint("authorities", string(id.Type), id.DomainName))
	return trace.Wrap(err)
}

// RegisterUsingToken calls the auth service API to register a new node using a registration token
// which was previously issued via CreateToken/UpsertToken.
func (c *HTTPClient) RegisterUsingToken(ctx context.Context, req *types.RegisterUsingTokenRequest) (*proto.Certs, error) {
	if err := req.CheckAndSetDefaults(); err != nil {
		return nil, trace.Wrap(err)
	}
	out, err := c.PostJSON(ctx, c.Endpoint("tokens", "register"), req)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var certs proto.Certs
	if err := json.Unmarshal(out.Bytes(), &certs); err != nil {
		return nil, trace.Wrap(err)
	}

	return &certs, nil
}

// UpsertReverseTunnel is used by admins to create a new reverse tunnel
// to the remote proxy to bypass firewall restrictions
func (c *HTTPClient) UpsertReverseTunnel(tunnel types.ReverseTunnel) error {
	data, err := services.MarshalReverseTunnel(tunnel)
	if err != nil {
		return trace.Wrap(err)
	}
	args := &upsertReverseTunnelRawReq{
		ReverseTunnel: data,
	}
	_, err = c.PostJSON(context.TODO(), c.Endpoint("reversetunnels"), args)
	return trace.Wrap(err)
}

// GetReverseTunnels returns the list of created reverse tunnels
func (c *HTTPClient) GetReverseTunnels(ctx context.Context, opts ...services.MarshalOption) ([]types.ReverseTunnel, error) {
	out, err := c.Get(ctx, c.Endpoint("reversetunnels"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		return nil, trace.Wrap(err)
	}
	tunnels := make([]types.ReverseTunnel, len(items))
	for i, raw := range items {
		tunnel, err := services.UnmarshalReverseTunnel(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		tunnels[i] = tunnel
	}
	return tunnels, nil
}

// DeleteReverseTunnel deletes reverse tunnel by domain name
func (c *HTTPClient) DeleteReverseTunnel(domainName string) error {
	// this is to avoid confusing error in case if domain empty for example
	// HTTP route will fail producing generic not found error
	// instead we catch the error here
	if strings.TrimSpace(domainName) == "" {
		return trace.BadParameter("empty domain name")
	}
	_, err := c.Delete(context.TODO(), c.Endpoint("reversetunnels", domainName))
	return trace.Wrap(err)
}

// UpsertTunnelConnection upserts tunnel connection
func (c *HTTPClient) UpsertTunnelConnection(conn types.TunnelConnection) error {
	data, err := services.MarshalTunnelConnection(conn)
	if err != nil {
		return trace.Wrap(err)
	}
	args := &upsertTunnelConnectionRawReq{
		TunnelConnection: data,
	}
	_, err = c.PostJSON(context.TODO(), c.Endpoint("tunnelconnections"), args)
	return trace.Wrap(err)
}

// GetTunnelConnections returns tunnel connections for a given cluster
func (c *HTTPClient) GetTunnelConnections(clusterName string, opts ...services.MarshalOption) ([]types.TunnelConnection, error) {
	if clusterName == "" {
		return nil, trace.BadParameter("missing cluster name parameter")
	}
	out, err := c.Get(context.TODO(), c.Endpoint("tunnelconnections", clusterName), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		return nil, trace.Wrap(err)
	}
	conns := make([]types.TunnelConnection, len(items))
	for i, raw := range items {
		conn, err := services.UnmarshalTunnelConnection(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		conns[i] = conn
	}
	return conns, nil
}

// GetAllTunnelConnections returns all tunnel connections
func (c *HTTPClient) GetAllTunnelConnections(opts ...services.MarshalOption) ([]types.TunnelConnection, error) {
	out, err := c.Get(context.TODO(), c.Endpoint("tunnelconnections"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		return nil, trace.Wrap(err)
	}
	conns := make([]types.TunnelConnection, len(items))
	for i, raw := range items {
		conn, err := services.UnmarshalTunnelConnection(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		conns[i] = conn
	}
	return conns, nil
}

// DeleteTunnelConnection deletes tunnel connection by name
func (c *HTTPClient) DeleteTunnelConnection(clusterName string, connName string) error {
	if clusterName == "" {
		return trace.BadParameter("missing parameter cluster name")
	}
	if connName == "" {
		return trace.BadParameter("missing parameter connection name")
	}
	_, err := c.Delete(context.TODO(), c.Endpoint("tunnelconnections", clusterName, connName))
	return trace.Wrap(err)
}

// DeleteTunnelConnections deletes all tunnel connections for cluster
func (c *HTTPClient) DeleteTunnelConnections(clusterName string) error {
	if clusterName == "" {
		return trace.BadParameter("missing parameter cluster name")
	}
	_, err := c.Delete(context.TODO(), c.Endpoint("tunnelconnections", clusterName))
	return trace.Wrap(err)
}

// DeleteAllTunnelConnections deletes all tunnel connections
func (c *HTTPClient) DeleteAllTunnelConnections() error {
	_, err := c.Delete(context.TODO(), c.Endpoint("tunnelconnections"))
	return trace.Wrap(err)
}

// GetRemoteClusters returns a list of remote clusters
func (c *HTTPClient) GetRemoteClusters(opts ...services.MarshalOption) ([]types.RemoteCluster, error) {
	out, err := c.Get(context.TODO(), c.Endpoint("remoteclusters"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		return nil, trace.Wrap(err)
	}
	conns := make([]types.RemoteCluster, len(items))
	for i, raw := range items {
		conn, err := services.UnmarshalRemoteCluster(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		conns[i] = conn
	}
	return conns, nil
}

// GetRemoteCluster returns a remote cluster by name
func (c *HTTPClient) GetRemoteCluster(clusterName string) (types.RemoteCluster, error) {
	if clusterName == "" {
		return nil, trace.BadParameter("missing cluster name")
	}
	out, err := c.Get(context.TODO(), c.Endpoint("remoteclusters", clusterName), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return services.UnmarshalRemoteCluster(out.Bytes())
}

// DeleteRemoteCluster deletes remote cluster by name
func (c *HTTPClient) DeleteRemoteCluster(clusterName string) error {
	if clusterName == "" {
		return trace.BadParameter("missing parameter cluster name")
	}
	_, err := c.Delete(context.TODO(), c.Endpoint("remoteclusters", clusterName))
	return trace.Wrap(err)
}

// DeleteAllRemoteClusters deletes all remote clusters
func (c *HTTPClient) DeleteAllRemoteClusters() error {
	_, err := c.Delete(context.TODO(), c.Endpoint("remoteclusters"))
	return trace.Wrap(err)
}

// CreateRemoteCluster creates remote cluster resource
func (c *HTTPClient) CreateRemoteCluster(rc types.RemoteCluster) error {
	data, err := services.MarshalRemoteCluster(rc)
	if err != nil {
		return trace.Wrap(err)
	}
	args := &createRemoteClusterRawReq{
		RemoteCluster: data,
	}
	_, err = c.PostJSON(context.TODO(), c.Endpoint("remoteclusters"), args)
	return trace.Wrap(err)
}

// UpsertAuthServer is used by auth servers to report their presence
// to other auth servers in form of hearbeat expiring after ttl period.
func (c *HTTPClient) UpsertAuthServer(ctx context.Context, s types.Server) error {
	data, err := services.MarshalServer(s)
	if err != nil {
		return trace.Wrap(err)
	}
	args := &upsertServerRawReq{
		Server: data,
	}
	_, err = c.PostJSON(ctx, c.Endpoint("authservers"), args)
	return trace.Wrap(err)
}

// GetAuthServers returns the list of auth servers registered in the cluster.
func (c *HTTPClient) GetAuthServers() ([]types.Server, error) {
	out, err := c.Get(context.TODO(), c.Endpoint("authservers"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		return nil, trace.Wrap(err)
	}
	re := make([]types.Server, len(items))
	for i, raw := range items {
		server, err := services.UnmarshalServer(raw, types.KindAuthServer)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		re[i] = server
	}
	return re, nil
}

// UpsertProxy is used by proxies to report their presence
// to other auth servers in form of heartbeat expiring after ttl period.
func (c *HTTPClient) UpsertProxy(ctx context.Context, s types.Server) error {
	data, err := services.MarshalServer(s)
	if err != nil {
		return trace.Wrap(err)
	}
	args := &upsertServerRawReq{
		Server: data,
	}
	_, err = c.PostJSON(ctx, c.Endpoint("proxies"), args)
	return trace.Wrap(err)
}

// GetProxies returns the list of auth servers registered in the cluster.
func (c *HTTPClient) GetProxies() ([]types.Server, error) {
	out, err := c.Get(context.TODO(), c.Endpoint("proxies"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var items []json.RawMessage
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		return nil, trace.Wrap(err)
	}
	re := make([]types.Server, len(items))
	for i, raw := range items {
		server, err := services.UnmarshalServer(raw, types.KindProxy)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		re[i] = server
	}
	return re, nil
}

// DeleteAllProxies deletes all proxies
func (c *HTTPClient) DeleteAllProxies() error {
	_, err := c.Delete(context.TODO(), c.Endpoint("proxies"))
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// DeleteProxy deletes proxy by name
func (c *HTTPClient) DeleteProxy(ctx context.Context, name string) error {
	if name == "" {
		return trace.BadParameter("missing parameter name")
	}
	_, err := c.Delete(ctx, c.Endpoint("proxies", name))
	if err != nil {
		return trace.Wrap(err)
	}
	return nil
}

// UpsertUser user updates user entry.
func (c *HTTPClient) UpsertUser(user types.User) error {
	data, err := services.MarshalUser(user)
	if err != nil {
		return trace.Wrap(err)
	}
	_, err = c.PostJSON(context.TODO(), c.Endpoint("users"), &upsertUserRawReq{User: data})
	return trace.Wrap(err)
}

// ExtendWebSession creates a new web session for a user based on another
// valid web session
func (c *HTTPClient) ExtendWebSession(ctx context.Context, req WebSessionReq) (types.WebSession, error) {
	out, err := c.PostJSON(ctx, c.Endpoint("users", req.User, "web", "sessions"), req)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return services.UnmarshalWebSession(out.Bytes())
}

// CreateWebSession creates a new web session for a user
func (c *HTTPClient) CreateWebSession(ctx context.Context, user string) (types.WebSession, error) {
	out, err := c.PostJSON(
		ctx,
		c.Endpoint("users", user, "web", "sessions"),
		WebSessionReq{User: user},
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return services.UnmarshalWebSession(out.Bytes())
}

// AuthenticateWebUser authenticates web user, creates and  returns web session
// in case if authentication is successful
func (c *HTTPClient) AuthenticateWebUser(ctx context.Context, req AuthenticateUserRequest) (types.WebSession, error) {
	out, err := c.PostJSON(
		ctx,
		c.Endpoint("users", req.Username, "web", "authenticate"),
		req,
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return services.UnmarshalWebSession(out.Bytes())
}

// AuthenticateSSHUser authenticates SSH console user, creates and  returns a pair of signed TLS and SSH
// short lived certificates as a result
func (c *HTTPClient) AuthenticateSSHUser(ctx context.Context, req AuthenticateSSHRequest) (*SSHLoginResponse, error) {
	out, err := c.PostJSON(
		ctx,
		c.Endpoint("users", req.Username, "ssh", "authenticate"),
		req,
	)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var re SSHLoginResponse
	if err := json.Unmarshal(out.Bytes(), &re); err != nil {
		return nil, trace.Wrap(err)
	}
	return &re, nil
}

// GetWebSessionInfo checks if a web sesion is valid, returns session id in case if
// it is valid, or error otherwise.
func (c *HTTPClient) GetWebSessionInfo(ctx context.Context, user, sessionID string) (types.WebSession, error) {
	out, err := c.Get(
		ctx,
		c.Endpoint("users", user, "web", "sessions", sessionID), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return services.UnmarshalWebSession(out.Bytes())
}

// DeleteWebSession deletes the web session specified with sid for the given user
func (c *HTTPClient) DeleteWebSession(ctx context.Context, user string, sid string) error {
	_, err := c.Delete(ctx, c.Endpoint("users", user, "web", "sessions", sid))
	return trace.Wrap(err)
}

// GenerateHostCert takes the public key in the Open SSH “authorized_keys“
// plain text format, signs it using Host Certificate Authority private key and returns the
// resulting certificate.
func (c *HTTPClient) GenerateHostCert(
	ctx context.Context, key []byte, hostID, nodeName string, principals []string, clusterName string, role types.SystemRole, ttl time.Duration,
) ([]byte, error) {
	out, err := c.PostJSON(ctx, c.Endpoint("ca", "host", "certs"),
		generateHostCertReq{
			Key:         key,
			HostID:      hostID,
			NodeName:    nodeName,
			Principals:  principals,
			ClusterName: clusterName,
			Roles:       types.SystemRoles{role},
			TTL:         ttl,
		})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var cert string
	if err := json.Unmarshal(out.Bytes(), &cert); err != nil {
		return nil, err
	}

	return []byte(cert), nil
}

// ValidateOIDCAuthCallback validates OIDC auth callback returned from redirect
func (c *HTTPClient) ValidateOIDCAuthCallback(ctx context.Context, q url.Values) (*OIDCAuthResponse, error) {
	out, err := c.PostJSON(ctx, c.Endpoint("oidc", "requests", "validate"), ValidateOIDCAuthCallbackReq{
		Query: q,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var rawResponse OIDCAuthRawResponse
	if err := json.Unmarshal(out.Bytes(), &rawResponse); err != nil {
		return nil, trace.Wrap(err)
	}
	response := OIDCAuthResponse{
		Username: rawResponse.Username,
		Identity: rawResponse.Identity,
		Cert:     rawResponse.Cert,
		Req:      rawResponse.Req,
		TLSCert:  rawResponse.TLSCert,
	}
	if len(rawResponse.Session) != 0 {
		session, err := services.UnmarshalWebSession(rawResponse.Session)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		response.Session = session
	}
	response.HostSigners = make([]types.CertAuthority, len(rawResponse.HostSigners))
	for i, raw := range rawResponse.HostSigners {
		ca, err := services.UnmarshalCertAuthority(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		response.HostSigners[i] = ca
	}
	return &response, nil
}

// ValidateSAMLResponse validates response returned by SAML identity provider
func (c *HTTPClient) ValidateSAMLResponse(ctx context.Context, re string, connectorID string) (*SAMLAuthResponse, error) {
	out, err := c.PostJSON(ctx, c.Endpoint("saml", "requests", "validate"), ValidateSAMLResponseReq{
		Response:    re,
		ConnectorID: connectorID,
	})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var rawResponse SAMLAuthRawResponse
	if err := json.Unmarshal(out.Bytes(), &rawResponse); err != nil {
		return nil, trace.Wrap(err)
	}
	response := SAMLAuthResponse{
		Username: rawResponse.Username,
		Identity: rawResponse.Identity,
		Cert:     rawResponse.Cert,
		Req:      rawResponse.Req,
		TLSCert:  rawResponse.TLSCert,
	}
	if len(rawResponse.Session) != 0 {
		session, err := services.UnmarshalWebSession(rawResponse.Session)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		response.Session = session
	}
	response.HostSigners = make([]types.CertAuthority, len(rawResponse.HostSigners))
	for i, raw := range rawResponse.HostSigners {
		ca, err := services.UnmarshalCertAuthority(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		response.HostSigners[i] = ca
	}
	return &response, nil
}

// ValidateGithubAuthCallback validates Github auth callback returned from redirect
func (c *HTTPClient) ValidateGithubAuthCallback(ctx context.Context, q url.Values) (*GithubAuthResponse, error) {
	out, err := c.PostJSON(ctx, c.Endpoint("github", "requests", "validate"),
		validateGithubAuthCallbackReq{Query: q})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var rawResponse githubAuthRawResponse
	if err := json.Unmarshal(out.Bytes(), &rawResponse); err != nil {
		return nil, trace.Wrap(err)
	}
	response := GithubAuthResponse{
		Username: rawResponse.Username,
		Identity: rawResponse.Identity,
		Cert:     rawResponse.Cert,
		Req:      rawResponse.Req,
		TLSCert:  rawResponse.TLSCert,
	}
	if len(rawResponse.Session) != 0 {
		session, err := services.UnmarshalWebSession(
			rawResponse.Session)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		response.Session = session
	}
	response.HostSigners = make([]types.CertAuthority, len(rawResponse.HostSigners))
	for i, raw := range rawResponse.HostSigners {
		ca, err := services.UnmarshalCertAuthority(raw)
		if err != nil {
			return nil, trace.Wrap(err)
		}
		response.HostSigners[i] = ca
	}
	return &response, nil
}

// GetSessionChunk allows clients to receive a byte array (chunk) from a recorded
// session stream, starting from 'offset', up to 'max' in length. The upper bound
// of 'max' is set to events.MaxChunkBytes
func (c *HTTPClient) GetSessionChunk(namespace string, sid session.ID, offsetBytes, maxBytes int) ([]byte, error) {
	if namespace == "" {
		return nil, trace.BadParameter(MissingNamespaceError)
	}
	response, err := c.Get(context.TODO(), c.Endpoint("namespaces", namespace, "sessions", string(sid), "stream"), url.Values{
		"offset": []string{strconv.Itoa(offsetBytes)},
		"bytes":  []string{strconv.Itoa(maxBytes)},
	})
	if err != nil {
		log.Error(err)
		return nil, trace.Wrap(err)
	}
	return response.Bytes(), nil
}

// Returns events that happen during a session sorted by time
// (oldest first).
//
// afterN allows to filter by "newer than N" value where N is the cursor ID
// of previously returned bunch (good for polling for latest)
func (c *HTTPClient) GetSessionEvents(namespace string, sid session.ID, afterN int, includePrintEvents bool) (retval []events.EventFields, err error) {
	if namespace == "" {
		return nil, trace.BadParameter(MissingNamespaceError)
	}
	query := make(url.Values)
	if afterN > 0 {
		query.Set("after", strconv.Itoa(afterN))
	}
	if includePrintEvents {
		query.Set("print", fmt.Sprintf("%v", includePrintEvents))
	}
	response, err := c.Get(context.TODO(), c.Endpoint("namespaces", namespace, "sessions", string(sid), "events"), query)
	if err != nil {
		return nil, trace.Wrap(err)
	}
	retval = make([]events.EventFields, 0)
	if err := json.Unmarshal(response.Bytes(), &retval); err != nil {
		return nil, trace.Wrap(err)
	}
	return retval, nil
}

// GetNamespaces returns a list of namespaces
func (c *HTTPClient) GetNamespaces() ([]types.Namespace, error) {
	out, err := c.Get(context.TODO(), c.Endpoint("namespaces"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	var re []types.Namespace
	if err := utils.FastUnmarshal(out.Bytes(), &re); err != nil {
		return nil, trace.Wrap(err)
	}
	return re, nil
}

// GetNamespace returns namespace by name
func (c *HTTPClient) GetNamespace(name string) (*types.Namespace, error) {
	if name == "" {
		return nil, trace.BadParameter("missing namespace name")
	}
	out, err := c.Get(context.TODO(), c.Endpoint("namespaces", name), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}
	return services.UnmarshalNamespace(out.Bytes())
}

// UpsertNamespace upserts namespace
func (c *HTTPClient) UpsertNamespace(ns types.Namespace) error {
	_, err := c.PostJSON(context.TODO(), c.Endpoint("namespaces"), upsertNamespaceReq{Namespace: ns})
	return trace.Wrap(err)
}

// DeleteNamespace deletes namespace by name
func (c *HTTPClient) DeleteNamespace(name string) error {
	_, err := c.Delete(context.TODO(), c.Endpoint("namespaces", name))
	return trace.Wrap(err)
}

// GetClusterName returns a cluster name
func (c *HTTPClient) GetClusterName(opts ...services.MarshalOption) (types.ClusterName, error) {
	out, err := c.Get(context.TODO(), c.Endpoint("configuration", "name"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	cn, err := services.UnmarshalClusterName(out.Bytes())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return cn, err
}

// SetClusterName sets cluster name once, will
// return Already Exists error if the name is already set
func (c *HTTPClient) SetClusterName(cn types.ClusterName) error {
	data, err := services.MarshalClusterName(cn)
	if err != nil {
		return trace.Wrap(err)
	}

	_, err = c.PostJSON(context.TODO(), c.Endpoint("configuration", "name"), &setClusterNameReq{ClusterName: data})
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

// DeleteStaticTokens deletes static tokens
func (c *HTTPClient) DeleteStaticTokens() error {
	_, err := c.Delete(context.TODO(), c.Endpoint("configuration", "static_tokens"))
	return trace.Wrap(err)
}

// GetStaticTokens returns a list of static register tokens
func (c *HTTPClient) GetStaticTokens() (types.StaticTokens, error) {
	out, err := c.Get(context.TODO(), c.Endpoint("configuration", "static_tokens"), url.Values{})
	if err != nil {
		return nil, trace.Wrap(err)
	}

	st, err := services.UnmarshalStaticTokens(out.Bytes())
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return st, err
}

// SetStaticTokens sets a list of static register tokens
func (c *HTTPClient) SetStaticTokens(st types.StaticTokens) error {
	data, err := services.MarshalStaticTokens(st)
	if err != nil {
		return trace.Wrap(err)
	}

	_, err = c.PostJSON(context.TODO(), c.Endpoint("configuration", "static_tokens"), &setStaticTokensReq{StaticTokens: data})
	if err != nil {
		return trace.Wrap(err)
	}

	return nil
}

func (c *HTTPClient) ValidateTrustedCluster(ctx context.Context, validateRequest *ValidateTrustedClusterRequest) (*ValidateTrustedClusterResponse, error) {
	validateRequestRaw, err := validateRequest.ToRaw()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	out, err := c.PostJSON(ctx, c.Endpoint("trustedclusters", "validate"), validateRequestRaw)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	var validateResponseRaw ValidateTrustedClusterResponseRaw
	err = json.Unmarshal(out.Bytes(), &validateResponseRaw)
	if err != nil {
		return nil, trace.Wrap(err)
	}

	validateResponse, err := validateResponseRaw.ToNative()
	if err != nil {
		return nil, trace.Wrap(err)
	}

	return validateResponse, nil
}
