# Quick Health Check Protocol (QHCP) using QUIC
This project implements a Quick Health Check Protocol (QHCP) using the QUIC transport protocol. The QHCP is designed to monitor the health of multiple backend servers and facilitate load balancing decisions. The solution consists of a load balancer and server components.

## Protocol Design
The Quick Health Check Protocol (QHCP) is a stateful protocol that follows a specific Deterministic Finite Automaton (DFA) for its operation. Both the client (load balancer) and server must implement and validate the statefulness of the protocol according to the defined DFA.

The protocol operates on a hardcoded port number chosen by the user, which is documented in the protocol design. The client defaults to this port number when connecting to the server.

## Usage
The project provides a single binary that can be used to run both the load balancer and the server(s). Configuration information for the client and server is provided via command line arguments.

### Load Balancer
To start the load balancer, use the following command:
go run echo.go -loadbalancer 
-servers "server1:port,server2:port" 
-loadbalancer-port 4242 
-max-fail-attempts 3 
-check-interval 10 
-reconnect-interval 30 
-cert-file "path/to/cert/file"
Copy code
The load balancer accepts the following command line arguments:
- `-servers`: Comma-separated list of server addresses in the format `host:port`.
- `-loadbalancer-port`: Port number for the load balancer to listen on (default: 4242).
- `-max-fail-attempts`: Maximum number of failed health check attempts before marking a server as down (default: 3).
- `-check-interval`: Interval in seconds for performing health checks and displaying server status (default: 10).
- `-reconnect-interval`: Interval in seconds for attempting to reconnect to down servers (default: 30).
- `-cert-file`: Path to the TLS certificate file (optional).

### Server
To start a server, use the following command:
go run echo.go -server 
-server-ip "0.0.0.0" 
-server-port 4243 
-key-file "path/to/key/file" 
-cert-file "path/to/cert/file"
Copy code
The server accepts the following command line arguments:
- `-server-ip`: IP address for the server to listen on (default: "0.0.0.0").
- `-server-port`: Port number for the server to listen on (default: 4243).
- `-key-file`: Path to the TLS key file (optional).
- `-cert-file`: Path to the TLS certificate file (optional).

## Customization and Extensibility
The project provides a workable prototype implementation of the Quick Health Check Protocol using QUIC. It can be customized and extended based on specific requirements. The code serves as a foundation for implementing a health check protocol and can be modified to include additional features, metrics, or customizations as needed.

## Concurrency and Asynchronous Server (Extra Credit)
The current server implementation does not support concurrent handling of multiple clients. As an extra credit feature, the server can be enhanced to handle multiple clients concurrently using a process/thread/coroutine model. This allows the server to efficiently serve multiple clients simultaneously.

## Dependencies
The project relies on the following dependencies:
- `github.com/quic-go/quic-go`: QUIC transport protocol implementation for Go.
- `github.com/shirou/gopsutil`: Library for retrieving system information and metrics.

Make sure to install the dependencies before running the project.

## Limitations and Considerations
- The project assumes a simple authentication mechanism using JSON Web Tokens (JWT). In a production environment, a more robust authentication and authorization system should be considered.
- The code provided is a prototype and may require further optimizations, error handling, and testing before deployment in a production environment.
- The project focuses on the core functionality of the health check protocol and does not includ