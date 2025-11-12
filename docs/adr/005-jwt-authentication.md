# ADR-005: JWT-based Authentication

## Status
Accepted

## Date
2024-01-25

## Context

Our distributed microservices architecture requires a stateless authentication mechanism that:
- Works across multiple services without shared session storage
- Scales horizontally without session synchronization
- Supports mobile and web clients
- Enables fine-grained authorization
- Provides token refresh capabilities

We evaluated several authentication approaches:

| Approach | Pros | Cons |
|----------|------|------|
| **Session Cookies** | Simple, well-understood | Requires shared storage, doesn't work well with mobile |
| **JWT** | Stateless, portable, self-contained | Token size, revocation challenges |
| **OAuth 2.0** | Standard, delegated auth | Complex for internal auth |
| **API Keys** | Simple for service-to-service | No user context, no expiration |

## Decision

We will use **JWT (JSON Web Tokens)** for authentication with the following implementation:

### Token Structure

1. **Access Token** (short-lived: 24 hours)
   ```json
   {
     "header": {
       "alg": "HS256",
       "typ": "JWT"
     },
     "payload": {
       "sub": "user-uuid",
       "email": "user@example.com",
       "name": "John Doe",
       "role": "admin",
       "tenant_id": "tenant-uuid",
       "iat": 1234567890,
       "exp": 1234654290
     }
   }
   ```

2. **Refresh Token** (long-lived: 7 days)
   ```json
   {
     "sub": "user-uuid",
     "type": "refresh",
     "iat": 1234567890,
     "exp": 1235172690
   }
   ```

### Implementation Details

1. **Token Generation**
   ```go
   func GenerateTokens(user *User) (*Tokens, error) {
       accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
           "sub": user.ID,
           "email": user.Email,
           "role": user.Role,
           "tenant_id": user.TenantID,
           "exp": time.Now().Add(24 * time.Hour).Unix(),
       })

       refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
           "sub": user.ID,
           "type": "refresh",
           "exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
       })

       return &Tokens{
           AccessToken: accessToken.SignedString(jwtSecret),
           RefreshToken: refreshToken.SignedString(jwtSecret),
       }
   }
   ```

2. **Authentication Flow**
   ```
   Client                 Auth Service              Resource Service
     |                         |                           |
     |--POST /auth/login------>|                           |
     |                         |                           |
     |<--{access, refresh}-----|                           |
     |                         |                           |
     |--GET /users (+ Bearer)->|                           |
     |                         |                           |
     |                         |--Verify JWT-------------->|
     |                         |                           |
     |<------------------------|---------User data---------|
   ```

3. **Token Refresh Flow**
   ```
   Client                 Auth Service
     |                         |
     |--POST /auth/refresh---->|
     |   {refresh_token}       |
     |                         |
     |<--{new_access_token}----|
   ```

### Security Measures

1. **Token Signing**
   - Algorithm: HS256 (HMAC with SHA-256)
   - Secret: 32+ character random string from environment
   - Regular secret rotation (every 90 days)

2. **Token Validation**
   - Signature verification
   - Expiration check
   - Tenant validation
   - Role-based access control

3. **Storage**
   - **Web**: httpOnly cookies (XSS protection)
   - **Mobile**: Secure storage (Keychain/Keystore)
   - **Never**: localStorage (XSS vulnerable)

4. **Transmission**
   - HTTPS only in production
   - Authorization header: `Bearer <token>`

## Consequences

### Positive

- **Stateless**: No server-side session storage needed
- **Scalable**: Services can verify tokens independently
- **Performance**: No database lookup on every request
- **Cross-domain**: Works with CORS and mobile apps
- **Self-contained**: All user info in token
- **Microservices-friendly**: Each service can validate independently
- **Standard**: Well-supported libraries and tools

### Negative

- **Token Size**: Larger than session IDs (~200-300 bytes)
- **Revocation**: Cannot invalidate token before expiration
- **Secret Management**: Must protect JWT secret
- **Token Theft**: Stolen token valid until expiration
- **Payload Exposure**: Base64-encoded, not encrypted
- **Clock Sync**: Services must have synchronized clocks

### Mitigation Strategies

1. **Token Revocation**
   ```go
   // Maintain blacklist for revoked tokens (rare operation)
   type TokenBlacklist struct {
       Token     string
       ExpiresAt time.Time
   }

   // Check on critical operations only
   if blacklist.IsRevoked(token) {
       return ErrUnauthorized
   }
   ```

2. **Short Expiration**
   - Access token: 24 hours (balance security/UX)
   - Refresh token: 7 days
   - Critical operations: Re-authenticate

3. **Sensitive Data**
   - Don't put sensitive data in JWT
   - Use user ID and fetch fresh data when needed

4. **Token Rotation**
   ```go
   // Issue new refresh token on refresh
   func RefreshTokens(refreshToken string) (*Tokens, error) {
       // Verify old refresh token
       // Issue new access token AND new refresh token
       // Invalidate old refresh token
   }
   ```

## Authorization

### Role-Based Access Control (RBAC)

```go
type Role string

const (
    RoleAdmin  Role = "admin"   // Full access
    RoleUser   Role = "user"    // Standard access
    RoleViewer Role = "viewer"  // Read-only
)

// Middleware
func RequireRole(role Role) gin.HandlerFunc {
    return func(c *gin.Context) {
        claims := c.MustGet("claims").(jwt.MapClaims)
        if claims["role"] != string(role) {
            c.AbortWithStatus(http.StatusForbidden)
            return
        }
        c.Next()
    }
}

// Usage
router.DELETE("/users/:id", RequireRole(RoleAdmin), deleteUser)
```

## Token Refresh Strategy

1. **Frontend**: Check token expiration before requests
2. **Refresh**: If < 5 min remaining, refresh token
3. **Retry**: If 401, try refresh and retry original request
4. **Logout**: If refresh fails, redirect to login

```typescript
// Axios interceptor
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response?.status === 401) {
      const refreshed = await refreshToken()
      if (refreshed) {
        return axios.request(error.config)
      }
      logout()
    }
    return Promise.reject(error)
  }
)
```

## Alternatives Considered

### Session-Based Authentication
**Rejected** because:
- Requires Redis or database for session storage
- Doesn't scale well across multiple instances
- Complex with microservices
- Poor mobile app support

### OAuth 2.0
**Partially adopted**:
- Too complex for primary authentication
- May add for third-party integrations later
- JWT is OAuth 2.0 compatible

## Related Decisions

- [ADR-001](./001-microservices-architecture.md) - Microservices Architecture
- [ADR-002](./002-multi-tenancy-strategy.md) - Multi-tenancy Strategy

## References

- [RFC 7519: JSON Web Token (JWT)](https://tools.ietf.org/html/rfc7519)
- [JWT.io](https://jwt.io/)
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
