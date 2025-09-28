# SoleMate Troubleshooting Guide

## Database Authentication Issues

### Problem: Services failing with "password authentication failed"

**Symptoms:**
- Microservices constantly restarting
- Error: `failed SASL auth (FATAL: password authentication failed for user "solemate"`
- Services can't connect to PostgreSQL despite correct credentials

**Root Cause:**
PostgreSQL 15+ defaults to SCRAM-SHA-256 authentication, which may be incompatible with some Go PostgreSQL drivers.

**Solution Applied:**
1. **Updated Docker Compose configuration:**
   ```yaml
   postgres:
     environment:
       POSTGRES_HOST_AUTH_METHOD: md5
       POSTGRES_INITDB_ARGS: "--auth-host=md5"
   ```

2. **Custom pg_hba.conf configuration:**
   ```
   # Use MD5 for Docker container connections
   host all all all md5
   ```

3. **Manual fix for existing installations:**
   ```bash
   # Change authentication method in running container
   docker exec solemate-postgres sed -i 's/scram-sha-256/md5/g' /var/lib/postgresql/data/pg_hba.conf

   # Reset user password to MD5 encoding
   docker exec solemate-postgres psql -U solemate -d solemate_db -c "ALTER USER solemate PASSWORD 'password';"

   # Reload PostgreSQL configuration
   docker exec solemate-postgres su postgres -c 'pg_ctl reload -D /var/lib/postgresql/data'

   # Restart affected services
   docker-compose restart user-service product-service order-service
   ```

### Verification

Test database connectivity from external container:
```bash
docker run --rm --network solemate-network -e PGPASSWORD=password postgres:15 psql -h postgres -U solemate -d solemate_db -c "SELECT current_user;"
```

Expected result: Should connect successfully without authentication errors.

## API Gateway URL Issues

### Problem: "unsupported protocol scheme" errors

**Symptoms:**
- Error: `unsupported protocol scheme "product-service"`
- API Gateway cannot route requests to services

**Solution:**
Ensure all service URLs in API Gateway configuration include the `http://` protocol:

```yaml
api-gateway:
  environment:
    - USER_SERVICE_URL=http://user-service:8080      # ✅ Correct
    - PRODUCT_SERVICE_URL=http://product-service:8081 # ✅ Correct
    # Not: user-service:8080                         # ❌ Missing protocol
```

## Phone Number Validation Issues

### Problem: Phone field required despite being marked optional

**Solution:**
Updated validation schema to properly handle empty phone numbers:

```typescript
phone: yup
  .string()
  .transform((value) => (value === '' ? undefined : value))
  .matches(/^[\+]?[0-9\s\-\(\)\.]{7,20}$/, 'Please enter a valid phone number')
  .optional()
```

## Service Health Checks

Verify all services are running:
```bash
# Check service status
docker-compose ps

# Test individual service health
curl http://localhost:8080/health  # User Service
curl http://localhost:8081/health  # Product Service
curl http://localhost:8083/health  # Cart Service
curl http://localhost:8084/health  # Order Service
curl http://localhost:8000/health  # API Gateway

# Test API routing
curl "http://localhost:8000/api/v1/products?page=1&limit=3"
```

All services should return `{"service":"<service-name>","status":"healthy"}`.

## Frontend CORS Issues

### Problem: Frontend getting 403 errors on API calls

**Symptoms:**
- Browser console shows: `Access to fetch at 'http://localhost:8000/api/v1/products' from origin 'http://localhost:3002' has been blocked by CORS policy`
- API Gateway logs show 403 responses for OPTIONS requests
- Frontend unable to load data from backend

**Root Cause:**
Frontend running on port 3002 but CORS configuration only allows ports 3000-3001.

**Solution:**
Update CORS configuration in API Gateway to include the frontend port:

```go
// api-gateway/internal/middleware/cors.go
AllowOrigins: []string{
    "http://localhost:3000",
    "http://localhost:3001",
    "http://localhost:3002", // Add current frontend port
    "http://127.0.0.1:3000",
    "http://127.0.0.1:3001",
    "http://127.0.0.1:3002", // Add current frontend port
    // ... other origins
},
```

**Verification:**
```bash
# Test CORS preflight request
curl -X OPTIONS http://localhost:8000/api/v1/products \
  -H "Origin: http://localhost:3002" \
  -H "Access-Control-Request-Method: GET" -i

# Should return 204 No Content with proper CORS headers
```