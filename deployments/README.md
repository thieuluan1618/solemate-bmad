# SoleMate E-commerce Platform - Deployment Guide

## Overview

This guide provides comprehensive instructions for deploying the SoleMate e-commerce platform to AWS using ECS Fargate with automated CI/CD pipelines.

## Architecture

```
┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────────┐
│   Internet          │    │   Application       │    │   Database          │
│                     │    │   Load Balancer     │    │   Layer             │
│  ┌─────────────┐    │    │                     │    │                     │
│  │   Users     │────┼────┤ Application Load    ├────┼──┐ ┌─────────────┐   │
│  │             │    │    │ Balancer (ALB)      │    │  ├─┤ PostgreSQL  │   │
│  └─────────────┘    │    │                     │    │  │ │ RDS         │   │
│                     │    └─────────────────────┘    │  │ └─────────────┘   │
└─────────────────────┘                               │  │                   │
                                                      │  │ ┌─────────────┐   │
┌─────────────────────┐    ┌─────────────────────┐    │  ├─┤ Redis       │   │
│   Compute Layer     │    │   ECS Fargate       │    │  │ │ ElastiCache │   │
│                     │    │   Cluster           │    │  │ └─────────────┘   │
│  ┌─────────────┐    │    │                     │    │  │                   │
│  │ API Gateway │────┼────┤ ┌─────────────────┐ ├────┼──┘                   │
│  │             │    │    │ │ User Service    │ │    │                      │
│  └─────────────┘    │    │ │ Product Service │ │    │                      │
│                     │    │ │ Cart Service    │ │    │                      │
│  ┌─────────────┐    │    │ │ Order Service   │ │    │                      │
│  │ Frontend    │────┼────┤ │ Payment Service │ ├────┼────────────────────  │
│  │ (Future)    │    │    │ └─────────────────┘ │    │                      │
│  └─────────────┘    │    └─────────────────────┘    │                      │
└─────────────────────┘                               └─────────────────────┘
```

## Prerequisites

### Required Tools
- AWS CLI v2 configured with appropriate permissions
- Docker and Docker Compose
- Go 1.21+
- Git
- Make

### AWS Permissions Required
- ECS (Full access)
- ECR (Full access)
- RDS (Full access)
- ElastiCache (Full access)
- VPC (Full access)
- IAM (Role creation and management)
- Secrets Manager (Full access)
- CloudFormation (Full access)
- Application Load Balancer (Full access)

## Deployment Process

### Phase 1: Infrastructure Setup

#### 1. Clone Repository
```bash
git clone <repository-url>
cd solemate
```

#### 2. Configure AWS Environment
```bash
# Configure AWS CLI
aws configure

# Set environment variables
export AWS_REGION=us-east-1
export ENVIRONMENT=production
export AWS_ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
```

#### 3. Create AWS Secrets
```bash
# Setup all production secrets
cd deployments/aws
./secrets-setup.sh

# Update placeholder values manually in AWS Secrets Manager console
```

#### 4. Deploy Infrastructure
```bash
# Deploy VPC, ECS Cluster, RDS, ElastiCache, and Load Balancer
aws cloudformation create-stack \
  --stack-name solemate-infrastructure \
  --template-body file://cloudformation.yaml \
  --parameters ParameterKey=EnvironmentName,ParameterValue=production \
               ParameterKey=DBPassword,ParameterValue=YOUR_SECURE_PASSWORD \
               ParameterKey=StripeApiKey,ParameterValue=YOUR_STRIPE_KEY \
  --capabilities CAPABILITY_NAMED_IAM \
  --region $AWS_REGION

# Wait for stack creation (10-15 minutes)
aws cloudformation wait stack-create-complete \
  --stack-name solemate-infrastructure \
  --region $AWS_REGION
```

### Phase 2: Container Registry Setup

#### 1. Create ECR Repositories
```bash
# Repositories are created automatically by CloudFormation
# Verify they exist
aws ecr describe-repositories --region $AWS_REGION
```

#### 2. Build and Push Images
```bash
# Login to ECR
aws ecr get-login-password --region $AWS_REGION | \
  docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com

# Build and push all service images
for service in user-service product-service cart-service order-service payment-service api-gateway; do
  echo "Building $service..."
  docker build -f services/$service/Dockerfile -t $service .
  docker tag $service:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/solemate/$service:latest
  docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/solemate/$service:latest
done
```

