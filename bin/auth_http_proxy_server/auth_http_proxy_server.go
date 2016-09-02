package main

import (
	"fmt"
	"net"
	"net/http"
	"runtime"

	auth_client "github.com/bborbe/auth/client/verify_group_service"
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/forward"
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/bborbe/auth_http_proxy/verifier"
	auth_verifier "github.com/bborbe/auth_http_proxy/verifier/auth"
	"github.com/bborbe/auth_http_proxy/verifier/cache"
	file_verifier "github.com/bborbe/auth_http_proxy/verifier/file"
	ldap_verifier "github.com/bborbe/auth_http_proxy/verifier/ldap"
	flag "github.com/bborbe/flagenv"
	http_client_builder "github.com/bborbe/http/client_builder"
	http_requestbuilder "github.com/bborbe/http/requestbuilder"
	"github.com/bborbe/http_handler/auth_basic"
	"github.com/bborbe/http_handler/auth_html"
	debug_handler "github.com/bborbe/http_handler/debug"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/golang/glog"
	"time"
)

const (
	defaultPort                      int = 8080
	parameterPort                        = "port"
	parameterTargetAddress               = "target-address"
	parameterBasicAuthRealm              = "basic-auth-realm"
	parameterRequiredGroups              = "required-groups"
	parameterVerifierType                = "verifier"
	parameterKind                        = "kind"
	parameterConfig                      = "config"
	parameterFileUsers                   = "file-users"
	parameterAuthUrl                     = "auth-url"
	parameterAuthApplicationName         = "auth-application-name"
	parameterAuthApplicationPassword     = "auth-application-password"
	parameterLdapBase                    = "ldap-base"
	parameterLdapHost                    = "ldap-host"
	parameterLdapPort                    = "ldap-port"
	parameterLdapUseSSL                  = "ldap-use-ssl"
	parameterLdapBindDN                  = "ldap-bind-dn"
	parameterLdapBindPassword            = "ldap-bind-password"
	parameterLdapUserFilter              = "ldap-user-filter"
	parameterLdapGroupFilter             = "ldap-group-filter"
	parameterCacheTTL                    = "cache-ttl"
)

var (
	portPtr           = flag.Int(parameterPort, defaultPort, "port")
	authUrlPtr        = flag.String(parameterAuthUrl, "", "auth url")
	basicAuthRealmPtr = flag.String(parameterBasicAuthRealm, "", "basic auth realm")
	targetAddressPtr  = flag.String(parameterTargetAddress, "", "target address")
	verifierPtr       = flag.String(parameterVerifierType, "", "verifier (auth,file,ldap)")
	kindPtr           = flag.String(parameterKind, "", "(basic,html)")
	configPtr         = flag.String(parameterConfig, "", "config")
	requiredGroupsPtr = flag.String(parameterRequiredGroups, "", "required groups reperated by comma")
	cacheTTLPtr       = flag.Duration(parameterCacheTTL, 5*time.Minute, "cache ttl")
	// file params
	fileUseresPtr = flag.String(parameterFileUsers, "", "users")
	// auth params
	authApplicationNamePtr     = flag.String(parameterAuthApplicationName, "", "auth application name")
	authApplicationPasswordPtr = flag.String(parameterAuthApplicationPassword, "", "auth application password")
	// ldap params
	ldapBasePtr         = flag.String(parameterLdapBase, "", "ldap-base")
	ldapHostPtr         = flag.String(parameterLdapHost, "", "ldap-host")
	ldapPortPtr         = flag.Int(parameterLdapPort, 0, "ldap-port")
	ldapUseSSLPtr       = flag.Bool(parameterLdapUseSSL, false, "ldap-use-ssl")
	ldapBindDNPtr       = flag.String(parameterLdapBindDN, "", "ldap-bind-dn")
	ldapBindPasswordPtr = flag.String(parameterLdapBindPassword, "", "ldap-bind-password")
	ldapUserFilterPtr   = flag.String(parameterLdapUserFilter, "", "ldap-user-filter")
	ldapGroupFilterPtr  = flag.String(parameterLdapGroupFilter, "", "ldap-group-filter")
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	if err := do(); err != nil {
		glog.Exit(err)
	}
}

