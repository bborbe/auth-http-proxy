package main

import (
	"fmt"
	"net/http"

	debug_handler "github.com/bborbe/http_handler/debug"

	flag "github.com/bborbe/flagenv"
	http_client_builder "github.com/bborbe/http/client_builder"
	http_requestbuilder "github.com/bborbe/http/requestbuilder"
	"github.com/golang/glog"

	"net"

	"strings"

	"runtime"

	auth_client "github.com/bborbe/auth/client/verify_group_service"
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/forward"
	"github.com/bborbe/auth_http_proxy/model"
	auth_verifier "github.com/bborbe/auth_http_proxy/verifier/auth"
	"github.com/bborbe/http_handler/auth_basic"
	"github.com/facebookgo/grace/gracehttp"
)

const (
	defaultPort                      int = 8080
	parameterPort                        = "port"
	parameterAuthUrl                     = "auth-url"
	parameterAuthApplicationName         = "auth-application-name"
	parameterAuthApplicationPassword     = "auth-application-password"
	parameterTargetAddress               = "target-address"
	parameterAuthRealm                   = "auth-realm"
	parameterAuthGroups                  = "auth-groups"
	parameterDebug                       = "debug"
)

var (
	portPtr                    = flag.Int(parameterPort, defaultPort, "port")
	authUrlPtr                 = flag.String(parameterAuthUrl, "", "auth url")
	authApplicationNamePtr     = flag.String(parameterAuthApplicationName, "", "auth application name")
	authApplicationPasswordPtr = flag.String(parameterAuthApplicationPassword, "", "auth application password")
	authRealmPtr               = flag.String(parameterAuthRealm, "", "basic auth realm")
	authGroupsPtr              = flag.String(parameterAuthGroups, "", "required groups reperated by comma")
	targetAddressPtr           = flag.String(parameterTargetAddress, "", "target address")
	debugPtr                   = flag.Bool(parameterDebug, false, "debug")
)

func main() {
	defer glog.Flush()
	glog.CopyStandardLogTo("info")
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())

	err := do(
		*portPtr,
		*debugPtr,
		auth_model.Url(*authUrlPtr),
		auth_model.ApplicationName(*authApplicationNamePtr),
		auth_model.ApplicationPassword(*authApplicationPasswordPtr),
		*authRealmPtr,
		*authGroupsPtr,
		*targetAddressPtr,
	)
	if err != nil {
		glog.Exit(err)
	}
}

func do(
	port int,
	debug bool,
	authUrl auth_model.Url,
	authApplicationName auth_model.ApplicationName,
	authApplicationPassword auth_model.ApplicationPassword,
	authRealm string,
	authGroups string,
	targetAddress string,
) error {
	server, err := createServer(
		port,
		debug,
		authUrl,
		authApplicationName,
		authApplicationPassword,
		authRealm,
		authGroups,
		targetAddress,
	)
	if err != nil {
		return err
	}
	glog.V(2).Infof("start server")
	return gracehttp.Serve(server)
}

func createServer(
	port int,
	debug bool,
	authUrl auth_model.Url,
	authApplicationName auth_model.ApplicationName,
	authApplicationPassword auth_model.ApplicationPassword,
	authRealm string,
	authGroups string,
	targetAddress string,
) (*http.Server, error) {
	glog.Infof("port %d debug %v authUrl %v authApplicationName %v authApplicationPassword-length %d authRealm %v authGroups %v targetAddress %v", port, debug, authUrl, authApplicationName, len(authApplicationPassword), authRealm, authGroups, targetAddress)

	if port <= 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterPort)
	}
	if len(targetAddress) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterTargetAddress)
	}
	if len(authRealm) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthRealm)
	}

	if len(authUrl) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthUrl)
	}
	if len(authApplicationName) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationName)
	}
	if len(authApplicationPassword) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationPassword)
	}

	httpRequestBuilderProvider := http_requestbuilder.NewHTTPRequestBuilderProvider()
	httpClient := http_client_builder.New().WithoutProxy().Build()
	authClient := auth_client.New(httpClient.Do, httpRequestBuilderProvider, authUrl, authApplicationName, authApplicationPassword)
	dialer := (&net.Dialer{
		Timeout: http_client_builder.DEFAULT_TIMEOUT,
	})
	forwardHandler := forward.New(targetAddress, func(address string, req *http.Request) (resp *http.Response, err error) {
		return http_client_builder.New().WithoutProxy().WithDialFunc(func(network, address string) (net.Conn, error) {
			return dialer.Dial(network, targetAddress)
		}).Build().Do(req)
	})
	authVerifier := auth_verifier.New(authClient.Auth, createGroups(authGroups)...)
	var handler http.Handler = auth_basic.New(forwardHandler.ServeHTTP, func(username string, password string) (bool, error) {
		return authVerifier.Verify(model.UserName(username), model.Password(password))
	}, authRealm)

	if debug {
		glog.V(2).Infof("add debug handler")
		handler = debug_handler.New(handler)
	}

	return &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: handler}, nil
}

func createGroups(groupNames string) []model.GroupName {
	parts := strings.Split(groupNames, ",")
	groups := make([]model.GroupName, 0)
	for _, groupName := range parts {
		if len(groupName) > 0 {
			groups = append(groups, model.GroupName(groupName))
		}
	}
	glog.V(1).Infof("required groups: %v", groups)
	return groups
}
