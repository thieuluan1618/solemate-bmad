# SoleMate Operations Runbook

## Emergency Contacts
- **On-Call Engineer**: [Your contact info]
- **DevOps Lead**: [Contact info]
- **Technical Lead**: [Contact info]
- **Security Team**: [Contact info]

## Quick Reference

### Critical Service URLs
```bash
# Production Load Balancer
PROD_LB="http://production-ALB-XXXXXXXXX.us-east-1.elb.amazonaws.com"

# Health Check
curl -f "$PROD_LB/health"

# Service Status
curl -f "$PROD_LB/api/v1/health/detailed"
```

### AWS Resources
- **ECS Cluster**: `production-cluster`
- **ECS Service**: `solemate-backend`
- **RDS Instance**: `production-postgres`
- **Redis Cluster**: `production-redis`
- **CloudFormation Stack**: `solemate-infrastructure`

## Emergency Procedures

### 1. Service Outage Response

#### Immediate Actions (0-5 minutes)
1. **Acknowledge the incident** in monitoring system
2. **Check overall system status**:
   ```bash
   # Check ECS service health
   aws ecs describe-services \
     --cluster production-cluster \
     --services solemate-backend \
     --region us-east-1

   # Check running tasks
   aws ecs list-tasks \
     --cluster production-cluster \
     --service-name solemate-backend \
     --desired-status RUNNING \
     --region us-east-1
   ```

3. **Check recent deployments**:
   ```bash
   # Check recent GitHub Actions runs
   gh run list --limit 5

   # Check CloudFormation stack events
   aws cloudformation describe-stack-events \
     --stack-name solemate-infrastructure \
     --max-items 10 \
     --region us-east-1
   ```

#### Assessment Phase (5-15 minutes)
1. **Identify the scope**:
   - Partial service degradation vs complete outage
   - Which services/endpoints are affected
   - Geographic impact

2. **Check infrastructure**:
   ```bash
   # Database status
   aws rds describe-db-instances \
     --db-instance-identifier production-postgres \
     --region us-east-1

   # Redis status
   aws elasticache describe-cache-clusters \
     --cache-cluster-id production-redis \
     --show-cache-node-info \
     --region us-east-1

   # Load balancer status
   aws elbv2 describe-target-health \
     --target-group-arn $(aws elbv2 describe-target-groups \
       --names production-TG \
       --query 'TargetGroups[0].TargetGroupArn' \
       --output text) \
     --region us-east-1
   ```

3. **Review logs**:
   ```bash
   # Application logs (last 30 minutes)
   aws logs filter-log-events \
     --log-group-name /ecs/solemate \
     --start-time $(date -d '30 minutes ago' +%s)000 \
     --region us-east-1

   # Specific service errors
   aws logs filter-log-events \
     --log-group-name /ecs/solemate \
     --filter-pattern "ERROR" \
     --start-time $(date -d '1 hour ago' +%s)000 \
     --region us-east-1
   ```

#### Mitigation Phase (15-30 minutes)
1. **Quick fixes**:
   ```bash
   # Restart ECS service (force new deployment)
   aws ecs update-service \
     --cluster production-cluster \
     --service solemate-backend \
     --force-new-deployment \
     --region us-east-1

   # Scale up if needed
   aws ecs update-service \
     --cluster production-cluster \
     --service solemate-backend \
     --desired-count 4 \
     --region us-east-1
   ```

2. **Rollback if necessary**:
   ```bash
   # Get previous task definition
   PREVIOUS_REVISION=$(aws ecs list-task-definitions \
     --family-prefix solemate-services \
     --status ACTIVE \
     --sort DESC \
     --query 'taskDefinitionArns[1]' \
     --output text | sed 's/.*://')

   # Rollback to previous version
   aws ecs update-service \
     --cluster production-cluster \
     --service solemate-backend \
     --task-definition solemate-services:$PREVIOUS_REVISION \
     --region us-east-1
   ```

### 2. Database Issues

#### Connection Pool Exhaustion
```bash
# Check active connections
aws rds describe-db-instances \
  --db-instance-identifier production-postgres \
  --query 'DBInstances[0].{ConnectionCount:DbInstanceStatus,MaxConnections:MaxAllocatedStorage}' \
  --region us-east-1

# Restart services to reset connections
aws ecs update-service \
  --cluster production-cluster \
  --service solemate-backend \
  --force-new-deployment \
  --region us-east-1
```

#### Database Performance Issues
```bash
# Check CPU utilization
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name CPUUtilization \
  --dimensions Name=DBInstanceIdentifier,Value=production-postgres \
  --start-time $(date -d '1 hour ago' -u +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Average \
  --region us-east-1

# Check database locks
# Connect to RDS and run: SELECT * FROM pg_locks WHERE NOT granted;
```