func do() error {
	config, err := createConfig()
	if err != nil {
		return err
	}
	server, err := createServer(config)
	if err != nil {
		return err
	}
	glog.V(2).Infof("start server")
	return gracehttp.Serve(server)
}

func createConfig() (*model.Config, error) {
	var config *model.Config
	var err error
	configPath := model.ConfigPath(*configPtr)
	if configPath.IsValue() {
		glog.V(2).Infof("parse config from file: %v", configPath)
		config, err = configPath.Parse()
		if err != nil {
			glog.Warningf("parse config failed: %v", err)
			return nil, err
		}
	} else {
		glog.V(2).Infof("create empty config")
		config = new(model.Config)
	}
	if config.Port <= 0 {
		config.Port = model.Port(*portPtr)
	}
	if len(config.Kind) == 0 {
		config.Kind = model.Kind(*kindPtr)
	}
	if len(config.UserFile) == 0 {
		config.UserFile = model.UserFile(*fileUseresPtr)
	}
	if len(config.VerifierType) == 0 {
		config.VerifierType = model.VerifierType(*verifierPtr)
	}
	if len(config.BasicAuthRealm) == 0 {
		config.BasicAuthRealm = model.BasicAuthRealm(*basicAuthRealmPtr)
	}
	if len(config.TargetAddress) == 0 {
		config.TargetAddress = model.TargetAddress(*targetAddressPtr)
	}
	if config.CacheTTL.IsEmpty() {
		config.CacheTTL = model.CacheTTL(*cacheTTLPtr)
	}
	if len(config.AuthUrl) == 0 {
		config.AuthUrl = model.AuthUrl(*authUrlPtr)
	}
	if len(config.AuthApplicationName) == 0 {
		config.AuthApplicationName = model.AuthApplicationName(*authApplicationNamePtr)
	}
	if len(config.AuthApplicationPassword) == 0 {
		config.AuthApplicationPassword = model.AuthApplicationPassword(*authApplicationPasswordPtr)
	}
	if len(config.RequiredGroups) == 0 {
		config.RequiredGroups = model.CreateGroupsFromString(*requiredGroupsPtr)
	}
	if len(config.LdapBase) == 0 {
		config.LdapBase = model.LdapBase(*ldapBasePtr)
	}
	if len(config.LdapHost) == 0 {
		config.LdapHost = model.LdapHost(*ldapHostPtr)
	}
	if config.LdapPort <= 0 {
		config.LdapPort = model.LdapPort(*ldapPortPtr)
	}
	if !config.LdapUseSSL {
		config.LdapUseSSL = model.LdapUseSSL(*ldapUseSSLPtr)
	}
	if len(config.LdapBindDN) == 0 {
		config.LdapBindDN = model.LdapBindDN(*ldapBindDNPtr)
	}
	if len(config.LdapBindPassword) == 0 {
		config.LdapBindPassword = model.LdapBindPassword(*ldapBindPasswordPtr)
	}
	if len(config.LdapUserFilter) == 0 {
		config.LdapUserFilter = model.LdapUserFilter(*ldapUserFilterPtr)
	}
	if len(config.LdapGroupFilter) == 0 {
		config.LdapGroupFilter = model.LdapGroupFilter(*ldapGroupFilterPtr)
	}
	return config, nil
}

func createServer(config *model.Config) (*http.Server, error) {
	glog.V(2).Infof("create server")
	handler, err := createHandler(config)
	if err != nil {
		return nil, err
	}
	if config.Port <= 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterPort)
	}
	return &http.Server{Addr: fmt.Sprintf(":%d", config.Port), Handler: handler}, nil
}

func createHandler(config *model.Config) (http.Handler, error) {
	glog.V(2).Infof("create handler")
	handler, err := createHttpFilter(config)
	if err != nil {
		return nil, err
	}
	if glog.V(4) {
		glog.V(2).Infof("add debug handler")
		handler = debug_handler.New(handler)
	}
	return handler, nil
}

