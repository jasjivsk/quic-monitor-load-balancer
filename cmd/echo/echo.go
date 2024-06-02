package main

import (
	"flag"
	"fmt"
	"strings"

	"drexel.edu/net-quic/pkg/loadbalancer"
	"drexel.edu/net-quic/pkg/server"
)

var (
	// GENERAL PARAMETERS
	GENERATE_TLS      = true
	MODE_LOADBALANCER = false
	MODE_SERVER       = false
	CERT_FILE         = ""
	// SERVER PARAMETERS
	SERVER_IP   = "0.0.0.0"
	SERVER_PORT = 4243
	KEY_FILE    = ""

	// LOADBALANCER PARAMETERS
	SERVERS            = ""
	LOADBALANCER_PORT  = 4242
	MAX_FAIL_ATTEMPTS  = 3
	CHECK_INTERVAL     = 10
	RECONNECT_INTERVAL = 30
)

func processFlags() {
	lbMode := flag.Bool("loadbalancer", MODE_LOADBALANCER, "loadbalancer mode")
	svrMode := flag.Bool("server", MODE_SERVER, "server mode")
	tlsMode := flag.Bool("tls-gen", GENERATE_TLS, "generate tls config")
	flag.StringVar(&CERT_FILE, "cert-file", CERT_FILE, "tls certificate file")
	flag.StringVar(&KEY_FILE, "key-file", KEY_FILE, "[server mode] tls key file")
	flag.StringVar(&SERVER_IP, "server-ip", SERVER_IP, "[server mode] server IP")
	flag.IntVar(&SERVER_PORT, "server-port", SERVER_PORT, "[server mode] server port")
	flag.StringVar(&SERVERS, "servers", SERVERS, "[loadbalancer mode] comma-separated list of server addresses (host:port)")

	flag.IntVar(&LOADBALANCER_PORT, "loadbalancer-port", LOADBALANCER_PORT, "[loadbalancer mode] port for the loadbalancer")
	flag.IntVar(&MAX_FAIL_ATTEMPTS, "max-fail-attempts", MAX_FAIL_ATTEMPTS, "[loadbalancer mode] maximum fail attempts before marking server as down")
	flag.IntVar(&CHECK_INTERVAL, "check-interval", CHECK_INTERVAL, "[loadbalancer mode] interval for health checks and status display in seconds")
	flag.IntVar(&RECONNECT_INTERVAL, "reconnect-interval", RECONNECT_INTERVAL, "[loadbalancer mode] interval for attempting to reconnect to down servers in seconds")

	flag.Parse()
	MODE_LOADBALANCER = *lbMode
	MODE_SERVER = *svrMode
	GENERATE_TLS = *tlsMode

	if !MODE_SERVER {
		// If the server wasn't selected, let's make the loadbalancer the default
		MODE_LOADBALANCER = true
	}
}
func main() {
	processFlags()
	if MODE_LOADBALANCER {
		serverList := make([]string, 0)
		for _, serverAddr := range strings.Split(SERVERS, ",") {
			if !strings.Contains(serverAddr, ":") {
				serverAddr = fmt.Sprintf("%s:%d", serverAddr, 4243)
			}
			serverList = append(serverList, serverAddr)
		}

		lbConfig := loadbalancer.LoadBalancerConfig{
			Servers:           serverList,
			CertFile:          CERT_FILE,
			MaxFailAttempts:   MAX_FAIL_ATTEMPTS,
			CheckInterval:     CHECK_INTERVAL,
			ReconnectInterval: RECONNECT_INTERVAL,
			Port:              LOADBALANCER_PORT,
		}
		lb := loadbalancer.NewLoadBalancer(lbConfig)
		lb.Run()
	} else {
		serverConfig := server.ServerConfig{
			GenTLS:   GENERATE_TLS,
			CertFile: CERT_FILE,
			KeyFile:  KEY_FILE,
			Address:  SERVER_IP,
			Port:     SERVER_PORT,
		}

		server := server.NewServer(serverConfig)
		server.Run()
	}
}
