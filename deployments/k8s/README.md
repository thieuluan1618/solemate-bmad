# Kubernetes Deployment Guide for SoleMate

## Quick Start

### Prerequisites
- Kubernetes cluster (1.24+)
- kubectl configured
- Docker registry access

### Deployment Steps

1. **Apply all configurations:**
```bash
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f secrets.yaml
kubectl apply -f postgres-statefulset.yaml
kubectl apply -f redis-deployment.yaml
kubectl apply -f rabbitmq-statefulset.yaml
kubectl apply -f user-service-deployment.yaml
kubectl apply -f product-service-deployment.yaml
kubectl apply -f cart-service-deployment.yaml
kubectl apply -f order-service-deployment.yaml
kubectl apply -f payment-service-deployment.yaml
kubectl apply -f api-gateway-deployment.yaml
kubectl apply -f frontend-deployment.yaml
kubectl apply -f ingress.yaml
kubectl apply -f hpa-product-service.yaml
kubectl apply -f hpa-api-gateway.yaml
kubectl apply -f hpa-frontend.yaml
```

2. **Or use the deployment script:**
```bash
chmod +x deploy-k8s.sh
./deploy-k8s.sh --all
```

## Configuration Files

### Core Infrastructure
- `namespace.yaml` - SoleMate namespace
- `configmap.yaml` - Service URLs and configuration
- `secrets.yaml` - Sensitive credentials

### Databases & Cache
- `postgres-statefulset.yaml` - PostgreSQL with persistent storage
- `redis-deployment.yaml` - Redis cache
- `rabbitmq-statefulset.yaml` - Message queue

### Microservices
- `user-service-deployment.yaml` - User authentication service
- `product-service-deployment.yaml` - Product catalog service
- `cart-service-deployment.yaml` - Shopping cart service
- `order-service-deployment.yaml` - Order processing service
- `payment-service-deployment.yaml` - Payment processing service
- `inventory-service-deployment.yaml` - Inventory management
- `notification-service-deployment.yaml` - Email/SMS notifications

### Frontend & Gateway
- `api-gateway-deployment.yaml` - API routing
- `frontend-deployment.yaml` - Next.js application
- `ingress.yaml` - External access configuration

### Auto-scaling
- `hpa-*.yaml` - Horizontal Pod Autoscalers for each service

## Monitoring

Check deployment status:
```bash
kubectl get all -n solemate
kubectl get hpa -n solemate
kubectl top pods -n solemate
```

View logs:
```bash
kubectl logs -n solemate deployment/product-service
kubectl logs -n solemate -f deployment/api-gateway
```

## Scaling

Manual scaling:
```bash
kubectl scale deployment product-service --replicas=10 -n solemate
```

Auto-scaling is configured via HPA to handle up to 50,000 concurrent users.

## Troubleshooting

Common issues:
1. **Pods not starting**: Check logs with `kubectl logs`
2. **Service unavailable**: Verify service endpoints with `kubectl get endpoints`
3. **Database connection**: Check secrets and configmap values
4. **High memory usage**: Review resource limits in deployment files