func createHttpFilter(config *model.Config) (http.Handler, error) {
	if len(config.Kind) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterKind)
	}
	glog.V(2).Infof("get auth filter for: %v", config.Kind)
	switch config.Kind {
	case "html":
		return createHtmlAuthHttpFilter(config)
	case "basic":
		return createBasicAuthHttpFilter(config)
	}
	return nil, fmt.Errorf("parameter %s invalid", parameterKind)
}

func createHtmlAuthHttpFilter(config *model.Config) (http.Handler, error) {
	verifier, err := createVerifier(config)
	if err != nil {
		return nil, err
	}
	forwardHandler, err := createForwardHandler(config)
	if err != nil {
		return nil, err
	}
	handler := auth_html.New(forwardHandler.ServeHTTP,
		func(username string, password string) (bool, error) {
			return verifier.Verify(model.UserName(username), model.Password(password))
		})
	return handler, nil
}

func createBasicAuthHttpFilter(config *model.Config) (http.Handler, error) {
	verifier, err := createVerifier(config)
	if err != nil {
		return nil, err
	}
	forwardHandler, err := createForwardHandler(config)
	if err != nil {
		return nil, err
	}
	if len(config.BasicAuthRealm) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterBasicAuthRealm)
	}
	handler := auth_basic.New(forwardHandler.ServeHTTP,
		func(username string, password string) (bool, error) {
			return verifier.Verify(model.UserName(username), model.Password(password))
		}, config.BasicAuthRealm.String())
	return handler, nil
}

func createForwardHandler(config *model.Config) (http.Handler, error) {
	dialer := (&net.Dialer{
		Timeout: http_client_builder.DEFAULT_TIMEOUT,
	})
	if len(config.TargetAddress) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterTargetAddress)
	}
	forwardHandler := forward.New(config.TargetAddress,
		func(address string, req *http.Request) (resp *http.Response, err error) {
			return http_client_builder.New().WithoutProxy().WithoutRedirects().WithDialFunc(
				func(network, address string) (net.Conn, error) {
					return dialer.Dial(network, config.TargetAddress.String())
				}).BuildRoundTripper().RoundTrip(req)
		})
	return forwardHandler, nil
}

func createVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.VerifierType) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterVerifierType)
	}
	glog.V(2).Infof("get verifier for: %v", config.VerifierType)
	switch config.VerifierType {
	case "auth":
		return createAuthVerifier(config)
	case "ldap":
		return createLdapVerifier(config)
	case "file":
		return createFileVerifier(config)
	}
	return nil, fmt.Errorf("parameter %s invalid", parameterVerifierType)
}

func createAuthVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.AuthUrl) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthUrl)
	}
	if len(config.AuthApplicationName) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationName)
	}
	if len(config.AuthApplicationPassword) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationPassword)
	}
	httpRequestBuilderProvider := http_requestbuilder.NewHTTPRequestBuilderProvider()
	httpClient := http_client_builder.New().WithoutProxy().Build()
	authClient := auth_client.New(httpClient.Do, httpRequestBuilderProvider, auth_model.Url(config.AuthUrl), auth_model.ApplicationName(config.AuthApplicationName), auth_model.ApplicationPassword(config.AuthApplicationPassword))
	return cache.New(auth_verifier.New(
		authClient.Auth,
		config.RequiredGroups...,
	), config.CacheTTL), nil
}

func createLdapVerifier(config *model.Config) (verifier.Verifier, error) {
	return cache.New(ldap_verifier.New(
		config.LdapBase,
		config.LdapHost,
		config.LdapPort,
		config.LdapUseSSL,
		config.LdapBindDN,
		config.LdapBindPassword,
		config.LdapUserFilter,
		config.LdapGroupFilter,
		config.RequiredGroups...,
	), config.CacheTTL), nil
}

func createFileVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.UserFile) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterFileUsers)
	}
	return cache.New(file_verifier.New(config.UserFile), config.CacheTTL), nil
}