#### Database Failover
```bash
# Force failover (Multi-AZ only)
aws rds reboot-db-instance \
  --db-instance-identifier production-postgres \
  --force-failover \
  --region us-east-1
```

### 3. High Load Response

#### Scale Out Quickly
```bash
# Increase ECS service count
aws ecs update-service \
  --cluster production-cluster \
  --service solemate-backend \
  --desired-count 6 \
  --region us-east-1

# Monitor scaling progress
watch aws ecs describe-services \
  --cluster production-cluster \
  --services solemate-backend \
  --query 'services[0].{Running:runningCount,Desired:desiredCount,Pending:pendingCount}' \
  --region us-east-1
```

#### Check Resource Utilization
```bash
# ECS cluster utilization
aws ecs describe-clusters \
  --clusters production-cluster \
  --include ATTACHMENTS,CONFIGURATION,STATISTICS,TAGS \
  --region us-east-1

# Individual task resource usage
aws ecs describe-tasks \
  --cluster production-cluster \
  --tasks $(aws ecs list-tasks \
    --cluster production-cluster \
    --service-name solemate-backend \
    --query 'taskArns[0]' \
    --output text) \
  --include TAGS \
  --region us-east-1
```

### 4. Security Incident Response

#### Suspicious Activity Detection
1. **Check WAF logs** (if implemented)
2. **Review access logs**:
   ```bash
   # ALB access logs (if enabled)
   aws s3 ls s3://your-alb-logs-bucket/AWSLogs/ACCOUNT-ID/elasticloadbalancing/us-east-1/

   # CloudTrail events
   aws logs filter-log-events \
     --log-group-name CloudTrail/SoleMateAudit \
     --filter-pattern "{ ($.errorCode = \"*UnauthorizedOperation\") || ($.errorCode = \"AccessDenied*\") }" \
     --start-time $(date -d '1 hour ago' +%s)000 \
     --region us-east-1
   ```

#### Immediate Isolation
```bash
# Block suspicious traffic via Security Groups
aws ec2 authorize-security-group-ingress \
  --group-id sg-LOADBALANCER-ID \
  --protocol tcp \
  --port 80 \
  --source-group sg-DENY-ALL \
  --region us-east-1

# Force rotate secrets
aws secretsmanager rotate-secret \
  --secret-id solemate/jwt-access-secret \
  --force-rotate-immediately \
  --region us-east-1
```

## Monitoring and Alerts

### Key Metrics to Monitor

#### Application Metrics
- **Response Time**: < 500ms p95
- **Error Rate**: < 1%
- **Throughput**: Baseline varies by time of day
- **Availability**: > 99.9%

#### Infrastructure Metrics
- **ECS CPU Utilization**: < 70%
- **ECS Memory Utilization**: < 80%
- **Database Connections**: < 80% of max
- **Database CPU**: < 70%
- **Redis Memory**: < 80%

#### Business Metrics
- **Order Processing Rate**
- **Payment Success Rate**: > 99%
- **Cart Abandonment Rate**
- **Search Response Time**

### Alert Escalation

#### P1 - Critical (Immediate Response)
- Complete service outage
- Payment processing failures
- Data breach suspected
- Database unavailable

#### P2 - High (Response within 1 hour)
- Partial service degradation
- High error rates (>5%)
- Performance degradation
- Single service failure

#### P3 - Medium (Response within 4 hours)
- Minor performance issues
- Non-critical service warnings
- Capacity warnings

#### P4 - Low (Response within 24 hours)
- Informational alerts
- Maintenance reminders
- Optimization opportunities

## Maintenance Procedures

### Scheduled Maintenance

#### Database Maintenance
```bash
# Create snapshot before maintenance
aws rds create-db-snapshot \
  --db-instance-identifier production-postgres \
  --db-snapshot-identifier production-postgres-maintenance-$(date +%Y%m%d) \
  --region us-east-1

# Apply pending updates (during maintenance window)
aws rds modify-db-instance \
  --db-instance-identifier production-postgres \
  --apply-immediately \
  --region us-east-1
```

#### ECS Cluster Updates
```bash
# Update ECS agent (if using EC2)
aws ecs update-container-instances-state \
  --cluster production-cluster \
  --container-instances CONTAINER_INSTANCE_ARN \
  --status DRAINING \
  --region us-east-1
```

### Deployment Procedures

#### Blue-Green Deployment
1. **Deploy to staging environment first**
2. **Run full test suite**
3. **Deploy to production via CI/CD**
4. **Monitor for 30 minutes**
5. **Rollback if issues detected**

#### Emergency Hotfix
```bash
# Create hotfix branch
git checkout -b hotfix/critical-fix

# Make minimal changes
# ...

# Deploy via CI/CD
git push origin hotfix/critical-fix

# Monitor deployment
gh run watch $(gh run list --limit 1 --json databaseId --jq '.[0].databaseId')
```

## Backup and Recovery

