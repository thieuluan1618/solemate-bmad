# Docker Swarm Deployment Guide for SoleMate

## Quick Start

### Initialize Swarm
```bash
docker swarm init
```

### Deploy Stack
```bash
docker stack deploy -c docker-stack.yml solemate
```

### Or use deployment script:
```bash
chmod +x deploy-swarm.sh
./deploy-swarm.sh deploy
```

## Stack Components

### Infrastructure Services
- **PostgreSQL**: Primary database (1 replica, persistent volume)
- **Redis**: Caching layer (1 replica)
- **RabbitMQ**: Message queue (1 replica)
- **Elasticsearch**: Search engine (1 replica)

### Application Services
- **Frontend**: Next.js app (5 replicas, scalable to 25)
- **API Gateway**: Traefik load balancer (3 replicas)
- **Microservices**: 2-5 replicas each based on load

### Monitoring
- **Prometheus**: Metrics collection
- **Grafana**: Visualization dashboards

## Scaling

### Manual Scaling
```bash
# Scale individual service
docker service scale solemate_frontend=10

# Scale multiple services
docker service scale \
  solemate_frontend=10 \
  solemate_product-service=8 \
  solemate_api-gateway=5
```

### Auto-scaling Script
```bash
# Scale for different load levels
./swarm-scale-for-load.sh normal  # 5k users
./swarm-scale-for-load.sh high    # 20k users
./swarm-scale-for-load.sh peak    # 50k users
./swarm-scale-for-load.sh auto    # CPU-based auto-scaling
```

## High Availability Setup

For production, setup multi-node cluster:
```bash
# On manager nodes
docker swarm join-token manager

# On worker nodes
docker swarm join --token <worker-token> <manager-ip>:2377

# Label nodes for placement
docker node update --label-add type=database worker1
docker node update --label-add type=app worker2
```

## Monitoring

### Check Services
```bash
docker stack services solemate
docker service ps solemate_frontend
docker stack ps solemate
```

### View Logs
```bash
docker service logs solemate_frontend
docker service logs -f solemate_api-gateway
```

### Access Dashboards
- Traefik Dashboard: http://localhost:8080
- Grafana: http://grafana.solemate.com (default: admin/admin)
- RabbitMQ Management: http://localhost:15672

## Rolling Updates

Update service with zero downtime:
```bash
docker service update \
  --image solemate/frontend:v2.0 \
  --update-parallelism 2 \
  --update-delay 10s \
  solemate_frontend
```

## Backup & Recovery

### Backup volumes
```bash
docker run --rm \
  -v solemate_postgres_data:/data \
  -v $(pwd):/backup \
  alpine tar czf /backup/postgres-backup.tar.gz /data
```

### Remove stack (preserves volumes)
```bash
docker stack rm solemate
```

### Remove everything including volumes
```bash
./deploy-swarm.sh remove --volumes
```

## Performance Tuning

### Optimize for 50k concurrent users:
1. Increase service replicas using peak scaling
2. Adjust resource limits in docker-stack.yml
3. Enable caching at Traefik level
4. Use placement constraints for database nodes
5. Monitor with Prometheus/Grafana

## Troubleshooting

### Service not starting
```bash
docker service ps solemate_<service-name> --no-trunc
docker service logs solemate_<service-name>
```

### Network issues
```bash
docker network ls
docker network inspect solemate_backend
```

### Resource constraints
```bash
docker node ls
docker node inspect <node-id>
```