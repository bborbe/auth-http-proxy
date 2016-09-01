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

	"runtime"

	auth_client "github.com/bborbe/auth/client/verify_group_service"
	auth_model "github.com/bborbe/auth/model"
	"github.com/bborbe/auth_http_proxy/forward"
	"github.com/bborbe/auth_http_proxy/model"
	"github.com/bborbe/auth_http_proxy/verifier"
	auth_verifier "github.com/bborbe/auth_http_proxy/verifier/auth"
	"github.com/bborbe/auth_http_proxy/verifier/file"
	"github.com/bborbe/auth_http_proxy/verifier/ldap"
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
	parameterBasicAuthRealm              = "basic-auth-realm"
	parameterAuthGroups                  = "auth-groups"
	parameterDebug                       = "debug"
	parameterVerifierType                = "verifier"
	parameterFileUsers                   = "file-users"
)

var (
	portPtr                    = flag.Int(parameterPort, defaultPort, "port")
	authUrlPtr                 = flag.String(parameterAuthUrl, "", "auth url")
	basicAuthRealmPtr          = flag.String(parameterBasicAuthRealm, "", "basic auth realm")
	targetAddressPtr           = flag.String(parameterTargetAddress, "", "target address")
	debugPtr                   = flag.Bool(parameterDebug, false, "debug")
	verifierPtr                = flag.String(parameterVerifierType, "", "verifier type")
	authApplicationNamePtr     = flag.String(parameterAuthApplicationName, "", "auth application name")
	authApplicationPasswordPtr = flag.String(parameterAuthApplicationPassword, "", "auth application password")
	authGroupsPtr              = flag.String(parameterAuthGroups, "", "required groups reperated by comma")
	fileUsersPtr               = flag.String(parameterFileUsers, "", "users")
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
	server, err := createServer()
	if err != nil {
		return err
	}
	glog.V(2).Infof("start server")
	return gracehttp.Serve(server)
}

func createServer() (*http.Server, error) {
	glog.V(2).Infof("create server")

	handler, err := createHandler()
	if err != nil {
		return nil, err
	}

	if *portPtr <= 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterPort)
	}
	return &http.Server{Addr: fmt.Sprintf(":%d", *portPtr), Handler: handler}, nil
}

func createHandler() (http.Handler, error) {
	glog.V(2).Infof("create handler")

	if len(*targetAddressPtr) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterTargetAddress)
	}
	dialer := (&net.Dialer{
		Timeout: http_client_builder.DEFAULT_TIMEOUT,
	})
	forwardHandler := forward.New(*targetAddressPtr, func(address string, req *http.Request) (resp *http.Response, err error) {
		return http_client_builder.New().WithoutProxy().WithDialFunc(func(network, address string) (net.Conn, error) {
			return dialer.Dial(network, *targetAddressPtr)
		}).Build().Do(req)
	})
	if len(*basicAuthRealmPtr) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterBasicAuthRealm)
	}
	verifier, err := getVerifierByType()
	if err != nil {
		return nil, err
	}
	var handler http.Handler = auth_basic.New(forwardHandler.ServeHTTP, func(username string, password string) (bool, error) {
		return verifier.Verify(model.UserName(username), model.Password(password))
	}, *basicAuthRealmPtr)

	if *debugPtr {
		glog.V(2).Infof("add debug handler")
		handler = debug_handler.New(handler)
	}
	return handler, nil
}

func getVerifierByType() (verifier.Verifier, error) {
	if len(*verifierPtr) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterVerifierType)
	}
	glog.V(2).Infof("get verifier for type: %v", *verifierPtr)
	switch *verifierPtr {
	case "auth":
		return createAuthVerifier()
	case "ldap":
		return createLdapVerifier()
	case "file":
		return createFileVerifier()
	}
	return nil, fmt.Errorf("parameter %s invalid", parameterVerifierType)
}

func createAuthVerifier() (verifier.Verifier, error) {

	if len(*authUrlPtr) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthUrl)
	}
	if len(*authApplicationNamePtr) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationName)
	}
	if len(*authApplicationPasswordPtr) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterAuthApplicationPassword)
	}

	httpRequestBuilderProvider := http_requestbuilder.NewHTTPRequestBuilderProvider()
	httpClient := http_client_builder.New().WithoutProxy().Build()
	authClient := auth_client.New(httpClient.Do, httpRequestBuilderProvider, auth_model.Url(*authUrlPtr), auth_model.ApplicationName(*authApplicationNamePtr), auth_model.ApplicationPassword(*authApplicationPasswordPtr))
	groups := auth_verifier.CreateGroupsFromString(*authGroupsPtr)
	return auth_verifier.New(authClient.Auth, groups...), nil
}

func createLdapVerifier() (verifier.Verifier, error) {
	return ldap.New(), nil
}

func createFileVerifier() (verifier.Verifier, error) {
	if len(*fileUsersPtr) == 0 {
		return nil, fmt.Errorf("parameter %s missing", parameterFileUsers)
	}
	return file.New(file.UserFile(*fileUsersPtr)), nil
}
