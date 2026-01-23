package httpconf

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/brnsampson/ezconf/file"
	"github.com/brnsampson/optional"
)

// Loaders for generic fields

type Loader[Conf any] interface {
	Update() (Conf, error)
	Previous() (Conf, error)
}

type HTTPConfLoader Loader[HttpServerConfig]

// Loading HTTP and TLS config from files is not that bad, but this removes a lot of boilerplate that exists in every web-based project.
type HttpServerConfigProtos int

const (
	HTTP HttpServerConfigProtos = iota
	HTTPS
	HTTP2
	UNENCRYPTEDHTTP2
)

func (p HttpServerConfigProtos) String() string {
	switch p {
	case HTTP:
		return "http"
	case HTTPS:
		return "https"
	case HTTP2:
		return "https"
	case UNENCRYPTEDHTTP2:
		return "hcl"
	default:
		return "unknown"
	}
}

func (p HttpServerConfigProtos) GetHttpProtos() *http.Protocols {
	protos := new(http.Protocols)
	switch p {
	case HTTP | HTTPS:
		protos.SetHTTP1(true)
	case HTTP2:
		protos.SetHTTP2(true)
	case UNENCRYPTEDHTTP2:
		protos.SetUnencryptedHTTP2(true)
	default: // Default to just allowing either http or http2
		protos.SetHTTP1(true)
		protos.SetHTTP2(true)
	}

	return protos
}

// HttpServerConfig is the struct produced by the loader. It has pretty much everything needed to create an http.Server
type HttpServerConfig struct {
	Protos            *http.Protocols
	Hostname          string      // The hostname as given OR ip address if no hostname was given.
	BindAddr          string      // The address to bind to
	Port              uint16      // The port to bind to
	RemoteAddress     string      // The address clients should connect to. This is generally [proto]://[hostname]:[port] (although port is omitted if it is the standard http[s] port)
	TlsConf           *tls.Config // TLS config to use. If tls was disabled you can still use this and it will correctly be a non-TLS connetion.
	handler           http.Handler
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	maxHeaderBytes    int
	errorLog          *log.Logger
}

type HttpServerConfigOption func(HttpServerConfig) HttpServerConfig

func HttpHandler(handler http.Handler) HttpServerConfigOption {
	return func(c HttpServerConfig) HttpServerConfig {
		c.handler = handler
		return c
	}
}

func HttpReadTimeout(timeout time.Duration) HttpServerConfigOption {
	return func(c HttpServerConfig) HttpServerConfig {
		c.readTimeout = timeout
		return c
	}
}

func HttpReadHeaderTimeout(timeout time.Duration) HttpServerConfigOption {
	return func(c HttpServerConfig) HttpServerConfig {
		c.readHeaderTimeout = timeout
		return c
	}
}

func HttpMaxHeaderBytes(count int) HttpServerConfigOption {
	return func(c HttpServerConfig) HttpServerConfig {
		c.maxHeaderBytes = count
		return c
	}
}

func HttpErrorLog(errLog *log.Logger) HttpServerConfigOption {
	return func(c HttpServerConfig) HttpServerConfig {
		c.errorLog = errLog
		return c
	}
}

func (c HttpServerConfig) With(o HttpServerConfigOption) HttpServerConfig {
	return o(c)
}

// NewHttpServer returns an *http.Http configured according to HttpServerConfig's fields.
//
// Calling (HttpServerConfig.NewHttpServer()).ListenAndServe() should do what you want most of the time unless you
// have specific needs.
func (c HttpServerConfig) NewHttpServer() *http.Server {
	port := strconv.FormatUint(uint64(c.Port), 10)
	var addr string
	if c.BindAddr != "" {
		ip := net.ParseIP(c.BindAddr)
		addr = ip.String() + ":" + port
	} else {
		// http.Server will accept an empty ip address to bind to all available interfaces.
		addr = ":" + port
	}
	return &http.Server{Addr: addr, Handler: c.handler, TLSConfig: c.TlsConf, ReadTimeout: c.readTimeout, ReadHeaderTimeout: c.readHeaderTimeout, ErrorLog: c.errorLog, Protocols: c.Protos}
}

// HttpServerLoader gets parameters from the environment and user overrides in order to produce an HttpServerConfig struct.
// The HttpServerConfig struct in turn can be used to create a new http.Server.
type HttpServerLoader struct {
	Protocol          optional.Option[HttpServerConfigProtos] // Default behavior is to set this based on Tls.TlsEnabled.
	Hostname          optional.Str
	BindAddr          optional.Str    // an empty string will cause us to bind to all interfaces. Defaults to 127.0.0.1
	BindPort          optional.Uint16 // Defaults to 80 for HTTP, 443 for HTTPS
	Tls               Loader[*tls.Config]
	ReadTimeout       optional.Duration // Defaults to 0. Same as http.Server
	ReadHeaderTimeout optional.Duration // Defaults to 0. Same as http.Server
	MaxHeaderBytes    optional.Int
	handler           http.Handler
	errorLog          *log.Logger
	prev              HttpServerConfig
}

