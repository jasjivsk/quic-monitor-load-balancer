package server

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"drexel.edu/net-quic/pkg/pdu"
	"drexel.edu/net-quic/pkg/util"
	"github.com/quic-go/quic-go"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

type ServerConfig struct {
	GenTLS   bool
	CertFile string
	KeyFile  string
	Address  string
	Port     int
}

type Server struct {
	cfg ServerConfig
	tls *tls.Config
	ctx context.Context
}

func NewServer(cfg ServerConfig) *Server {
	server := &Server{
		cfg: cfg,
	}
	server.tls = server.getTLS()
	server.ctx = context.TODO()
	return server
}

func (s *Server) getTLS() *tls.Config {
	if s.cfg.GenTLS {
		tlsConfig, err := util.GenerateTLSConfig()
		if err != nil {
			log.Fatal(err)
		}
		return tlsConfig
	} else {
		tlsConfig, err := util.BuildTLSConfig(s.cfg.CertFile, s.cfg.KeyFile)
		if err != nil {
			log.Fatal(err)
		}
		return tlsConfig
	}
}
func (s *Server) Run() error {
	address := fmt.Sprintf("%s:%d", s.cfg.Address, s.cfg.Port)
	listener, err := quic.ListenAddr(address, s.tls, nil)
	if err != nil {
		log.Printf("error listening: %s", err)
		return err
	}
	log.Printf("[server] Listening on %s", address)
	//SERVER LOOP
	for {
		log.Println("[server] Waiting for loadbalancer to connect..")
		sess, err := listener.Accept(s.ctx)
		if err != nil {
			log.Printf("error accepting: %s", err)
			return err
		}

		go s.streamHandler(sess)
	}
}
func (s *Server) streamHandler(sess quic.Connection) {
	for {
		log.Print("[server] waiting for client to open stream")
		stream, err := sess.AcceptStream(s.ctx)
		if err != nil {
			log.Printf("[server] stream closed: %s", err)
			break
		}
		//Handle protocol activity on stream
		s.protocolHandler(stream)
	}
}
func (s *Server) protocolHandler(stream quic.Stream) error {
	//THIS IS WHERE YOU START HANDLING YOUR APP PROTOCOL
	buff := pdu.MakePduBuffer()
	for {
		n, err := stream.Read(buff)
		if err != nil {
			log.Printf("[server] Error Reading Raw Data: %s", err)
			return err
		}

		data, err := pdu.PduFromBytes(buff[:n])
		if err != nil {
			log.Printf("[server] Error decoding PDU: %s", err)
			return err
		}

		log.Printf("[server] Data In: [%s] %s",
			data.GetTypeAsString(), string(data.Data))

		switch data.Mtype {
		case pdu.TYPE_HELLO:
			// Process HELLO message and send ACK
			var hello struct {
				SupportedMetrics []string `json:"supported_metrics"`
				CheckInterval    int      `json:"check_interval"`
				AuthToken        string   `json:"auth_token"`
				Version          float64  `json:"version"`
			}
			json.Unmarshal(data.Data, &hello)
			// Verify JWT token (simplified, just trust the token)
			serverID := fmt.Sprintf("server-%d", s.cfg.Port)
			// Send ACK
			ackData := map[string]interface{}{
				"confirmed_metrics": hello.SupportedMetrics,
				"check_interval":    hello.CheckInterval,
				"server_id":         serverID,
			}
			ackBytes, _ := json.Marshal(ackData)
			ackPdu := pdu.PDU{
				Mtype:  pdu.TYPE_ACK,
				Length: uint16(len(ackBytes)),
				Data:   ackBytes,
			}
			ackBytes, _ = pdu.PduToBytes(&ackPdu)
			stream.Write(ackBytes)

		case pdu.TYPE_HEALTH_REQUEST:
			// Send current health metrics
			healthData := s.getHealthData()
			rspPdu := pdu.PDU{
				Mtype:  pdu.TYPE_HEALTH_RESPONSE,
				Length: uint16(len(healthData)),
				Data:   healthData,
			}
			rspBytes, _ := pdu.PduToBytes(&rspPdu)
			_, err := stream.Write(rspBytes)
			if err != nil {
				log.Printf("[server] Error sending health response: %s", err)
				return err
			}
			log.Println("[server] Sent health response")

		case pdu.TYPE_CONFIG_UPDATE:
			// Update health check configuration
			var configUpdate struct {
				NewMetrics       []string `json:"new_metrics"`
				NewCheckInterval int      `json:"new_check_interval"`
			}
			json.Unmarshal(data.Data, &configUpdate)
			s.updateHealthCheckConfig(configUpdate.NewMetrics, configUpdate.NewCheckInterval)
			// Send CONFIG_ACK
			ackData := map[string]interface{}{
				"update_status": "success",
				"message":       "Configuration updated successfully.",
			}
			ackBytes, _ := json.Marshal(ackData)
			ackPdu := pdu.PDU{
				Mtype:  pdu.TYPE_CONFIG_ACK,
				Length: uint16(len(ackBytes)),
				Data:   ackBytes,
			}
			ackBytes, _ = pdu.PduToBytes(&ackPdu)
			stream.Write(ackBytes)

		case pdu.TYPE_TERMINATE:
			// Acknowledge termination and close the stream
			ackData := map[string]interface{}{
				"message": "Session terminated successfully.",
			}
			ackBytes, _ := json.Marshal(ackData)
			ackPdu := pdu.PDU{
				Mtype:  pdu.TYPE_TERMINATE_ACK,
				Length: uint16(len(ackBytes)),
				Data:   ackBytes,
			}
			ackBytes, _ = pdu.PduToBytes(&ackPdu)
			stream.Write(ackBytes)
			return nil

		default:
			// Handle unknown message types
			errorData := map[string]interface{}{
				"error_code":    404,
				"error_message": "Unknown message type.",
			}
			errorBytes, _ := json.Marshal(errorData)
			errorPdu := pdu.PDU{
				Mtype:  pdu.TYPE_ERROR,
				Length: uint16(len(errorBytes)),
				Data:   errorBytes,
			}
			errorBytes, _ = pdu.PduToBytes(&errorPdu)
			_, err = stream.Write(errorBytes)
			if err != nil {
				log.Printf("[server] Error sending error response: %s", err)
				return err
			}
			return nil
		}
	}
}
func (s *Server) getHealthData() []byte {
	// Get the current CPU usage percentage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		log.Printf("[server] Error getting CPU usage: %v", err)
		cpuPercent = []float64{0.0}
	}
	// Get the current memory usage percentage
	memStat, err := mem.VirtualMemory()
	if err != nil {
		log.Printf("[server] Error getting memory usage: %v", err)
		memStat = &mem.VirtualMemoryStat{}
	}
	memPercent := memStat.UsedPercent

	healthData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"metrics": map[string]float64{
			"cpu_usage_percent":    cpuPercent[0],
			"memory_usage_percent": memPercent,
		},
	}
	jsonData, _ := json.Marshal(healthData)
	return jsonData
}
func (s *Server) updateHealthCheckConfig(newMetrics []string, newCheckInterval int) {
	// Update health check configuration with the new metrics and interval
	log.Printf("[server] Updated health check configuration: metrics=%v, interval=%d", newMetrics, newCheckInterval)
	// Implement the logic to update the server's health check configuration
}
