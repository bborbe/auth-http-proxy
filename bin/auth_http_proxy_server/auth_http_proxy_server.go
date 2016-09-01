package main

import (
	"fmt"
	auth_client "github.com/bborbe/auth/client/verify_group_service"
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/forward"
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/bborbe/auth_http_proxy/verifier"
	auth_verifier "github.com/bborbe/auth_http_proxy/verifier/auth"
	"github.com/bborbe/auth_http_proxy/verifier/file"
	"github.com/bborbe/auth_http_proxy/verifier/ldap"
	flag "github.com/bborbe/flagenv"
	http_client_builder "github.com/bborbe/http/client_builder"
	http_requestbuilder "github.com/bborbe/http/requestbuilder"
	"github.com/bborbe/http_handler/auth_basic"
	"github.com/bborbe/http_handler/auth_html"
	debug_handler "github.com/bborbe/http_handler/debug"
	"github.com/facebookgo/grace/gracehttp"
	"github.com/golang/glog"
	"net"
	"net/http"
	"runtime"
)

const (
	defaultPort                      int = 8080
	parameterPort                        = "port"
	parameterAuthUrl                     = "auth-url"
	parameterAuthApplicationName         = "auth-application-name"
	parameterAuthApplicationPassword     = "auth-application-password"
	parameterTargetAddress               = "target-address"
	parameterBasicAuthRealm              = "basic-auth-realm"
	parameterAuthGroups                  = "auth-groups"
	parameterDebug                       = "debug"
	parameterVerifierType                = "verifier"
	parameterFileUsers                   = "file-users"
	parameterKind                        = "kind"
	parameterConfig                      = "config"
)

var (
	portPtr                    = flag.Int(parameterPort, defaultPort, "port")
	authUrlPtr                 = flag.String(parameterAuthUrl, "", "auth url")
	basicAuthRealmPtr          = flag.String(parameterBasicAuthRealm, "", "basic auth realm")
	targetAddressPtr           = flag.String(parameterTargetAddress, "", "target address")
	debugPtr                   = flag.Bool(parameterDebug, false, "debug")
	verifierPtr                = flag.String(parameterVerifierType, "", "verifier (auth,file,ldap)")
	authApplicationNamePtr     = flag.String(parameterAuthApplicationName, "", "auth application name")
	authApplicationPasswordPtr = flag.String(parameterAuthApplicationPassword, "", "auth application password")
	authGroupsPtr              = flag.String(parameterAuthGroups, "", "required groups reperated by comma")
	userFilePtr                = flag.String(parameterFileUsers, "", "users")
	kindPtr                    = flag.String(parameterKind, "", "(basic,html)")
	configPtr                  = flag.String(parameterConfig, "", "config")
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
	if config.Port == 0 {
		config.Port = model.Port(*portPtr)
	}
	if !config.Debug {
		config.Debug = model.Debug(*debugPtr)
	}
	if len(config.Kind) == 0 {
		config.Kind = model.Kind(*kindPtr)
	}
	if len(config.UserFile) == 0 {
		config.UserFile = model.UserFile(*userFilePtr)
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
	if len(config.AuthUrl) == 0 {
		config.AuthUrl = model.AuthUrl(*authUrlPtr)
	}
	if len(config.AuthApplicationName) == 0 {
		config.AuthApplicationName = model.AuthApplicationName(*authApplicationNamePtr)
	}
	if len(config.AuthApplicationPassword) == 0 {
		config.AuthApplicationPassword = model.AuthApplicationPassword(*authApplicationPasswordPtr)
	}
	if len(config.AuthGroups) == 0 {
		config.AuthGroups = model.CreateGroupsFromString(*authGroupsPtr)
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
	if config.Debug {
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
	verifier, err := getVerifierByType(config)
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
	verifier, err := getVerifierByType(config)
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
			return http_client_builder.New().WithoutProxy().WithDialFunc(
				func(network, address string) (net.Conn, error) {
					return dialer.Dial(network, config.TargetAddress.String())
				}).Build().Do(req)
		})
	return forwardHandler, nil
}

func getVerifierByType(config *model.Config) (verifier.Verifier, error) {
	if len(config.VerifierType) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterVerifierType)
	}
	glog.V(2).Infof("get verifier for: %v", config.VerifierType)
	switch config.VerifierType {
	case "auth":
		return createAuthVerifier(config)
	case "ldap":
		return createLdapVerifier()
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
	return auth_verifier.New(authClient.Auth, config.AuthGroups...), nil
}

func createLdapVerifier() (verifier.Verifier, error) {
	return ldap.New(), nil
}

func createFileVerifier(config *model.Config) (verifier.Verifier, error) {
	if len(config.UserFile) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterFileUsers)
	}
	return file.New(config.UserFile), nil
}
