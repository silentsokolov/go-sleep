package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	auth "github.com/abbot/go-http-auth"

	"github.com/silentsokolov/go-sleep/log"
	"github.com/silentsokolov/go-sleep/provider"
)

// Server ...
type Server struct {
	InstanceStore *InstanceStore
	templates     *template.Template
	stopChan      chan bool
	signals       chan os.Signal
	portWeb       string
	secretKey     string
	serverRoutes  map[string]map[string]*serverRoute
}

type serverRoute struct {
	Hostname     string
	BackendPort  int
	InstanceName string
	basicUsers   map[string]string
	basicAuth    *auth.BasicAuth
	Certificates []tls.Certificate
}

type pageContext struct {
	Message      string     `json:"message,omitempty"`
	StartRequest *time.Time `json:"request_start_at,omitempty"`
	Error        string     `json:"error,omitempty"`
}

func (route *serverRoute) secretBasic(user, realm string) string {
	if secret, ok := route.basicUsers[user]; ok {
		return secret
	}

	log.Printf("User not found: %s", user)

	return ""
}

// NewServer ...
func NewServer(conf *Config) *Server {
	server := new(Server)

	server.InstanceStore = NewInstanceStore()
	server.stopChan = make(chan bool, 1)
	server.signals = make(chan os.Signal, 1)
	server.portWeb = conf.Port
	server.serverRoutes = make(map[string]map[string]*serverRoute)
	signal.Notify(server.signals, syscall.SIGINT, syscall.SIGTERM)

	return server
}

// Start ...
func (server *Server) Start() {
	server.startServers()
	go startWebServer(server.portWeb)
	go server.listenSignals()
}

// Wait ...
func (server *Server) Wait() {
	<-server.stopChan
}

// Close ...
func (server *Server) Close() {
	server.InstanceStore.Close()
	signal.Stop(server.signals)
	close(server.signals)
	close(server.stopChan)
}

// Stop ...
func (server *Server) Stop() {
	server.stopChan <- true
}

func (server *Server) loadConfig(config *Config) {
	// TODO Hot-reload config
	var (
		instance     *ComputeInstance
		sleep        time.Duration
		instanceHash string
		err          error
	)

	server.secretKey = config.SecretKey

	serverBasicAuthUsers := make(map[string]map[string]string)
	for groupName, group := range config.AuthBasic {
		serverBasicAuthUsers[groupName], err = parserBasicUsers(group.Users)
		if err != nil {
			log.Fatal(err)
		}
	}

	for _, conf := range config.EC2 {
		log.Println("Initialization EC2 instances ...")
		sleep = sleepDuration(conf.SleepAfter)
		instance = NewComputeInstance(provider.NewEC2(conf.AccessKeyID, conf.SecretAccessKey, conf.Region, conf.InstanceID, conf.UseInternalIP), sleep)
		instanceHash = instance.Hash()
		server.InstanceStore.Set(instanceHash, instance)

		log.Printf("Found... %s", instance.String())
		server.buildServerRoutes(conf.Routes, instanceHash, serverBasicAuthUsers)
	}

	for _, conf := range config.GCE {
		log.Println("Initialization GCE instances ...")
		sleep = sleepDuration(conf.SleepAfter)
		instance = NewComputeInstance(provider.NewGCE(conf.JWTPath, conf.ProjectID, conf.Zone, conf.Name, conf.UseInternalIP), sleep)
		instanceHash = instance.Hash()
		server.InstanceStore.Set(instanceHash, instance)

		log.Printf("Found... %s", instance.String())
		server.buildServerRoutes(conf.Routes, instanceHash, serverBasicAuthUsers)
	}
}

func (server *Server) buildServerRoutes(routes []*RouteConfig, instanceKey string, authUsers map[string]map[string]string) {
	var err error

	for _, route := range routes {
		for _, name := range route.Hostnames {
			// Set default address if not set
			if len(route.Address) == 0 {
				route.Address = defaultAddress
			}
			// Set default backend port if not set
			if route.BackendPort == 0 {
				route.BackendPort, err = strconv.Atoi(strings.Replace(route.Address, ":", "", -1))
				if err != nil {
					log.Fatal(err)
				}
			}

			if _, ok := server.serverRoutes[route.Address]; !ok {
				server.serverRoutes[route.Address] = make(map[string]*serverRoute)
			}

			srvRoute := serverRoute{
				Hostname:     name,
				BackendPort:  route.BackendPort,
				InstanceName: instanceKey,
			}
			// Init and add cret
			for _, cretOptions := range route.Certificates {
				cert, err := tls.LoadX509KeyPair(cretOptions.CertFile, cretOptions.KeyFile)
				if err != nil {
					log.Fatal("Error load certificate: ", err)
				}
				srvRoute.Certificates = append(srvRoute.Certificates, cert)
			}

			if users, ok := authUsers[route.AuthGroup]; ok {
				srvRoute.basicUsers = users
				srvRoute.basicAuth = auth.NewBasicAuthenticator("go-sleep", srvRoute.secretBasic)
			}

			server.serverRoutes[route.Address][name] = &srvRoute
		}
	}
}

