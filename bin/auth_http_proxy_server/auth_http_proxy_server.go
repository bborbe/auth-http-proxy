package main

import (
	"fmt"
	"net/http"
	"os"

	flag "github.com/bborbe/flagenv"
	http_client_builder "github.com/bborbe/http/client_builder"
	http_requestbuilder "github.com/bborbe/http/requestbuilder"
	"github.com/bborbe/log"

	"net"

	"strings"

	auth_api "github.com/bborbe/auth/api"
	auth_client "github.com/bborbe/auth/client"
	"github.com/bborbe/auth_http_proxy/auth_verifier"
	"github.com/bborbe/auth_http_proxy/forward"
	"github.com/bborbe/server/handler/auth_basic"
	"github.com/facebookgo/grace/gracehttp"
)

var logger = log.DefaultLogger

const (
	DEFAULT_PORT                        int = 8080
	PARAMETER_LOGLEVEL                      = "loglevel"
	PARAMETER_PORT                          = "port"
	PARAMETER_AUTH_ADDRESS                  = "auth-address"
	PARAMETER_AUTH_APPLICATION_NAME         = "auth-application-name"
	PARAMETER_AUTH_APPLICATION_PASSWORD     = "auth-application-password"
	PARAMETER_TARGET_ADDRESS                = "target-address"
	PARAMETER_AUTH_REALM                    = "auth-realm"
	PARAMETER_AUTH_GROUPS                   = "auth-groups"
)

var (
	logLevelPtr                = flag.String(PARAMETER_LOGLEVEL, log.INFO_STRING, "one of OFF,TRACE,DEBUG,INFO,WARN,ERROR")
	portPtr                    = flag.Int(PARAMETER_PORT, DEFAULT_PORT, "port")
	authAddressPtr             = flag.String(PARAMETER_AUTH_ADDRESS, "", "auth address")
	authApplicationNamePtr     = flag.String(PARAMETER_AUTH_APPLICATION_NAME, "", "auth application name")
	authApplicationPasswordPtr = flag.String(PARAMETER_AUTH_APPLICATION_PASSWORD, "", "auth application password")
	authRealmPtr               = flag.String(PARAMETER_AUTH_REALM, "", "basic auth realm")
	authGroupsPtr              = flag.String(PARAMETER_AUTH_GROUPS, "", "required groups reperated by comma")
	targetAddressPtr           = flag.String(PARAMETER_TARGET_ADDRESS, "", "target address")
)

func main() {
	defer logger.Close()
	flag.Parse()

	logger.SetLevelThreshold(log.LogStringToLevel(*logLevelPtr))
	logger.Debugf("set log level to %s", *logLevelPtr)

	server, err := createServer(
		*portPtr,
		auth_api.Address(*authAddressPtr),
		auth_api.ApplicationName(*authApplicationNamePtr),
		auth_api.ApplicationPassword(*authApplicationPasswordPtr),
		*authRealmPtr,
		*authGroupsPtr,
		*targetAddressPtr,
	)
	if err != nil {
		logger.Fatal(err)
		logger.Close()
		os.Exit(1)
	}
	logger.Debugf("start server")
	gracehttp.Serve(server)
}

func createServer(
	port int,
	authAddress auth_api.Address,
	authApplicationName auth_api.ApplicationName,
	authApplicationPassword auth_api.ApplicationPassword,
	authRealm string,
	authGroups string,
	targetAddress string,
) (*http.Server, error) {
	if port <= 0 {
		return nil, fmt.Errorf("parameter %s missing", PARAMETER_PORT)
	}
	if len(authAddress) == 0 {
		return nil, fmt.Errorf("parameter %s missing", PARAMETER_AUTH_ADDRESS)
	}
	if len(authApplicationName) == 0 {
		return nil, fmt.Errorf("parameter %s missing", PARAMETER_AUTH_APPLICATION_NAME)
	}
	if len(authApplicationPassword) == 0 {
		return nil, fmt.Errorf("parameter %s missing", PARAMETER_AUTH_APPLICATION_PASSWORD)
	}
	if len(targetAddress) == 0 {
		return nil, fmt.Errorf("parameter %s missing", PARAMETER_TARGET_ADDRESS)
	}
	if len(authRealm) == 0 {
		return nil, fmt.Errorf("parameter %s missing", PARAMETER_AUTH_REALM)
	}

	logger.Debugf("create server on port: %d with target: %s", port, targetAddress)

	httpRequestBuilderProvider := http_requestbuilder.NewHttpRequestBuilderProvider()
	httpClient := http_client_builder.New().WithoutProxy().Build()
	authClient := auth_client.New(httpClient.Do, httpRequestBuilderProvider, authAddress, authApplicationName, authApplicationPassword)
	dialer := (&net.Dialer{
		Timeout: http_client_builder.DEFAULT_TIMEOUT,
	})
	forwardHandler := forward.New(targetAddress, func(address string, req *http.Request) (resp *http.Response, err error) {
		return http_client_builder.New().WithoutProxy().WithDialFunc(func(network, address string) (net.Conn, error) {
			return dialer.Dial(network, targetAddress)
		}).Build().Do(req)
	})
	authVerifier := auth_verifier.New(authClient.Auth, createGroups(authGroups))
	authHandler := auth_basic.New(forwardHandler.ServeHTTP, authVerifier.Verify, authRealm)

	return &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: authHandler}, nil
}

func createGroups(groupNames string) []auth_api.GroupName {
	parts := strings.Split(groupNames, ",")
	groups := make([]auth_api.GroupName, 0)
	for _, groupName := range parts {
		if len(groupName) > 0 {
			groups = append(groups, auth_api.GroupName(groupName))
		}
	}
	logger.Debugf("required groups: %v", groups)
	return groups
}
