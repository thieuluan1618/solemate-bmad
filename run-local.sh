#!/bin/bash
# run-local.sh - Run SoleMate locally with different orchestration options

set -e

PROJECT_ROOT=$(cd "$(dirname "$0")" && pwd)
cd $PROJECT_ROOT

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

show_menu() {
    echo -e "${GREEN}SoleMate Local Development Runner${NC}"
    echo "======================================"
    echo "1) Docker Compose (Development)"
    echo "2) Docker Swarm (Single Node)"
    echo "3) Kubernetes (Minikube)"
    echo "4) Minimal Services (Frontend + API + DB)"
    echo "5) Stop All"
    echo "6) Clean Everything"
    echo "0) Exit"
}

run_docker_compose() {
    echo -e "${YELLOW}Starting with Docker Compose...${NC}"
    docker-compose up -d
    echo -e "${GREEN}Services running at:${NC}"
    echo "Frontend: http://localhost:3000"
    echo "API Gateway: http://localhost:8000"
    echo "Postgres: localhost:5432"
    echo "Redis: localhost:6379"
}

run_docker_swarm() {
    echo -e "${YELLOW}Starting with Docker Swarm...${NC}"
    
    # Initialize swarm if needed
    if docker info | grep -q "Swarm: inactive"; then
        docker swarm init
    fi
    
    # Build images locally
    echo "Building images..."
    docker-compose build
    
    # Deploy stack
    docker stack deploy -c docker-stack.yml solemate
    
    echo -e "${GREEN}Swarm stack deployed!${NC}"
    echo "Check status: docker stack services solemate"
}

run_kubernetes() {
    echo -e "${YELLOW}Starting with Minikube...${NC}"
    
    # Check if minikube is installed
    if ! command -v minikube &> /dev/null; then
        echo "Installing Minikube..."
        brew install minikube
    fi
    
    # Start minikube if not running
    if ! minikube status | grep -q "Running"; then
        minikube start --cpus=4 --memory=8192
    fi
    
    # Use minikube's docker
    eval $(minikube docker-env)
    
    # Build images in minikube
    echo "Building images in Minikube..."
    docker build -t solemate/frontend:latest ./frontend
    docker build -t solemate/user-service:latest -f services/user-service/Dockerfile .
    docker build -t solemate/product-service:latest -f services/product-service/Dockerfile .
    
    # Apply k8s configs
    kubectl apply -f deployments/k8s/namespace.yaml
    kubectl apply -f deployments/k8s/configmap.yaml
    kubectl apply -f deployments/k8s/postgres-statefulset.yaml
    kubectl apply -f deployments/k8s/redis-deployment.yaml
    kubectl apply -f deployments/k8s/user-service-deployment.yaml
    kubectl apply -f deployments/k8s/product-service-deployment.yaml
    kubectl apply -f deployments/k8s/frontend-deployment.yaml
    
    echo -e "${GREEN}Kubernetes deployment complete!${NC}"
    echo "Getting service URLs..."
    minikube service list -n solemate
}

run_minimal() {
    echo -e "${YELLOW}Starting minimal services...${NC}"
    
    # Start only essential services
    docker-compose up -d postgres redis
    sleep 5
    docker-compose up -d user-service product-service api-gateway frontend
    
    echo -e "${GREEN}Minimal services running!${NC}"
    echo "Frontend: http://localhost:3000"
}

stop_all() {
    echo -e "${YELLOW}Stopping all services...${NC}"
    
    # Stop Docker Compose
    docker-compose down
    
    # Stop Swarm stack
    docker stack rm solemate 2>/dev/null || true
    
    # Stop Minikube
    minikube stop 2>/dev/null || true
    
    echo -e "${GREEN}All services stopped${NC}"
}

clean_all() {
    echo -e "${YELLOW}Cleaning everything...${NC}"
    
    stop_all
    
    # Remove volumes
    docker volume prune -f
    
    # Remove images
    docker image prune -a -f
    
    # Clean minikube
    minikube delete 2>/dev/null || true
    
    echo -e "${GREEN}Cleanup complete${NC}"
}

check_requirements() {
    echo "Checking requirements..."
    
    # Check Docker
    if ! command -v docker &> /dev/null; then
        echo "❌ Docker not installed"
        exit 1
    fi
    
    # Check memory
    if [[ "$OSTYPE" == "darwin"* ]]; then
        TOTAL_MEM=$(sysctl -n hw.memsize | awk '{print $1/1024/1024/1024}')
        if (( $(echo "$TOTAL_MEM < 8" | bc -l) )); then
            echo "⚠️  Warning: Less than 8GB RAM available"
        fi
    fi
    
    echo "✅ Requirements OK"
}

# Main menu loop
check_requirements

while true; do
    show_menu
    read -p "Select option: " choice
    
    case $choice in
        1)
            run_docker_compose
            ;;
        2)
            run_docker_swarm
            ;;
        3)
            run_kubernetes
            ;;
        4)
            run_minimal
            ;;
        5)
            stop_all
            ;;
        6)
            clean_all
            ;;
        0)
            echo "Exiting..."
            exit 0
            ;;
        *)
            echo "Invalid option"
            ;;
    esac
    
    echo ""
    read -p "Press Enter to continue..."
done