func (server *Server) listenSignals() {
	<-server.signals
	// FIXME graceful shutdown stopping
	log.Println("Server stopping ...")
	server.Stop()
}

func (server *Server) createTLSConfig(routes map[string]*serverRoute) *tls.Config {
	crets := []tls.Certificate{}

	for _, route := range routes {
		if len(route.Certificates) != 0 {
			crets = append(crets, route.Certificates...)
		}
	}

	if len(crets) != 0 {
		cfg := tls.Config{}
		cfg.Certificates = crets
		cfg.NextProtos = []string{"h2", "http/1.1"}
		cfg.BuildNameToCertificate()

		return &cfg
	}
	return nil
}

func (server *Server) startServers() {
	for addr, routes := range server.serverRoutes {
		tlsConfig := server.createTLSConfig(routes)
		handler := server.middlewareAuth(server.middlewareWakeup(server.defaultReverseProxy(addr), addr), addr)

		srv := &http.Server{
			Addr:      addr,
			Handler:   handler,
			TLSConfig: tlsConfig,
		}

		go server.startServer(srv)
	}
}

func (server *Server) startServer(srv *http.Server) {
	if srv.TLSConfig != nil {
		log.Printf("Starting server on %s with TLS", srv.Addr)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			log.Fatal("Error creating server with TLS: ", err)
		}
	} else {
		log.Printf("Starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal("Error creating server: ", err)
		}
	}
}

func (server *Server) defaultReverseProxy(address string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			route, computer, err := server.routeComputer(r.Host, address)
			if err == nil {
				r.Header.Set("Host", r.Host)
				r.Header.Set("X-Go-Sleep-Key", server.secretKey)
				r.URL.Scheme = "http"
				r.URL.Host = fmt.Sprintf("%s:%d", computer.IP, route.BackendPort)
				r.RequestURI = ""
			} else {
				log.Warnf("%q is not routed", r.Host)
			}
		},
	}
}

func (server *Server) middlewareAuth(next http.Handler, address string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		route, ok := server.serverRoutes[address][host]
		if ok && route.basicAuth != nil {
			if username := route.basicAuth.CheckAuth(r); username == "" {
				log.Printf("Basic auth failed...")
				route.basicAuth.RequireAuth(w, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (server *Server) middlewareWakeup(next http.Handler, address string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		context := pageContext{}

		_, computer, err := server.routeComputer(r.Host, address)
		if err != nil {
			context.Error = err.Error()
			responseJSON(w, context, http.StatusInternalServerError)
			return
		}

		if computer.lastError != nil {
			context.Error = computer.lastError.Error()
			responseJSON(w, context, http.StatusOK)
			return
		}

		switch computer.Status() {
		case provider.StatusInstanceRunning:
			computer.SetLastAccess()
			next.ServeHTTP(w, r)
			return
		case provider.StatusInstanceNotRun:
			computer.Start()
			context.Message = "We sent a request to start the instance"
		case provider.StatusInstanceStarting:
			context.Message = "Waiting for the server to start"
			context.StartRequest = &computer.startRequest
		case provider.StatusInstanceError:
			computer.Start()
			context.Error = computer.lastError.Error()
		case provider.StatusInstanceStopping:
			context.Message = "The server is stopped, we will launch it later"
		}

		responseJSON(w, context, http.StatusOK)
	})
}

func (server *Server) routeComputer(rawHost, address string) (*serverRoute, *ComputeInstance, error) {
	host, _, err := net.SplitHostPort(rawHost)
	if err != nil {
		return nil, nil, err
	}
	route, ok := server.serverRoutes[address][host]
	if !ok {
		return nil, nil, fmt.Errorf("Not found hostname: %s", host)
	}
	computer, ok := server.InstanceStore.Get(route.InstanceName)
	if !ok {
		return nil, nil, fmt.Errorf("Not found instance for hostname: %s", host)
	}
	return route, computer, nil
}

func parserBasicUsers(users []string) (map[string]string, error) {
	userMap := make(map[string]string)
	for _, user := range users {
		split := strings.SplitN(user, ":", 2)
		if len(split) != 2 {
			return nil, fmt.Errorf("Error parsing auth user: %v", user)
		}
		userMap[split[0]] = split[1]
	}
	return userMap, nil
}

func responseJSON(w http.ResponseWriter, context interface{}, status int) {
	js, err := json.Marshal(context)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func sleepDuration(currentSleep int64) time.Duration {
	if currentSleep > 0 {
		return time.Duration(currentSleep) * time.Second
	} else if currentSleep < 0 {
		return time.Duration(math.MaxInt64)
	}
	return defaultSleepAfter
}
