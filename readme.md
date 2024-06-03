# Quick Health Check Protocol (QHCP) using QUIC

This project implements a Quick Health Check Protocol (QHCP) using the QUIC transport protocol. The QHCP is designed to monitor the health of multiple backend servers and facilitate load balancing decisions. The solution consists of a load balancer and server components.

## Protocol Design

The Quick Health Check Protocol (QHCP) is a stateful protocol that follows a specific Deterministic Finite Automaton (DFA) for its operation. Both the client (load balancer) and server must implement and validate the statefulness of the protocol according to the defined DFA.

The protocol operates on port 4242 by default, which can be customized by the user. The client defaults to this port number when connecting to the server.

## Functionality

The QHCP implementation provides the following key functionalities:

1. **Health Monitoring**: The load balancer periodically sends health check requests to the connected servers to assess their health status. The servers respond with their current health metrics, including CPU usage and memory usage percentages.

2. **Load Balancing**: Based on the health status of the servers, the load balancer makes intelligent decisions to distribute the traffic among the healthy servers. It maintains a map of server health statuses and dynamically updates it based on the health check responses.

3. **Fault Tolerance**: The load balancer is designed to handle server failures gracefully. It keeps track of the number of failed health check attempts for each server and marks a server as down if it exceeds a configurable threshold. Down servers are periodically retried for reconnection.

4. **Secure Communication**: The QHCP implementation leverages the security features provided by QUIC. It supports TLS encryption for secure communication between the load balancer and servers. The TLS configuration can be customized using certificate and key files.

## Usage

The project provides a user-friendly Bash script `run_quic.sh` that offers an interactive menu to run either the load balancer or the server(s). The script prompts the user for necessary configuration options such as server addresses, ports, and TLS settings. It initializes the Go module and runs the appropriate command based on the user's selections.

### Load Balancer

To create a load balancer, follow these steps:

1. Run the `run_quic.sh` script:
./run_quic.sh

2. Select option "1. Run Load Balancer" from the menu.

3. Provide the necessary configuration options when prompted:

Enter the comma-separated list of server addresses (host:port) (e.g., localhost:4243,localhost:4244): localhost:4243,localhost:4244
Enter the port for the load balancer (default: 4242): 4242
Enter the maximum fail attempts before marking a server as down (default: 3): 3
Enter the interval for health checks and status display in seconds (default: 10): 10
Enter the interval for attempting to reconnect to down servers in seconds (default: 30): 30
Enter the TLS certificate file (leave blank for default):

In this example, the load balancer is configured to connect to two servers (`localhost:4243` and `localhost:4244`), listen on port 4242, use the default values for maximum fail attempts, health check interval, and reconnect interval, and use the default TLS certificate file.

### Server(s)

The `run_quic.sh` script allows you to start a single server or multiple servers. To create server(s), follow these steps:

1. Run the `run_quic.sh` script:
./run_quic.sh

2. Select option "2. Run Server" from the menu.

3. Provide the necessary configuration options when prompted:
```
Enter the server IP (default: 0.0.0.0): 0.0.0.0
Enter the server port (default: 4243): 4243
Enter the TLS key file (leave blank for default):
Enter the TLS certificate file (leave blank for default):
```
In this example, a single server is started with the default IP address (`0.0.0.0`), listening on port 4243, and using the default TLS key and certificate files.

To start multiple servers, repeat steps 1-3 and provide different port numbers for each server instance. For example:
```
Enter the server IP (default: 0.0.0.0): 0.0.0.0
Enter the server port (default: 4243): 4244
Enter the TLS key file (leave blank for default):
Enter the TLS certificate file (leave blank for default):
```
This will start another server instance listening on port 4244.

By following these steps and providing the desired configuration options, you can create a load balancer and server(s) using the `run_quic.sh` script. The script allows you to easily start a single server or multiple servers, depending on your requirements.

Remember to adjust the configuration options based on your specific setup and needs. The provided examples demonstrate the flexibility and ease of use of the `run_quic.sh` script for creating load balancers and servers.

## Customization and Extensibility

The QHCP implementation serves as a solid foundation for building a health check and load balancing system. It can be easily customized and extended to incorporate additional features and requirements. The modular design of the codebase allows for seamless integration of new health metrics, load balancing algorithms, and configuration options.

