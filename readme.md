# Quick Health Check Protocol (QHCP) using QUIC

This project implements a Quick Health Check Protocol (QHCP) using the QUIC transport protocol. The QHCP is designed to monitor the health of multiple backend servers and facilitate load balancing decisions. The solution consists of a load balancer and server components.

## Functionality

The QHCP implementation provides the following key functionalities:

1. **Health Monitoring**: The load balancer periodically sends health check requests to the backend servers and monitors their health status based on the responses.
2. **Server Reconnection**: If a server goes down, the load balancer attempts to reconnect to the server at regular intervals defined by the reconnect interval.
3. **Metrics Collection**: The server collects and sends back health metrics such as CPU usage percentage and memory usage percentage to the load balancer.
4. **Authentication**: The protocol incorporates JSON Web Tokens (JWT) for authentication between the load balancer and servers.
Protocol Messaging: The protocol defines various message types for communication between the load balancer and servers, including HELLO, ACK, HEALTH_REQUEST, HEALTH_RESPONSE, CONFIG_UPDATE, CONFIG_ACK, ERROR, TERMINATE, and TERMINATE_ACK.
5. **Secure Communication**: The protocol utilizes QUIC's built-in encryption for secure data transmission between the load balancer and servers.

## Usage

The project provides a user-friendly Bash script `run_quic.sh` that offers an interactive menu to run either the load balancer or the server(s). The script prompts the user for necessary configuration options such as server addresses, ports, and TLS settings. It initializes the Go module and runs the appropriate command based on the user's selections.

### Load Balancer

To create a load balancer, follow these steps:

1. Run the `run_quic.sh` script:
./run_quic.sh

2. Select option "1. Run Load Balancer" from the menu.

3. Provide the necessary configuration options when prompted:

```
Enter the comma-separated list of server addresses (host:port) (e.g., localhost:4243,localhost:4244): localhost:4243,localhost:4244
Enter the port for the load balancer (default: 4242): 4242
Enter the maximum fail attempts before marking a server as down (default: 3): 3
Enter the interval for health checks and status display in seconds (default: 10): 10
Enter the interval for attempting to reconnect to down servers in seconds (default: 30): 30
Enter the TLS certificate file (leave blank for default):
```
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

