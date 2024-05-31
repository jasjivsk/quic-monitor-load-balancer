package util

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func BuildTLSClientConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"quic-echo-example"},
	}
}

func BuildTLSClientConfigWithCert(certFile string) (*tls.Config, error) {
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		log.Println("[client] error reading server certificate:", err)
		return nil, fmt.Errorf("error reading server certificate: %w", err)
	}

	// Parse the certificate
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Create a tls.Config object with the server's certificate
	return &tls.Config{
		RootCAs:    caCertPool,
		NextProtos: []string{"quic-echo-example"},
	}, nil
}

func BuildTLSConfig(cert string, key string) (*tls.Config, error) {
	tlsCert, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}, nil
}

func GenerateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"quic-echo-example"},
	}, nil
}

func GenerateJWT(clientID string) string {
	// Create a new JWT token with the client ID as a claim
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"client_id": clientID,
		"exp":       time.Now().Add(time.Hour * 24).Unix(),
	})

	// Sign the token with a secret key (replace with your own secret)
	secretKey := []byte("your_secret_key")
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		log.Printf("Error generating JWT: %v", err)
		return ""
	}

	return tokenString
}

func VerifyJWT(tokenString string) (string, error) {
	// Parse the JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method and return the secret key
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		secretKey := []byte("your_secret_key")
		return secretKey, nil
	})

	if err != nil {
		return "", err
	}

	// Extract the client ID from the token claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		clientID := claims["client_id"].(string)
		return clientID, nil
	}

	return "", fmt.Errorf("invalid JWT token")
}
