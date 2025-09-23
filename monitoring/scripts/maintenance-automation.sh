#!/bin/bash
set -e

# SoleMate Maintenance Automation Script
# This script performs automated maintenance tasks for the production environment

# Configuration
ENVIRONMENT=${ENVIRONMENT:-production}
AWS_REGION=${AWS_REGION:-us-east-1}
ECS_CLUSTER="${ENVIRONMENT}-cluster"
ECS_SERVICE="solemate-backend"
DB_INSTANCE="${ENVIRONMENT}-postgres"
REDIS_CLUSTER="${ENVIRONMENT}-redis"
SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL:-""}
LOG_FILE="/tmp/solemate-maintenance-$(date +%Y%m%d-%H%M%S).log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_info() {
    echo -e "${GREEN}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

# Slack notification function
send_slack_notification() {
    local message="$1"
    local color="$2"

    if [[ -n "$SLACK_WEBHOOK_URL" ]]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"$color\",\"text\":\"🔧 SoleMate Maintenance\\n$message\"}]}" \
            "$SLACK_WEBHOOK_URL" >/dev/null 2>&1
    fi
}

# Health check function
check_service_health() {
    log_info "Performing health checks..."

    # Get load balancer DNS
    LB_DNS=$(aws elbv2 describe-load-balancers \
        --names "${ENVIRONMENT}-ALB" \
        --query 'LoadBalancers[0].DNSName' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "")

    if [[ -n "$LB_DNS" ]]; then
        # Test health endpoint
        if curl -f -s "http://$LB_DNS/health" >/dev/null; then
            log_info "Application health check: PASSED"
            return 0
        else
            log_error "Application health check: FAILED"
            return 1
        fi
    else
        log_error "Could not find load balancer DNS"
        return 1
    fi
}

# Database maintenance function
database_maintenance() {
    log_info "Starting database maintenance..."

    # Check database status
    DB_STATUS=$(aws rds describe-db-instances \
        --db-instance-identifier "$DB_INSTANCE" \
        --query 'DBInstances[0].DBInstanceStatus' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "unknown")

    if [[ "$DB_STATUS" != "available" ]]; then
        log_error "Database is not available. Current status: $DB_STATUS"
        return 1
    fi

    log_info "Database status: $DB_STATUS"

    # Check for pending updates
    PENDING_UPDATES=$(aws rds describe-db-instances \
        --db-instance-identifier "$DB_INSTANCE" \
        --query 'DBInstances[0].PendingModifiedValues' \
        --output json \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ "$PENDING_UPDATES" != "null" && "$PENDING_UPDATES" != "{}" ]]; then
        log_warn "Database has pending updates: $PENDING_UPDATES"
    fi

    # Create manual snapshot before maintenance
    SNAPSHOT_ID="${DB_INSTANCE}-maintenance-$(date +%Y%m%d-%H%M%S)"
    log_info "Creating database snapshot: $SNAPSHOT_ID"

    aws rds create-db-snapshot \
        --db-instance-identifier "$DB_INSTANCE" \
        --db-snapshot-identifier "$SNAPSHOT_ID" \
        --region "$AWS_REGION" >/dev/null

    if [[ $? -eq 0 ]]; then
        log_info "Database snapshot created successfully"
    else
        log_error "Failed to create database snapshot"
        return 1
    fi

    # Check database performance metrics
    log_info "Checking database performance metrics..."

    CPU_UTILIZATION=$(aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name CPUUtilization \
        --dimensions Name=DBInstanceIdentifier,Value="$DB_INSTANCE" \
        --start-time "$(date -d '1 hour ago' -u +%Y-%m-%dT%H:%M:%S)" \
        --end-time "$(date -u +%Y-%m-%dT%H:%M:%S)" \
        --period 3600 \
        --statistics Average \
        --query 'Datapoints[0].Average' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "0")

    if (( $(echo "$CPU_UTILIZATION > 80" | bc -l) )); then
        log_warn "Database CPU utilization is high: ${CPU_UTILIZATION}%"
    else
        log_info "Database CPU utilization: ${CPU_UTILIZATION}%"
    fi

    log_info "Database maintenance completed"
}

