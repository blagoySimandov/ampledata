# Debug Authentication Bypass

## Overview

A clean debug authentication bypass has been implemented for local development and testing. This allows you to test API endpoints without needing a valid WorkOS JWT token.

## How It Works

When `DEBUG_AUTH_BYPASS=true` is set, the authentication middleware automatically injects a fake debug user into all requests:

```go
{
  ID:        "debug-user-id",
  Email:     "debug@localhost",
  FirstName: "Debug",
  LastName:  "User",
}
```

**Key Features:**
- ✅ **Zero code changes needed** - All handlers work normally
- ✅ **Single point of control** - Enabled/disabled via one environment variable
- ✅ **No scattered if/else statements** - Clean implementation in middleware only
- ✅ **Transparent to handlers** - They get a user from context as usual

## Usage

### 1. Enable Debug Bypass

Add to your `.env` file or export in your shell:

```bash
export DEBUG_AUTH_BYPASS=true
```

### 2. Start the Server

```bash
go run ./cmd/server/main.go
```

You'll see this warning at startup:
```
⚠️  DEBUG_AUTH_BYPASS enabled - authentication disabled for development
```

### 3. Make API Calls Without Auth

Now you can call APIs without the `Authorization` header:

```bash
# Upload a file for enrichment
curl -X POST http://localhost:8080/api/v1/enrichment-signed-url

# List jobs
curl http://localhost:8080/api/v1/jobs

# Start a job
curl -X POST http://localhost:8080/api/v1/jobs/YOUR_JOB_ID/start

# Check progress
curl http://localhost:8080/api/v1/jobs/YOUR_JOB_ID/progress

# Get results
curl http://localhost:8080/api/v1/jobs/YOUR_JOB_ID/results
```

All requests will be authenticated as the debug user (`debug-user-id`).

## Implementation Details

### Files Modified

1. **`internal/config/config.go`**
   - Added `DebugAuthBypass bool` field to Config struct
   - Added `getEnvBool()` helper function
   - Reads from `DEBUG_AUTH_BYPASS` environment variable

2. **`internal/auth/middleware.go`**
   - Added `debugBypass bool` field to `JWTVerifier` struct
   - Modified `NewJWTVerifier()` to accept and store bypass flag
   - Modified `Middleware()` to check bypass flag and inject debug user
   - Skips JWKS fetching when bypass is enabled

3. **`cmd/server/main.go`**
   - Passes `cfg.DebugAuthBypass` to `auth.NewJWTVerifier()`

### Architecture

The bypass is implemented at the middleware level:

```
Request → Middleware → (bypass check) → inject debug user → Handler
                              ↓
                         Normal auth flow (if bypass disabled)
```

**No handler code changes needed** - They all use `auth.GetUserFromRequest(r)` which returns the injected debug user when bypass is enabled.

## Security Warnings

⚠️ **NEVER enable this in production!**

- This completely disables authentication
- Anyone can access all API endpoints
- All requests are attributed to the same debug user
- Only use in local development environments

## Disable the Bypass

To disable, either:

1. Remove the environment variable:
   ```bash
   unset DEBUG_AUTH_BYPASS
   ```

2. Set it to false:
   ```bash
   export DEBUG_AUTH_BYPASS=false
   ```

3. Don't set it at all (defaults to `false`)

## Testing with Real Auth

If you prefer to test with real WorkOS authentication instead of using the bypass, you would need to:

1. Set up a WorkOS account and configure SSO
2. Obtain a valid JWT token from WorkOS
3. Include it in your API calls:
   ```bash
   curl -H "Authorization: Bearer YOUR_JWT_TOKEN" \
        http://localhost:8080/api/v1/jobs
   ```

The bypass makes local development much easier by eliminating this requirement.