### Database Backups
```bash
# Manual snapshot
aws rds create-db-snapshot \
  --db-instance-identifier production-postgres \
  --db-snapshot-identifier manual-snapshot-$(date +%Y%m%d-%H%M) \
  --region us-east-1

# List available snapshots
aws rds describe-db-snapshots \
  --db-instance-identifier production-postgres \
  --snapshot-type manual \
  --region us-east-1
```

### Point-in-Time Recovery
```bash
# Restore to specific time
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier production-postgres \
  --target-db-instance-identifier production-postgres-restored \
  --restore-time 2024-01-20T10:00:00.000Z \
  --region us-east-1
```

### Configuration Backups
```bash
# Export current ECS service configuration
aws ecs describe-services \
  --cluster production-cluster \
  --services solemate-backend \
  --region us-east-1 > backup-service-config-$(date +%Y%m%d).json

# Export task definition
aws ecs describe-task-definition \
  --task-definition solemate-services \
  --region us-east-1 > backup-task-definition-$(date +%Y%m%d).json
```

## Performance Tuning

### Database Optimization
```sql
-- Check slow queries
SELECT query, mean_time, calls
FROM pg_stat_statements
ORDER BY mean_time DESC
LIMIT 10;

-- Check index usage
SELECT schemaname, tablename, attname, n_distinct, correlation
FROM pg_stats
WHERE schemaname = 'public';
```

### Application Optimization
```bash
# Check memory usage by service
aws ecs describe-tasks \
  --cluster production-cluster \
  --tasks $(aws ecs list-tasks \
    --cluster production-cluster \
    --service-name solemate-backend \
    --query 'taskArns' \
    --output text) \
  --query 'tasks[0].containers[].{Name:name,Memory:memory,MemoryReservation:memoryReservation}' \
  --region us-east-1
```

### Load Testing
```bash
# Run performance tests
cd tests/performance
go test -v -timeout=30m -endpoint=$PROD_LB
```

## Common Issues and Solutions

### Issue: ECS Tasks Failing to Start
**Symptoms**: Tasks start and immediately stop
**Diagnosis**:
```bash
aws ecs describe-tasks \
  --cluster production-cluster \
  --tasks FAILED_TASK_ARN \
  --region us-east-1
```
**Solutions**:
- Check resource limits (CPU/Memory)
- Verify image exists in ECR
- Check secrets configuration
- Review security group rules

### Issue: High Database CPU
**Symptoms**: Slow application responses, database CPU > 80%
**Diagnosis**:
```bash
aws cloudwatch get-metric-statistics \
  --namespace AWS/RDS \
  --metric-name CPUUtilization \
  --dimensions Name=DBInstanceIdentifier,Value=production-postgres \
  --start-time $(date -d '1 hour ago' -u +%Y-%m-%dT%H:%M:%S) \
  --end-time $(date -u +%Y-%m-%dT%H:%M:%S) \
  --period 300 \
  --statistics Maximum \
  --region us-east-1
```
**Solutions**:
- Scale database instance class
- Optimize queries
- Add read replicas
- Enable connection pooling

### Issue: Load Balancer 502 Errors
**Symptoms**: Users receiving 502 Bad Gateway errors
**Diagnosis**:
```bash
aws elbv2 describe-target-health \
  --target-group-arn $(aws elbv2 describe-target-groups \
    --names production-TG \
    --query 'TargetGroups[0].TargetGroupArn' \
    --output text) \
  --region us-east-1
```
**Solutions**:
- Check ECS task health
- Verify security group rules
- Check application health endpoints
- Review load balancer configuration

## Useful Commands Reference

### ECS Management
```bash
# Force service update
aws ecs update-service --cluster CLUSTER --service SERVICE --force-new-deployment

# Scale service
aws ecs update-service --cluster CLUSTER --service SERVICE --desired-count COUNT

# Stop all tasks
aws ecs update-service --cluster CLUSTER --service SERVICE --desired-count 0

# Execute into running container
aws ecs execute-command --cluster CLUSTER --task TASK_ARN --container CONTAINER --interactive --command "/bin/sh"
```

### Logging
```bash
# Tail logs
aws logs tail LOG_GROUP --follow

# Search logs
aws logs filter-log-events --log-group-name LOG_GROUP --filter-pattern "ERROR"

# Export logs
aws logs create-export-task --log-group-name LOG_GROUP --from FROM_TIME --to TO_TIME --destination S3_BUCKET
```

### Secrets Management
```bash
# Get secret value
aws secretsmanager get-secret-value --secret-id SECRET_NAME

# Update secret
aws secretsmanager update-secret --secret-id SECRET_NAME --secret-string "NEW_VALUE"

# Rotate secret
aws secretsmanager rotate-secret --secret-id SECRET_NAME
```

---

**Last Updated**: January 2025
**Version**: 1.0
**Maintained By**: DevOps Team