# Redis cache maintenance function
redis_maintenance() {
    log_info "Starting Redis cache maintenance..."

    # Check Redis cluster status
    REDIS_STATUS=$(aws elasticache describe-cache-clusters \
        --cache-cluster-id "$REDIS_CLUSTER" \
        --query 'CacheClusters[0].CacheClusterStatus' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "unknown")

    log_info "Redis cluster status: $REDIS_STATUS"

    if [[ "$REDIS_STATUS" != "available" ]]; then
        log_error "Redis cluster is not available. Current status: $REDIS_STATUS"
        return 1
    fi

    # Check Redis performance metrics
    REDIS_CPU=$(aws cloudwatch get-metric-statistics \
        --namespace AWS/ElastiCache \
        --metric-name CPUUtilization \
        --dimensions Name=CacheClusterId,Value="$REDIS_CLUSTER" \
        --start-time "$(date -d '1 hour ago' -u +%Y-%m-%dT%H:%M:%S)" \
        --end-time "$(date -u +%Y-%m-%dT%H:%M:%S)" \
        --period 3600 \
        --statistics Average \
        --query 'Datapoints[0].Average' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "0")

    log_info "Redis CPU utilization: ${REDIS_CPU}%"

    REDIS_MEMORY=$(aws cloudwatch get-metric-statistics \
        --namespace AWS/ElastiCache \
        --metric-name DatabaseMemoryUsagePercentage \
        --dimensions Name=CacheClusterId,Value="$REDIS_CLUSTER" \
        --start-time "$(date -d '1 hour ago' -u +%Y-%m-%dT%H:%M:%S)" \
        --end-time "$(date -u +%Y-%m-%dT%H:%M:%S)" \
        --period 3600 \
        --statistics Average \
        --query 'Datapoints[0].Average' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "0")

    log_info "Redis memory usage: ${REDIS_MEMORY}%"

    if (( $(echo "$REDIS_MEMORY > 85" | bc -l) )); then
        log_warn "Redis memory usage is high: ${REDIS_MEMORY}%"
        log_info "Consider implementing cache eviction policies or scaling"
    fi

    log_info "Redis maintenance completed"
}

# ECS service maintenance function
ecs_maintenance() {
    log_info "Starting ECS service maintenance..."

    # Get current service status
    SERVICE_STATUS=$(aws ecs describe-services \
        --cluster "$ECS_CLUSTER" \
        --services "$ECS_SERVICE" \
        --query 'services[0].status' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "UNKNOWN")

    log_info "ECS service status: $SERVICE_STATUS"

    if [[ "$SERVICE_STATUS" != "ACTIVE" ]]; then
        log_error "ECS service is not active. Current status: $SERVICE_STATUS"
        return 1
    fi

    # Get task counts
    RUNNING_TASKS=$(aws ecs describe-services \
        --cluster "$ECS_CLUSTER" \
        --services "$ECS_SERVICE" \
        --query 'services[0].runningCount' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "0")

    DESIRED_TASKS=$(aws ecs describe-services \
        --cluster "$ECS_CLUSTER" \
        --services "$ECS_SERVICE" \
        --query 'services[0].desiredCount' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "0")

    log_info "ECS tasks - Running: $RUNNING_TASKS, Desired: $DESIRED_TASKS"

    if [[ "$RUNNING_TASKS" -lt "$DESIRED_TASKS" ]]; then
        log_warn "Running task count is less than desired count"
    fi

    # Check for stopped tasks in the last hour
    STOPPED_TASKS=$(aws ecs list-tasks \
        --cluster "$ECS_CLUSTER" \
        --service-name "$ECS_SERVICE" \
        --desired-status STOPPED \
        --query 'length(taskArns)' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null || echo "0")

    if [[ "$STOPPED_TASKS" -gt 0 ]]; then
        log_warn "Found $STOPPED_TASKS stopped tasks in the last period"

        # Get stop reasons for recent stopped tasks
        RECENT_TASKS=$(aws ecs list-tasks \
            --cluster "$ECS_CLUSTER" \
            --service-name "$ECS_SERVICE" \
            --desired-status STOPPED \
            --query 'taskArns[0:3]' \
            --output text \
            --region "$AWS_REGION" 2>/dev/null)

        if [[ -n "$RECENT_TASKS" ]]; then
            aws ecs describe-tasks \
                --cluster "$ECS_CLUSTER" \
                --tasks $RECENT_TASKS \
                --query 'tasks[].{TaskArn:taskArn,StopCode:stopCode,StoppedReason:stoppedReason}' \
                --output table \
                --region "$AWS_REGION" >> "$LOG_FILE"
        fi
    fi

    log_info "ECS maintenance completed"
}

