#!/bin/bash

# Function to initialize Go module
init_module() {
    echo "Initializing Go module..."
    go mod init your-module-path
    go mod download
    go mod tidy
}

# Function to display the menu
show_menu() {
    clear
    echo "Quick Health Check Protocol (QHCP) using QUIC"
    echo "----------------------------------------------"
    echo "1. Run Load Balancer"
    echo "2. Run Server"
    echo "3. Exit"
    echo -n "Enter your choice (1-3): "
}

# Function to run the load balancer
run_load_balancer() {
    echo "Enter the comma-separated list of server addresses (host:port) (e.g., localhost:4243,localhost:4244):"
    read servers
    echo "Enter the port for the load balancer (default: 4242):"
    read loadbalancer_port
    echo "Enter the maximum fail attempts before marking a server as down (default: 3):"
    read max_fail_attempts
    echo "Enter the interval for health checks and status display in seconds (default: 10):"
    read check_interval
    echo "Enter the interval for attempting to reconnect to down servers in seconds (default: 30):"
    read reconnect_interval
    echo "Enter the TLS certificate file (leave blank for default):"
    read cert_file

    init_module

    go run cmd/echo/echo.go -loadbalancer \
        -servers "$servers" \
        -loadbalancer-port "${loadbalancer_port:-4242}" \
        -max-fail-attempts "${max_fail_attempts:-3}" \
        -check-interval "${check_interval:-10}" \
        -reconnect-interval "${reconnect_interval:-30}" \
        -cert-file "${cert_file}"
}

# Function to run the server
run_server() {
    echo "Enter the server IP (default: 0.0.0.0):"
    read server_ip
    echo "Enter the server port (default: 4243):"
    read server_port
    echo "Enter the TLS key file (leave blank for default):"
    read key_file
    echo "Enter the TLS certificate file (leave blank for default):"
    read cert_file

    init_module

    go run cmd/echo/echo.go -server \
        -server-ip "${server_ip:-0.0.0.0}" \
        -server-port "${server_port:-4243}" \
        -key-file "${key_file}" \
        -cert-file "${cert_file}"
}

# Main loop
while true; do
    show_menu
    read choice

    case $choice in
        1)
            run_load_balancer
            ;;
        2)
            run_server
            ;;
        3)
            exit 0
            ;;
        *)
            echo "Invalid choice. Please try again."
            ;;
    esac

    read -p "Press Enter to continue..."
done