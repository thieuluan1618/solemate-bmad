#!/bin/bash
set -e

# AWS Secrets Manager setup script for SoleMate production environment
# This script creates all necessary secrets for the production deployment

AWS_REGION=${AWS_REGION:-us-east-1}
ENVIRONMENT=${ENVIRONMENT:-production}

echo "Setting up AWS Secrets Manager secrets for SoleMate $ENVIRONMENT environment"
echo "Region: $AWS_REGION"

# Function to create or update secret
create_or_update_secret() {
    local secret_name=$1
    local secret_value=$2
    local description=$3

    echo "Processing secret: $secret_name"

    if aws secretsmanager describe-secret --secret-id "$secret_name" --region "$AWS_REGION" >/dev/null 2>&1; then
        echo "  Updating existing secret: $secret_name"
        aws secretsmanager update-secret \
            --secret-id "$secret_name" \
            --secret-string "$secret_value" \
            --region "$AWS_REGION"
    else
        echo "  Creating new secret: $secret_name"
        aws secretsmanager create-secret \
            --name "$secret_name" \
            --description "$description" \
            --secret-string "$secret_value" \
            --region "$AWS_REGION"
    fi
}

# Database configuration
echo "Setting up database secrets..."
create_or_update_secret \
    "solemate/db-host" \
    "${DB_HOST:-production-postgres.region.rds.amazonaws.com}" \
    "PostgreSQL database host for SoleMate production"

create_or_update_secret \
    "solemate/db-password" \
    "${DB_PASSWORD:-CHANGE_ME_SECURE_PASSWORD}" \
    "PostgreSQL database password for SoleMate production"

# Redis configuration
echo "Setting up Redis secrets..."
create_or_update_secret \
    "solemate/redis-host" \
    "${REDIS_HOST:-production-redis.cache.amazonaws.com}" \
    "Redis cache host for SoleMate production"

# JWT secrets
echo "Setting up JWT secrets..."
JWT_ACCESS_SECRET=$(openssl rand -base64 64 | tr -d "\\n")
JWT_REFRESH_SECRET=$(openssl rand -base64 64 | tr -d "\\n")

create_or_update_secret \
    "solemate/jwt-access-secret" \
    "$JWT_ACCESS_SECRET" \
    "JWT access token secret for SoleMate production"

create_or_update_secret \
    "solemate/jwt-refresh-secret" \
    "$JWT_REFRESH_SECRET" \
    "JWT refresh token secret for SoleMate production"

# Stripe payment secrets
echo "Setting up Stripe secrets..."
create_or_update_secret \
    "solemate/stripe-api-key" \
    "${STRIPE_API_KEY:-sk_live_CHANGE_ME}" \
    "Stripe API key for SoleMate production payments"

create_or_update_secret \
    "solemate/stripe-webhook-secret" \
    "${STRIPE_WEBHOOK_SECRET:-whsec_CHANGE_ME}" \
    "Stripe webhook secret for SoleMate production"

# Elasticsearch configuration
echo "Setting up Elasticsearch secrets..."
create_or_update_secret \
    "solemate/elasticsearch-url" \
    "${ELASTICSEARCH_URL:-https://production-elasticsearch.region.es.amazonaws.com}" \
    "Elasticsearch URL for SoleMate production search"

# Application secrets
echo "Setting up application secrets..."
ENCRYPTION_KEY=$(openssl rand -base64 32 | tr -d "\\n")
create_or_update_secret \
    "solemate/encryption-key" \
    "$ENCRYPTION_KEY" \
    "Application encryption key for SoleMate production"

# Email service configuration
echo "Setting up email service secrets..."
create_or_update_secret \
    "solemate/smtp-password" \
    "${SMTP_PASSWORD:-CHANGE_ME_SMTP_PASSWORD}" \
    "SMTP password for SoleMate production email service"

# Third-party API keys
echo "Setting up third-party API secrets..."
create_or_update_secret \
    "solemate/aws-access-key" \
    "${AWS_ACCESS_KEY_ID:-CHANGE_ME}" \
    "AWS access key for SoleMate production services"

create_or_update_secret \
    "solemate/aws-secret-key" \
    "${AWS_SECRET_ACCESS_KEY:-CHANGE_ME}" \
    "AWS secret key for SoleMate production services"

echo ""
echo "✅ All secrets have been created/updated successfully!"
echo ""
echo "⚠️  IMPORTANT SECURITY REMINDERS:"
echo "  1. Update all placeholder values (CHANGE_ME) with actual production values"
echo "  2. Ensure IAM roles have minimum required permissions to access these secrets"
echo "  3. Enable secret rotation for database passwords and API keys"
echo "  4. Monitor secret access through CloudTrail"
echo "  5. Use different secrets for different environments (dev/staging/prod)"
echo ""
echo "📋 Next steps:"
echo "  1. Update the CloudFormation template with correct secret ARNs"
echo "  2. Deploy the ECS infrastructure using the CloudFormation template"
echo "  3. Update the ECS task definition with correct image URIs"
echo "  4. Deploy the ECS service"
echo ""