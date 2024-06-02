# Quick Health Check Protocol (QHCP) using QUIC

This project implements a Quick Health Check Protocol (QHCP) using the QUIC transport protocol. The QHCP is designed to monitor the health of multiple backend servers and facilitate load balancing decisions. The solution consists of a load balancer and server components.

## Features

- **Health Monitoring**: The load balancer periodically sends health check requests to the backend servers and monitors their health status based on the responses.
- **Dynamic Configuration**: The load balancer allows dynamic adjustment of health check parameters during runtime. It supports configuration updates for metrics and check intervals.
- **Server Reconnection**: If a server goes down, the load balancer attempts to reconnect to the server at regular intervals defined by the reconnect interval.
- **Metrics Collection**: The server collects and sends back health metrics such as CPU usage percentage and memory usage percentage to the load balancer.
- **Authentication**: The protocol incorporates JSON Web Tokens (JWT) for authentication between the load balancer and servers.
- **Protocol Messaging**: The protocol defines various message types for communication between the load balancer and servers, including HELLO, ACK, HEALTH_REQUEST, HEALTH_RESPONSE, CONFIG_UPDATE, CONFIG_ACK, ERROR, TERMINATE, and TERMINATE_ACK.
- **Secure Communication**: The protocol utilizes QUIC's built-in encryption for secure data transmission between the load balancer and servers.

## Usage

The project provides a single binary that can be used to run both the load balancer and the server.

### Load Balancer

To start the load balancer, use the following command:

 - go run main.go -loadbalancer -servers "server1:port1,server2:port2,server3:port3" [options]

 Replace `server1:port1`, `server2:port2`, and `server3:port3` with the respective server addresses and port numbers. If the port number is not specified, the default port 4243 will be used.

Available options for the load balancer:

- `-loadbalancer-port`: Port number for the load balancer (default: 4242)
- `-max-fail-attempts`: Maximum number of failed attempts before marking a server as down (default: 3)
- `-check-interval`: Interval for health checks and status display in seconds (default: 10)
- `-reconnect-interval`: Interval for attempting to reconnect to down servers in seconds (default: 15)
- `-cert-file`: TLS certificate file

### Server

To start a server, use the following command:

- go run main.go -server -server-port <port_number> [options]

Replace `<port_number>` with the desired port number for the server.

Available options for the server:

- `-server-ip`: Server IP address (default: 0.0.0.0)
- `-key-file`: TLS key file
- `-cert-file`: TLS certificate file

## Protocol Description

The QHCP protocol follows a client-server model, where the load balancer acts as the client and the backend servers act as servers. The load balancer establishes QUIC connections with the servers and periodically sends health check requests. The servers respond with their current health metrics, and the load balancer updates the health status of each server based on the responses.

The protocol supports dynamic configuration updates, allowing the load balancer to modify the health check parameters during runtime. If a server goes down, the load balancer attempts to reconnect to the server at regular intervals.

The protocol uses various message types for communication, including HELLO for initial handshake, ACK for acknowledgments, HEALTH_REQUEST and HEALTH_RESPONSE for health checks, CONFIG_UPDATE and CONFIG_ACK for configuration updates, ERROR for error handling, and TERMINATE and TERMINATE_ACK for session termination.

The protocol messages are encoded using JSON format and transmitted securely using QUIC's built-in encryption.

## Customization

The project can be customized and extended based on specific requirements. The code provides a foundation for implementing a health check protocol using QUIC, and it can be modified to include additional features, metrics, or customizations as needed.

## Dependencies

The project relies on the following dependencies:

- `github.com/quic-go/quic-go`: QUIC transport protocol implementation for Go
- `github.com/shirou/gopsutil`: Library for retrieving system information and metrics

Make sure to install the dependencies before running the project.

## Limitations

- The current implementation assumes a simple authentication mechanism using JSON Web Tokens (JWT). In a production environment, a more robust authentication and authorization system should be considered.
- The code provided is a starting point and may require further optimizations, error handling, and testing before deployment in a production environment.
- The project focuses on the core functionality of the health check protocol and does not include advanced features such as load balancing algorithms or distributed coordination among multiple load balancers.