### Phase 3: ECS Service Deployment

#### 1. Update Task Definition
```bash
# Update the task definition with your account ID and region
sed -i "s/ACCOUNT-ID/$AWS_ACCOUNT_ID/g" ecs-task-definition.json
sed -i "s/REGION/$AWS_REGION/g" ecs-task-definition.json

# Register task definition
aws ecs register-task-definition \
  --cli-input-json file://ecs-task-definition.json \
  --region $AWS_REGION
```

#### 2. Create ECS Service
```bash
# Update service definition
sed -i "s/ACCOUNT-ID/$AWS_ACCOUNT_ID/g" ecs-service.json
sed -i "s/REGION/$AWS_REGION/g" ecs-service.json

# Get subnet and security group IDs from CloudFormation outputs
PRIVATE_SUBNET_1=$(aws cloudformation describe-stacks \
  --stack-name solemate-infrastructure \
  --query 'Stacks[0].Outputs[?OutputKey==`PrivateSubnet1`].OutputValue' \
  --output text --region $AWS_REGION)

PRIVATE_SUBNET_2=$(aws cloudformation describe-stacks \
  --stack-name solemate-infrastructure \
  --query 'Stacks[0].Outputs[?OutputKey==`PrivateSubnet2`].OutputValue' \
  --output text --region $AWS_REGION)

ECS_SECURITY_GROUP=$(aws cloudformation describe-stacks \
  --stack-name solemate-infrastructure \
  --query 'Stacks[0].Outputs[?OutputKey==`ECSSecurityGroup`].OutputValue' \
  --output text --region $AWS_REGION)

TARGET_GROUP_ARN=$(aws cloudformation describe-stacks \
  --stack-name solemate-infrastructure \
  --query 'Stacks[0].Outputs[?OutputKey==`TargetGroup`].OutputValue' \
  --output text --region $AWS_REGION)

# Update service definition with actual values
sed -i "s/subnet-PRIVATE-1/$PRIVATE_SUBNET_1/g" ecs-service.json
sed -i "s/subnet-PRIVATE-2/$PRIVATE_SUBNET_2/g" ecs-service.json
sed -i "s/sg-ecs-backend/$ECS_SECURITY_GROUP/g" ecs-service.json
sed -i "s|arn:aws:elasticloadbalancing:REGION:ACCOUNT-ID:targetgroup/solemate-backend-tg/TARGET-GROUP-ID|$TARGET_GROUP_ARN|g" ecs-service.json

# Create ECS service
aws ecs create-service \
  --cli-input-json file://ecs-service.json \
  --region $AWS_REGION
```

### Phase 4: CI/CD Pipeline Setup

#### 1. Configure GitHub Secrets
Add the following secrets to your GitHub repository:
- `AWS_ACCESS_KEY_ID`: IAM user access key for deployments
- `AWS_SECRET_ACCESS_KEY`: IAM user secret key for deployments
- `AWS_ACCOUNT_ID`: Your AWS account ID

#### 2. Enable GitHub Actions
The CI/CD pipeline will automatically trigger on:
- Push to `main` branch (full deployment)
- Pull requests (testing only)

### Phase 5: DNS and SSL Setup

#### 1. Configure Route 53 (Optional)
```bash
# Create hosted zone for your domain
aws route53 create-hosted-zone \
  --name your-domain.com \
  --caller-reference $(date +%s)

# Get Load Balancer DNS name
LB_DNS=$(aws cloudformation describe-stacks \
  --stack-name solemate-infrastructure \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
  --output text --region $AWS_REGION)

echo "Point your domain to: $LB_DNS"
```

#### 2. SSL Certificate (Optional)
```bash
# Request ACM certificate
aws acm request-certificate \
  --domain-name your-domain.com \
  --subject-alternative-names www.your-domain.com \
  --validation-method DNS \
  --region $AWS_REGION
```

## Post-Deployment Verification

### 1. Health Checks
```bash
# Get Load Balancer URL
LB_URL=$(aws cloudformation describe-stacks \
  --stack-name solemate-infrastructure \
  --query 'Stacks[0].Outputs[?OutputKey==`LoadBalancerURL`].OutputValue' \
  --output text --region $AWS_REGION)

# Test health endpoint
curl -f "$LB_URL/health"

# Test API endpoints
curl -f "$LB_URL/api/v1/products"
```

