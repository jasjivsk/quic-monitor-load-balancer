package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
)

type ClientConfig struct {
	ServerAddr  string
	ServerPorts []int
	CertFile    string
}
type Client struct {
	cfg             ClientConfig
	tls             *tls.Config
	ctx             context.Context
	serverHealthMap map[string]*ServerHealth
	mu              sync.Mutex
}
type ServerHealth struct {
	ServerID        string
	IsHealthy       bool
	FailedAttempts  int
	MaxFailAttempts int
}

func NewClient(cfg ClientConfig) *Client {
	cli := &Client{
		cfg:             cfg,
		serverHealthMap: make(map[string]*ServerHealth),
	}
	if cfg.CertFile != "" {
		log.Printf("[cli] using cert file: %s", cfg.CertFile)
		t, err := util.BuildTLSClientConfigWithCert(cfg.CertFile)
		if err != nil {
			log.Fatal("[cli] error building TLS client config:", err)
			return nil
		}
		cli.tls = t
	} else {
		cli.tls = util.BuildTLSClientConfig()
	}

	cli.ctx = context.TODO()
	return cli
}
func (c *Client) Run() error {
	// Connect to each server and start health check
	for _, port := range c.cfg.ServerPorts {
		serverAddr := fmt.Sprintf("%s:%d", c.cfg.ServerAddr, port)
		go func(serverAddr string) {
			conn, err := quic.DialAddr(c.ctx, serverAddr, c.tls, nil)
			if err != nil {
				log.Printf("[cli] error dialing server %s: %v", serverAddr, err)
				return
			}
			serverID := c.protocolHandler(conn)
			c.mu.Lock()
			c.serverHealthMap[serverID] = &ServerHealth{
				ServerID:        serverID,
				IsHealthy:       true,
				FailedAttempts:  0,
				MaxFailAttempts: 3, // Set the maximum number of failed attempts before considering the server as down
			}
			c.mu.Unlock()
		}(serverAddr)
	}
	// Periodically check the health status of servers
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		healthyCount := 0
		c.mu.Lock()
		for serverID, health := range c.serverHealthMap {
			if health.IsHealthy {
				healthyCount++
			} else {
				if health.FailedAttempts >= health.MaxFailAttempts {
					log.Printf("[cli] Server %s is down", serverID)
				} else {
					log.Printf("[cli] Server %s health check failed (%d/%d)", serverID, health.FailedAttempts, health.MaxFailAttempts)
				}
			}
		}
		c.mu.Unlock()
		log.Printf("[cli] %d out of %d servers are healthy", healthyCount, len(c.serverHealthMap))
	}

	return nil
}
func (c *Client) protocolHandler(conn quic.Connection) string {
	stream, err := conn.OpenStreamSync(c.ctx)
	if err != nil {
		log.Printf("[cli] error opening stream %s", err)
		return ""
	}
	// Send HELLO PDU
	helloData := map[string]interface{}{
		"supported_metrics": []string{"cpu_load", "memory_usage", "response_time"},
		"check_interval":    5,
		"auth_token":        util.GenerateJWT("client123"),
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
		log.Printf("[cli] error writing to stream %s", err)
		return ""
	}
	log.Printf("[cli] wrote %d bytes to stream", n)

	// Read the ACK message from the server
	ackBuffer := pdu.MakePduBuffer()
	n, err = stream.Read(ackBuffer)
	if err != nil {
		log.Printf("[client] Error reading ACK from stream: %v", err)
		return ""
	}
	ackPdu, err := pdu.PduFromBytes(ackBuffer[:n])
	if err != nil {
		log.Printf("[client] Error converting ACK pdu from bytes %s", err)
		return ""
	}
	log.Printf("[client] Got ACK response: %s", ackPdu.ToJsonString())

	var ackData struct {
		ServerID string `json:"server_id"`
	}
	json.Unmarshal(ackPdu.Data, &ackData)

	// Periodically send health check requests
	ticker := time.NewTicker(time.Duration(helloData["check_interval"].(int)) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Send health check request
		reqPdu := pdu.PDU{Mtype: pdu.TYPE_HEALTH_REQUEST}
		reqBytes, _ := pdu.PduToBytes(&reqPdu)
		_, err := stream.Write(reqBytes)
		if err != nil {
			log.Printf("[client] Error sending health check request to server %s: %v", ackData.ServerID, err)
			c.mu.Lock()
			serverHealth, ok := c.serverHealthMap[ackData.ServerID]
			if ok {
				serverHealth.FailedAttempts++
				serverHealth.IsHealthy = serverHealth.FailedAttempts < serverHealth.MaxFailAttempts
			}
			c.mu.Unlock()
			continue
		}
		log.Printf("[client] Sent health check request to server %s", ackData.ServerID)

		// Read and process server response
		buffer := pdu.MakePduBuffer()
		n, err := stream.Read(buffer)
		if err != nil {
			log.Printf("[client] Error reading from stream for server %s: %v", ackData.ServerID, err)
			c.mu.Lock()
			serverHealth, ok := c.serverHealthMap[ackData.ServerID]
			if ok {
				serverHealth.FailedAttempts++
				serverHealth.IsHealthy = serverHealth.FailedAttempts < serverHealth.MaxFailAttempts
			}
			c.mu.Unlock()
			continue
		}

		log.Printf("[client] Received PDU bytes from server %s: %v", ackData.ServerID, buffer[:n])
		rsp, err := pdu.PduFromBytes(buffer[:n])
		if err != nil {
			log.Printf("[client] Error converting pdu from bytes for server %s: %s", ackData.ServerID, err)
			continue
		}
		rspDataString := string(rsp.Data)
		log.Printf("[client] Got response from server %s: %s", ackData.ServerID, rsp.ToJsonString())
		log.Printf("[client] Decoded string from server %s: %s", ackData.ServerID, rspDataString)
		switch rsp.Mtype {
		case pdu.TYPE_HEALTH_RESPONSE:
			var healthData struct {
				Timestamp string                 `json:"timestamp"`
				Metrics   map[string]interface{} `json:"metrics"`
			}
			json.Unmarshal(rsp.Data, &healthData)
			log.Printf("[client] Received health data from server %s: %+v", ackData.ServerID, healthData)
			c.mu.Lock()
			serverHealth, ok := c.serverHealthMap[ackData.ServerID]
			if ok {
				serverHealth.FailedAttempts = 0
				serverHealth.IsHealthy = true
			}
			c.mu.Unlock()
		case pdu.TYPE_ERROR:
			var errorData struct {
				ErrorCode    int    `json:"error_code"`
				ErrorMessage string `json:"error_message"`
			}
			json.Unmarshal(rsp.Data, &errorData)
			log.Printf("[client] Error from server %s: %d - %s", ackData.ServerID, errorData.ErrorCode, errorData.ErrorMessage)
			c.mu.Lock()
			serverHealth, ok := c.serverHealthMap[ackData.ServerID]
			if ok {
				serverHealth.FailedAttempts++
				serverHealth.IsHealthy = serverHealth.FailedAttempts < serverHealth.MaxFailAttempts
			}
			c.mu.Unlock()
		}
	}

	return ackData.ServerID
}