type HttpServerLoaderOption func(HttpServerLoader) HttpServerLoader

func HttpLoaderHandler(handler http.Handler) HttpServerLoaderOption {
	return func(c HttpServerLoader) HttpServerLoader {
		c.handler = handler
		return c
	}
}

func HttpLoaderErrorLog(errLog *log.Logger) HttpServerLoaderOption {
	return func(c HttpServerLoader) HttpServerLoader {
		c.errorLog = errLog
		return c
	}
}

func (c HttpServerLoader) With(o HttpServerLoaderOption) HttpServerLoader {
	return o(c)
}

func (l *HttpServerLoader) Previous() HttpServerConfig {
	return l.prev
}

func (l *HttpServerLoader) Update() (result HttpServerConfig, err error) {
	// Produce new config
	proto := optional.GetOr(l.Protocol, HTTPS) // Default to HTTPS because we don't have anything better to do.
	bindAddr := optional.GetOr(l.BindAddr, "127.0.0.0")
	hostname := optional.GetOr(l.Hostname, bindAddr)
	port, ok := l.BindPort.Get()
	if !ok {
		switch proto {
		case HTTP:
			port = 80
		case HTTPS:
			port = 443
		default:
			return result, fmt.Errorf("Failed to update HttpServerLoader: BindPort unset, but protocol only has default ports for HTTP and HTTPS")
		}
	}

	var remoteAddr string
	if ok {
		remoteAddr = proto.String() + "://" + hostname + ":" + strconv.FormatUint(uint64(port), 10)
	} else {
		// If we defaulted to a port, that means we should not have to specify it in the url
		remoteAddr = proto.String() + "://" + hostname
	}

	tlsConf, err := l.Tls.Update()
	if err != nil {
		return
	}

	// Update internal state
	result = HttpServerConfig{
		Protos:            proto.GetHttpProtos(),
		Hostname:          hostname,   // The hostname as given OR ip address if no hostname was given.
		BindAddr:          bindAddr,   // The address to bind to
		Port:              port,       // The port to bind to
		RemoteAddress:     remoteAddr, // The address clients should connect to. This is generally [proto]://[hostname]:[port] (although port is omitted if it is the standard http[s] port)
		TlsConf:           tlsConf,    // TLS config to use. If tls was disabled you can still use this and it will correctly be a non-TLS connetion.
		handler:           l.handler,
		readTimeout:       optional.GetOr(l.ReadTimeout, 0),
		readHeaderTimeout: optional.GetOr(l.ReadHeaderTimeout, 0),
		maxHeaderBytes:    optional.GetOr(l.MaxHeaderBytes, 0),
		errorLog:          l.errorLog,
	}

	l.prev = result
	return
}

type TlsConfigLoader struct {
	TlsEnabled         optional.Bool
	ServerName         optional.Str
	PrivateKey         file.PrivateKey `default:"tls/key.pem"`
	Certificate        file.Cert       `default:"tls/cert.pem"`
	InsecureSkipVerify optional.Bool   `default:"false"`
	prev               *tls.Config
}

func (l *TlsConfigLoader) Previous() *tls.Config {
	return l.prev
}

func (l *TlsConfigLoader) Update() (config *tls.Config, err error) {
	enabled := optional.GetOr(l.TlsEnabled, false)
	skipVerify := optional.GetOr(l.InsecureSkipVerify, false)
	name := l.ServerName
	cert := l.Certificate
	key := l.PrivateKey

	// Validate key error modes
	if enabled && (cert.IsNone() || key.IsNone()) {
		// Cert and key not specified, so we can't continue with tls enabled
		return config, fmt.Errorf("TLS was enabled, but cert or key file was not set.")
	}

	if enabled && !(name.IsSome() || skipVerify) {
		// If TLS is enabled, then ServerName must be specified unless InsecureSkipVerify is set.
		// Otherwise we could not actually validate against the certificate.
		return nil, fmt.Errorf("cannot make a valid tls config with enabled=true, insecureSkipVerify=false, serverName=None")
	}

	// Create the config
	if enabled {
		cert, err := key.ReadCert(cert)
		if err != nil {
			return config, err
		}

		config = &tls.Config{
			Certificates:     []tls.Certificate{cert},
			MinVersion:       tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			InsecureSkipVerify: skipVerify,
		}
	} else {
		config = &tls.Config{
			Certificates:     []tls.Certificate{},
			MinVersion:       tls.VersionTLS13,
			CurvePreferences: []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
			CipherSuites: []uint16{
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_RSA_WITH_AES_256_CBC_SHA,
			},
			InsecureSkipVerify: skipVerify,
		}
	}

	serverName, ok := name.Get()
	if ok {
		config.ServerName = serverName
	}

	l.prev = config
	return config, nil
}