### 2. Service Status
```bash
# Check ECS service status
aws ecs describe-services \
  --cluster production-cluster \
  --services solemate-backend \
  --region $AWS_REGION

# Check task health
aws ecs list-tasks \
  --cluster production-cluster \
  --service-name solemate-backend \
  --region $AWS_REGION
```

### 3. Logs
```bash
# View application logs
aws logs tail /ecs/solemate --follow --region $AWS_REGION

# View specific service logs
aws logs tail /ecs/solemate --log-stream-name-prefix user-service --follow --region $AWS_REGION
```

## Monitoring and Maintenance

### CloudWatch Dashboards
- Application Performance Monitoring (APM)
- Database Performance
- Load Balancer Metrics
- ECS Cluster Utilization

### Alerts
- High CPU/Memory usage
- Application errors
- Database connection issues
- Load balancer 5xx errors

### Backup Strategy
- RDS automated backups (7-day retention)
- Database snapshots before major deployments
- Configuration backups in S3

## Scaling

### Horizontal Scaling
```bash
# Scale ECS service
aws ecs update-service \
  --cluster production-cluster \
  --service solemate-backend \
  --desired-count 4 \
  --region $AWS_REGION
```

### Vertical Scaling
Update task definition with higher CPU/memory values and redeploy.

## Rollback Procedures

### 1. Emergency Rollback
```bash
# Rollback to previous task definition
aws ecs update-service \
  --cluster production-cluster \
  --service solemate-backend \
  --task-definition solemate-services:PREVIOUS_REVISION \
  --region $AWS_REGION
```

### 2. Database Rollback
```bash
# Restore from RDS snapshot
aws rds restore-db-instance-from-db-snapshot \
  --db-instance-identifier production-postgres-restored \
  --db-snapshot-identifier production-postgres-snapshot-TIMESTAMP \
  --region $AWS_REGION
```

## Troubleshooting

### Common Issues

#### 1. Service Not Starting
- Check CloudWatch logs
- Verify secrets are properly configured
- Ensure security groups allow traffic
- Check task definition resource limits

#### 2. Database Connection Issues
- Verify RDS security group rules
- Check database credentials in Secrets Manager
- Ensure subnets have proper routing

#### 3. Load Balancer Health Check Failures
- Verify health check endpoint responds correctly
- Check security group rules for ALB to ECS communication
- Review task startup time vs health check interval

### Debug Commands
```bash
# SSH into running task (if enabled)
aws ecs execute-command \
  --cluster production-cluster \
  --task TASK_ARN \
  --container user-service \
  --interactive \
  --command "/bin/sh"

# View detailed task information
aws ecs describe-tasks \
  --cluster production-cluster \
  --tasks TASK_ARN \
  --region $AWS_REGION
```

## Security Considerations

### 1. Network Security
- Services run in private subnets
- ALB in public subnets only
- Security groups with minimal required permissions

### 2. Access Control
- IAM roles with least privilege
- Secrets Manager for sensitive data
- No hardcoded credentials

### 3. Data Encryption
- RDS encryption at rest
- ELB SSL/TLS termination
- Secrets Manager encrypted storage

## Cost Optimization

### 1. Right-sizing
- Monitor CPU/memory utilization
- Use Fargate Spot for non-critical workloads
- Schedule services for non-production environments

### 2. Reserved Capacity
- RDS Reserved Instances
- ElastiCache Reserved Nodes
- Consider Savings Plans for consistent usage

## Disaster Recovery

### 1. Multi-AZ Deployment
- RDS Multi-AZ enabled
- ECS tasks distributed across AZs
- Load balancer spans multiple AZs

### 2. Backup Strategy
- Automated RDS backups
- Infrastructure as Code (CloudFormation)
- Container images in ECR

### 3. Recovery Procedures
- Infrastructure recreation from CloudFormation
- Database restoration from snapshots
- Application deployment via CI/CD pipeline

---

## Support Contacts

- **Infrastructure Issues**: DevOps Team
- **Application Issues**: Development Team
- **Security Issues**: Security Team
- **Emergency**: On-call rotation

## Documentation Updates

This documentation should be updated whenever:
- Infrastructure changes are made
- New services are added
- Security policies change
- Monitoring requirements change