# CloudWatch logs cleanup function
logs_cleanup() {
    log_info "Starting CloudWatch logs cleanup..."

    # Get log groups older than retention period
    LOG_GROUPS=$(aws logs describe-log-groups \
        --log-group-name-prefix "/ecs/${ENVIRONMENT}" \
        --query 'logGroups[?!retentionInDays].logGroupName' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$LOG_GROUPS" ]]; then
        log_info "Found log groups without retention policy: $LOG_GROUPS"

        # Set retention policy for log groups without one
        for log_group in $LOG_GROUPS; do
            log_info "Setting retention policy for $log_group"
            aws logs put-retention-policy \
                --log-group-name "$log_group" \
                --retention-in-days 30 \
                --region "$AWS_REGION" >/dev/null 2>&1
        done
    fi

    # Check log group sizes
    LARGE_LOG_GROUPS=$(aws logs describe-log-groups \
        --log-group-name-prefix "/ecs/${ENVIRONMENT}" \
        --query 'logGroups[?storedBytes > `1073741824`].{Name:logGroupName,Size:storedBytes}' \
        --output table \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$LARGE_LOG_GROUPS" && "$LARGE_LOG_GROUPS" != "[]" ]]; then
        log_warn "Found large log groups (>1GB):"
        echo "$LARGE_LOG_GROUPS" >> "$LOG_FILE"
    fi

    log_info "CloudWatch logs cleanup completed"
}

# Security checks function
security_checks() {
    log_info "Starting security checks..."

    # Check for unused security groups
    UNUSED_SG=$(aws ec2 describe-security-groups \
        --filters "Name=group-name,Values=*${ENVIRONMENT}*" \
        --query 'SecurityGroups[?!length(IpPermissions) && !length(IpPermissionsEgress[?IpProtocol!=`-1`])].GroupId' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$UNUSED_SG" ]]; then
        log_warn "Found potentially unused security groups: $UNUSED_SG"
    fi

    # Check for open security group rules
    OPEN_SG=$(aws ec2 describe-security-groups \
        --filters "Name=group-name,Values=*${ENVIRONMENT}*" \
        --query 'SecurityGroups[?IpPermissions[?IpRanges[?CidrIp==`0.0.0.0/0`]]].{GroupId:GroupId,GroupName:GroupName}' \
        --output table \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$OPEN_SG" && "$OPEN_SG" != "[]" ]]; then
        log_warn "Found security groups with open access (0.0.0.0/0):"
        echo "$OPEN_SG" >> "$LOG_FILE"
    fi

    # Check SSL certificate expiration
    CERT_ARN=$(aws acm list-certificates \
        --query 'CertificateSummaryList[?DomainName==`solemate.com`].CertificateArn' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$CERT_ARN" ]]; then
        CERT_EXPIRY=$(aws acm describe-certificate \
            --certificate-arn "$CERT_ARN" \
            --query 'Certificate.NotAfter' \
            --output text \
            --region "$AWS_REGION" 2>/dev/null)

        if [[ -n "$CERT_EXPIRY" ]]; then
            DAYS_TO_EXPIRY=$(( ($(date -d "$CERT_EXPIRY" +%s) - $(date +%s)) / 86400 ))
            log_info "SSL certificate expires in $DAYS_TO_EXPIRY days"

            if [[ "$DAYS_TO_EXPIRY" -lt 30 ]]; then
                log_warn "SSL certificate expires soon: $DAYS_TO_EXPIRY days"
            fi
        fi
    fi

    log_info "Security checks completed"
}

# Cost optimization checks function
cost_optimization() {
    log_info "Starting cost optimization checks..."

    # Check for unattached EBS volumes
    UNATTACHED_VOLUMES=$(aws ec2 describe-volumes \
        --filters "Name=status,Values=available" \
        --query 'Volumes[].{VolumeId:VolumeId,Size:Size,VolumeType:VolumeType}' \
        --output table \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$UNATTACHED_VOLUMES" && "$UNATTACHED_VOLUMES" != "[]" ]]; then
        log_warn "Found unattached EBS volumes:"
        echo "$UNATTACHED_VOLUMES" >> "$LOG_FILE"
    fi

    # Check for unused Elastic IPs
    UNUSED_EIP=$(aws ec2 describe-addresses \
        --query 'Addresses[?!AssociationId].AllocationId' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$UNUSED_EIP" ]]; then
        log_warn "Found unassociated Elastic IPs: $UNUSED_EIP"
    fi

    # Check RDS instance utilization
    RDS_CPU_AVG=$(aws cloudwatch get-metric-statistics \
        --namespace AWS/RDS \
        --metric-name CPUUtilization \
        --dimensions Name=DBInstanceIdentifier,Value="$DB_INSTANCE" \
        --start-time "$(date -d '7 days ago' -u +%Y-%m-%dT%H:%M:%S)" \
        --end-time "$(date -u +%Y-%m-%dT%H:%M:%S)" \
        --period 86400 \
        --statistics Average \
        --query 'Datapoints[].Average' \
        --output text \
        --region "$AWS_REGION" 2>/dev/null)

    if [[ -n "$RDS_CPU_AVG" ]]; then
        AVG_CPU=$(echo "$RDS_CPU_AVG" | awk '{sum+=$1; count++} END {print sum/count}')
        if (( $(echo "$AVG_CPU < 20" | bc -l) )); then
            log_warn "RDS instance may be over-provisioned. Average CPU: ${AVG_CPU}%"
        fi
    fi

    log_info "Cost optimization checks completed"
}

# Generate maintenance report function
generate_report() {
    log_info "Generating maintenance report..."

    REPORT_FILE="/tmp/solemate-maintenance-report-$(date +%Y%m%d-%H%M%S).json"

    cat << EOF > "$REPORT_FILE"
{
  "maintenance_run": {
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
    "environment": "$ENVIRONMENT",
    "duration_seconds": $(($(date +%s) - START_TIME)),
    "status": "completed",
    "log_file": "$LOG_FILE"
  },
  "health_checks": {
    "application": "$(check_service_health >/dev/null 2>&1 && echo 'passed' || echo 'failed')",
    "database": "$(database_maintenance >/dev/null 2>&1 && echo 'passed' || echo 'failed')",
    "redis": "$(redis_maintenance >/dev/null 2>&1 && echo 'passed' || echo 'failed')",
    "ecs": "$(ecs_maintenance >/dev/null 2>&1 && echo 'passed' || echo 'failed')"
  },
  "recommendations": [
    "Review CloudWatch alarms for any recent alerts",
    "Check application performance metrics",
    "Verify backup completion status",
    "Review security group configurations"
  ]
}
EOF

    log_info "Maintenance report generated: $REPORT_FILE"

    # Upload report to S3 if bucket exists
    S3_BUCKET="${ENVIRONMENT}-solemate-logs-$(aws sts get-caller-identity --query Account --output text)"
    if aws s3 ls "s3://$S3_BUCKET" >/dev/null 2>&1; then
        aws s3 cp "$REPORT_FILE" "s3://$S3_BUCKET/maintenance-reports/" --region "$AWS_REGION"
        log_info "Maintenance report uploaded to S3"
    fi
}

# Main execution function
main() {
    START_TIME=$(date +%s)

    log_info "Starting SoleMate maintenance automation"
    log_info "Environment: $ENVIRONMENT"
    log_info "AWS Region: $AWS_REGION"

    send_slack_notification "Starting maintenance automation for $ENVIRONMENT environment" "warning"

    # Pre-maintenance health check
    if ! check_service_health; then
        log_error "Pre-maintenance health check failed. Aborting maintenance."
        send_slack_notification "❌ Maintenance aborted due to failed health check" "danger"
        exit 1
    fi

    # Execute maintenance tasks
    database_maintenance
    redis_maintenance
    ecs_maintenance
    logs_cleanup
    security_checks
    cost_optimization

    # Post-maintenance health check
    sleep 30  # Wait for services to stabilize
    if check_service_health; then
        log_info "Post-maintenance health check passed"
        generate_report
        send_slack_notification "✅ Maintenance completed successfully for $ENVIRONMENT environment" "good"
    else
        log_error "Post-maintenance health check failed"
        send_slack_notification "⚠️ Maintenance completed but health check failed for $ENVIRONMENT environment" "danger"
    fi

    log_info "SoleMate maintenance automation completed"
    log_info "Log file: $LOG_FILE"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi