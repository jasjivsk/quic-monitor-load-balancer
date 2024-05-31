package client

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
)

type ClientConfig struct {
	ServerAddr string
	PortNumber int
	CertFile   string
}

type Client struct {
	cfg  ClientConfig
	tls  *tls.Config
	conn quic.Connection
	ctx  context.Context
}

func NewClient(cfg ClientConfig) *Client {
	cli := &Client{
		cfg: cfg,
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
	serverAddr := fmt.Sprintf("%s:%d", c.cfg.ServerAddr, c.cfg.PortNumber)
	conn, err := quic.DialAddr(c.ctx, serverAddr, c.tls, nil)
	if err != nil {
		log.Printf("[cli] error dialing server %s", err)
		return err
	}
	c.conn = conn
	return c.protocolHandler()
}

func (c *Client) protocolHandler() error {
	stream, err := c.conn.OpenStreamSync(c.ctx)
	if err != nil {
		log.Printf("[cli] error opening stream %s", err)
		return err
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
		return err
	}
	log.Printf("[cli] wrote %d bytes to stream", n)

	// Periodically send health check requests
	ticker := time.NewTicker(time.Duration(helloData["check_interval"].(int)) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send health check request
			reqPdu := pdu.PDU{Mtype: pdu.TYPE_HEALTH_REQUEST}
			reqBytes, _ := pdu.PduToBytes(&reqPdu)
			stream.Write(reqBytes)

		default:
			// Read and process server responses
			buffer := pdu.MakePduBuffer()
			n, err := stream.Read(buffer)
			if err != nil {
				log.Printf("[client] Error reading from stream: %v", err)
				return err
			}

			log.Printf("[client] Received PDU bytes: %v", buffer[:n])
			rsp, err := pdu.PduFromBytes(buffer[:n])
			if err != nil {
				log.Printf("[client] Error converting pdu from bytes %s", err)
				return err
			}
			rspDataString := string(rsp.Data)
			log.Printf("[client] Got response: %s", rsp.ToJsonString())
			log.Printf("[client] Decoded string: %s", rspDataString)
			switch rsp.Mtype {
			case pdu.TYPE_HEALTH_RESPONSE:
				var healthData struct {
					Timestamp string                 `json:"timestamp"`
					Metrics   map[string]interface{} `json:"metrics"`
				}
				json.Unmarshal(rsp.Data, &healthData)
				log.Printf("[client] Received health data: %+v", healthData)
			case pdu.TYPE_CONFIG_ACK:
				var configAck struct {
					UpdateStatus string `json:"update_status"`
					Message      string `json:"message"`
				}
				json.Unmarshal(rsp.Data, &configAck)
				log.Printf("[client] Configuration update: %s - %s", configAck.UpdateStatus, configAck.Message)

			case pdu.TYPE_TERMINATE_ACK:
				var terminateAck struct {
					Message string `json:"message"`
				}
				json.Unmarshal(rsp.Data, &terminateAck)
				log.Printf("[client] Session terminated: %s", terminateAck.Message)
				return nil

			case pdu.TYPE_ERROR:
				var errorData struct {
					ErrorCode    int    `json:"error_code"`
					ErrorMessage string `json:"error_message"`
				}
				json.Unmarshal(rsp.Data, &errorData)
				log.Printf("[client] Error: %d - %s", errorData.ErrorCode, errorData.ErrorMessage)
			}
		}
	}
}
