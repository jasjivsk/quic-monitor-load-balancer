package loadbalancer

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
)

type LoadBalancerConfig struct {
	Servers           []string
	CertFile          string
	MaxFailAttempts   int
	CheckInterval     int
	ReconnectInterval int
	Port              int
}
type LoadBalancer struct {
	cfg                LoadBalancerConfig
	tls                *tls.Config
	ctx                context.Context
	serverHealthMap    map[string]*ServerHealth
	serverFailureCount map[string]int
	mu                 sync.Mutex
}
type ServerHealth struct {
	ServerID        string
	IsHealthy       bool
	FailedAttempts  int
	MaxFailAttempts int
	conn            quic.Connection
}

func NewLoadBalancer(cfg LoadBalancerConfig) *LoadBalancer {
	lb := &LoadBalancer{
		cfg:                cfg,
		serverHealthMap:    make(map[string]*ServerHealth),
		serverFailureCount: make(map[string]int),
	}
	if cfg.CertFile != "" {
		log.Printf("[loadbalancer] using cert file: %s", cfg.CertFile)
		t, err := util.BuildTLSClientConfigWithCert(cfg.CertFile)
		if err != nil {
			log.Fatal("[loadbalancer] error building TLS client config:", err)
			return nil
		}
		lb.tls = t
	} else {
		lb.tls = util.BuildTLSClientConfig()
	}
	lb.ctx = context.TODO()
	return lb
}
func (lb *LoadBalancer) Run() error {
	// Connect to each server and start health check
	for _, serverAddr := range lb.cfg.Servers {
		go lb.connectAndMonitor(serverAddr)
	}
	// Periodically check the health status of servers
	healthCheckTicker := time.NewTicker(time.Duration(lb.cfg.CheckInterval) * time.Second)
	defer healthCheckTicker.Stop()

	// Ticker for displaying the count of healthy servers
	statusTicker := time.NewTicker(time.Duration(lb.cfg.CheckInterval) * time.Second)
	defer statusTicker.Stop()

	// Ticker for prompting user to update configuration
	updateConfigTicker := time.NewTicker(15 * time.Second)
	defer updateConfigTicker.Stop()

	// Ticker for attempting reconnection to down servers
	reconnectTicker := time.NewTicker(time.Duration(lb.cfg.ReconnectInterval) * time.Second)
	defer reconnectTicker.Stop()

	go func() {
		for range statusTicker.C {
			lb.displayHealthStatus()
		}
	}()

	for {
		select {
		case <-healthCheckTicker.C:
			lb.performHealthCheck()

		case <-updateConfigTicker.C:
			lb.promptUpdateConfiguration()

		case <-reconnectTicker.C:
			lb.reconnectDownServers()
		}
	}
}
func (lb *LoadBalancer) connectAndMonitor(serverAddr string) {
	conn, err := quic.DialAddr(lb.ctx, serverAddr, lb.tls, nil)
	lb.mu.Lock()
	if err != nil {
		log.Printf("[loadbalancer] error dialing server %s: %v", serverAddr, err)
		lb.serverFailureCount[serverAddr]++
		lb.mu.Unlock()
		return
	}
	serverID := lb.protocolHandler(conn)
	if serverID == "" {
		log.Printf("[loadbalancer] failed to get server ID for %s", serverAddr)
		lb.serverFailureCount[serverAddr]++
		lb.mu.Unlock()
		return
	}

	lb.serverHealthMap[serverID] = &ServerHealth{
		ServerID:        serverID,
		IsHealthy:       true,
		FailedAttempts:  0,
		MaxFailAttempts: lb.cfg.MaxFailAttempts,
		conn:            conn,
	}
	delete(lb.serverFailureCount, serverAddr)
	lb.mu.Unlock()
}
func (lb *LoadBalancer) performHealthCheck() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	for serverID, health := range lb.serverHealthMap {
		if !health.IsHealthy {
			if health.FailedAttempts >= health.MaxFailAttempts {
				log.Printf("[loadbalancer] Server %s is down", serverID)
			} else {
				log.Printf("[loadbalancer] Server %s health check failed (%d/%d)", serverID, health.FailedAttempts, health.MaxFailAttempts)
			}
		}
	}
}
func (lb *LoadBalancer) displayHealthStatus() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	healthyCount := 0
	for _, health := range lb.serverHealthMap {
		if health.IsHealthy {
			healthyCount++
		}
	}
	totalServers := len(lb.cfg.Servers)
	log.Printf("[loadbalancer] %d out of %d servers are healthy", healthyCount, totalServers)
	for serverAddr, failCount := range lb.serverFailureCount {
		log.Printf("[loadbalancer] Server %s failed to connect %d times", serverAddr, failCount)
	}
}
func (lb *LoadBalancer) promptUpdateConfiguration() {
	var updateConfig string
	log.Print("Do you want to update the configuration? (y/n): ")
	fmt.Scanln(&updateConfig)
	if strings.ToLower(updateConfig) == "y" {
		lb.updateConfiguration()
	}
}
func (lb *LoadBalancer) updateConfiguration() {
	var metrics string
	var interval int
	log.Print("Enter the new metrics (comma-separated): ")
	fmt.Scanln(&metrics)
	log.Print("Enter the new check interval (in seconds): ")
	fmt.Scanln(&interval)
	newMetrics := strings.Split(metrics, ",")
	configUpdateData := map[string]interface{}{
		"new_metrics":        newMetrics,
		"new_check_interval": interval,
	}
	configUpdateBytes, _ := json.Marshal(configUpdateData)
	configUpdatePdu := pdu.PDU{
		Mtype:  pdu.TYPE_CONFIG_UPDATE,
		Length: uint16(len(configUpdateBytes)),
		Data:   configUpdateBytes,
	}
	configUpdatePduBytes, _ := pdu.PduToBytes(&configUpdatePdu)

	lb.mu.Lock()
	defer lb.mu.Unlock()
	for serverID, health := range lb.serverHealthMap {
		if health.IsHealthy {
			stream, err := health.conn.OpenStreamSync(lb.ctx)
			if err != nil {
				log.Printf("[loadbalancer] Error opening stream for server %s: %v", serverID, err)
				continue
			}
			_, err = stream.Write(configUpdatePduBytes)
			if err != nil {
				log.Printf("[loadbalancer] Error sending CONFIG_UPDATE to server %s: %v", serverID, err)
				continue
			}
			log.Printf("[loadbalancer] Sent CONFIG_UPDATE to server %s", serverID)
		}
	}
}
func (lb *LoadBalancer) protocolHandler(conn quic.Connection) string {
	stream, err := conn.OpenStreamSync(lb.ctx)
	if err != nil {
		log.Printf("[loadbalancer] error opening stream %s", err)
		return ""
	}
	// Send HELLO PDU
	helloData := map[string]interface{}{
		"supported_metrics": []string{"cpu_load", "memory_usage", "response_time"},
		"check_interval":    5,
		"auth_token":        util.GenerateJWT("loadbalancer123"),
		"version":           1.0,
	}
	helloBytes, _ := json.Marshal(helloData)
	helloPdu := pdu.PDU{
		Mtype:  pdu.TYPE_HELLO,
		Length: uint16(len(helloBytes)),
		Data:   helloBytes,
	}
	pduBytes, _ := pdu.PduToBytes(&helloPdu)
	n, err := stream.Write(pduBytes)
	if err != nil {
		log.Printf("[loadbalancer] error writing to stream %s", err)
		return ""
	}
	log.Printf("[loadbalancer] wrote %d bytes to stream", n)
	// Read the ACK message from the server
	ackBuffer := pdu.MakePduBuffer()
	n, err = stream.Read(ackBuffer)
	if err != nil {
		log.Printf("[loadbalancer] Error reading ACK from stream: %v", err)
		return ""
	}
	ackPdu, err := pdu.PduFromBytes(ackBuffer[:n])
	if err != nil {
		log.Printf("[loadbalancer] Error converting ACK pdu from bytes %s", err)
		return ""
	}
	log.Printf("[loadbalancer] Got ACK response: %s", ackPdu.ToJsonString())

	var ackData struct {
		ServerID string `json:"server_id"`
	}
	json.Unmarshal(ackPdu.Data, &ackData)

	// Periodically send health check requests
	go lb.sendHealthChecks(conn, ackData.ServerID, stream, helloData["check_interval"].(int))

	return ackData.ServerID
}
func (lb *LoadBalancer) sendHealthChecks(conn quic.Connection, serverID string, stream quic.Stream, checkInterval int) {
	ticker := time.NewTicker(time.Duration(checkInterval) * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		// Send health check request
		reqPdu := pdu.PDU{Mtype: pdu.TYPE_HEALTH_REQUEST}
		reqBytes, _ := pdu.PduToBytes(&reqPdu)
		_, err := stream.Write(reqBytes)
		if err != nil {
			log.Printf("[loadbalancer] Error sending health check request to server %s: %v", serverID, err)
			lb.markServerUnhealthy(serverID)
			continue
		}
		log.Printf("[loadbalancer] Sent health check request to server %s", serverID)

		// Read and process server response
		buffer := pdu.MakePduBuffer()
		n, err := stream.Read(buffer)
		if err != nil {
			log.Printf("[loadbalancer] Error reading from stream for server %s: %v", serverID, err)
			lb.markServerUnhealthy(serverID)
			continue
		}

		log.Printf("[loadbalancer] Received PDU bytes from server %s: %v", serverID, buffer[:n])
		rsp, err := pdu.PduFromBytes(buffer[:n])
		if err != nil {
			log.Printf("[loadbalancer] Error converting pdu from bytes for server %s: %s", serverID, err)
			continue
		}
		rspDataString := string(rsp.Data)
		log.Printf("[loadbalancer] Got response from server %s: %s", serverID, rsp.ToJsonString())
		log.Printf("[loadbalancer] Decoded string from server %s: %s", serverID, rspDataString)
		switch rsp.Mtype {
		case pdu.TYPE_HEALTH_RESPONSE:
			var healthData struct {
				Timestamp string             `json:"timestamp"`
				Metrics   map[string]float64 `json:"metrics"`
			}
			json.Unmarshal(rsp.Data, &healthData)
			log.Printf("[loadbalancer] Received health data from server %s: CPU Usage: %.2f%%, Memory Usage: %.2f%%",
				serverID, healthData.Metrics["cpu_usage_percent"], healthData.Metrics["memory_usage_percent"])
			lb.markServerHealthy(serverID)
		case pdu.TYPE_ERROR:
			var errorData struct {
				ErrorCode    int    `json:"error_code"`
				ErrorMessage string `json:"error_message"`
			}
			json.Unmarshal(rsp.Data, &errorData)
			log.Printf("[loadbalancer] Error from server %s: %d - %s", serverID, errorData.ErrorCode, errorData.ErrorMessage)
			lb.markServerUnhealthy(serverID)
		case pdu.TYPE_CONFIG_ACK:
			var configAck struct {
				UpdateStatus string `json:"update_status"`
				Message      string `json:"message"`
			}
			json.Unmarshal(rsp.Data, &configAck)
			log.Printf("[loadbalancer] Configuration update ACK from server %s: %s - %s", serverID, configAck.UpdateStatus, configAck.Message)
		}
	}
}
func (lb *LoadBalancer) markServerHealthy(serverID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if serverHealth, ok := lb.serverHealthMap[serverID]; ok {
		serverHealth.FailedAttempts = 0
		serverHealth.IsHealthy = true
	}
}
func (lb *LoadBalancer) markServerUnhealthy(serverID string) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if serverHealth, ok := lb.serverHealthMap[serverID]; ok {
		serverHealth.FailedAttempts++
		if serverHealth.FailedAttempts >= serverHealth.MaxFailAttempts {
			serverHealth.IsHealthy = false
			log.Printf("[loadbalancer] Server %s is down", serverID)
		}
	}
}
func (lb *LoadBalancer) reconnectDownServers() {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	for serverAddr, failCount := range lb.serverFailureCount {
		go lb.connectAndMonitor(serverAddr)
		log.Printf("[loadbalancer] Attempting to reconnect to server %s (failed %d times)", serverAddr, failCount)
	}
}
