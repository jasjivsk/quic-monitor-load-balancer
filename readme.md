**# Quick Health Check Protocol (QHCP) using QUIC**
This project implements a Quick Health Check Protocol (QHCP) using the QUIC transport protocol. The QHCP is designed to monitor the health of multiple backend servers and facilitate load balancing decisions. The solution consists of a load balancer and server components.

**## Usage**
The project provides a single binary that can be used to run both the load balancer and the server(s). You can use the provided `run_quic.sh` script to run the load balancer or server with the necessary parameters.

**### Load Balancer**
To start the load balancer using the `run_quic.sh` script, follow these steps:

1. Run the script: `./run_quic.sh`
2. Select option "1. Run Load Balancer" from the menu.
3. When prompted, enter the comma-separated list of server addresses (e.g., `localhost:4243,localhost:4244`).
4. Optionally, you can provide values for the load balancer port, maximum fail attempts, check interval, reconnect interval, and TLS certificate file, or press Enter to use the default values.

Sample input:
```
Enter the comma-separated list of server addresses (host:port) (e.g., localhost:4243,localhost:4244): localhost:4243,localhost:4244
Enter the port for the load balancer (default: 4242): [Press Enter to use the default value]
Enter the maximum fail attempts before marking a server as down (default: 3): 5
Enter the interval for health checks and status display in seconds (default: 10): [Press Enter to use the default value]
Enter the interval for attempting to reconnect to down servers in seconds (default: 30): [Press Enter to use the default value]
Enter the TLS certificate file (leave blank for default): [Press Enter to use the default value]
```

**### Server**
To start a server using the `run_quic.sh` script, follow these steps:

1. Run the script: `./run_quic.sh`
2. Select option "2. Run Server" from the menu.
3. Optionally, you can provide values for the server IP, server port, TLS key file, and TLS certificate file, or press Enter to use the default values.

Sample input:
```
Enter the server IP (default: 0.0.0.0): [Press Enter to use the default value]
Enter the server port (default: 4243): 4244
Enter the TLS key file (leave blank for default): [Press Enter to use the default value]
Enter the TLS certificate file (leave blank for default): [Press Enter to use the default value]
```

**## Customization**
The project can be customized and extended based on specific requirements. The code provides a foundation for implementing a health check protocol using QUIC, and it can be modified to include additional features, metrics, or customizations as needed.

**## Dependencies**
The project relies on the following dependencies:
- `github.com/quic-go/quic-go`: QUIC transport protocol implementation for Go
- `github.com/shirou/gopsutil`: Library for retrieving system information and metrics

Make sure to install the dependencies before running the project.

**## Limitations**
- The current implementation assumes a simple authentication mechanism using JSON Web Tokens (JWT). In a production environment, a more robust authentication and authorization system should be considered.
- The code provided is a starting point and may require further optimizations, error handling, and testing before deployment in a production environment.
- The project focuses on the core functionality of the health check protocol and does not include advanced features such as load balancing algorithms or distributed coordination among multiple load balancers.