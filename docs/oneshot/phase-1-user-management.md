# Phase 1 User Management: Minimal Token-Based Authentication

## Overview

Simple user management with API token authentication. No passwords, sessions, access control, or admin interfaces. Just basic user creation and token-based API authentication.

## Database Schema

```sql
-- Simple user table (UUID v7 generated in application)
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    is_staff BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- API tokens for authentication (UUID v7 generated in application)
CREATE TABLE user_tokens (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_users_username ON users (username);
CREATE INDEX idx_user_tokens_token ON user_tokens (token);
```

## API Design

### Authentication

```go
// HTTP Bearer token authentication
// Authorization: Bearer <token>

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := extractBearerToken(r.Header.Get("Authorization"))
        if token == "" {
            http.Error(w, "Missing token", http.StatusUnauthorized)
            return
        }
        
        user, err := getUserByToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        
        ctx := context.WithValue(r.Context(), "user", user)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

## Token Generation

```go
func GenerateToken() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
```

## CLI Commands

### Create Admin User

```bash
# Command to create initial staff user
go run cmd/create-staff-user/main.go \
  --username admin
```

```go
// cmd/create-staff-user/main.go
func main() {
    flag.Parse()
    
    user := &User{
        Username: *username,
        IsStaff:  true,
    }
    
    userID, err := createUser(user)
    if err != nil {
        log.Fatal(err)
    }
    
    token := generateToken()
    err = createToken(&UserToken{
        UserID: userID,
        Token:  token,
        Name:   "Staff Token",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Staff user created.\nToken: %s\n", token)
}
```

## SQL Queries

### User Queries

```sql
-- name: GetUserByToken :one
SELECT u.id, u.username, u.is_staff
FROM users u
JOIN user_tokens t ON u.id = t.user_id
WHERE t.token = $1 
  AND (t.expires_at IS NULL OR t.expires_at > NOW());

-- name: CreateUser :one
INSERT INTO users (username, is_staff)
VALUES ($1, $2)
RETURNING id, username, is_staff, created_at;

-- name: CreateToken :one
INSERT INTO user_tokens (user_id, token, name, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, token, name, expires_at, created_at;
```

## Implementation Steps

### Week 1: Database and CLI
- Create user and token tables
- Implement create-admin-user command
- Basic token generation

### Week 2: API Authentication  
- Token authentication middleware
- User profile endpoint
- Integration with booking endpoints

---

This minimal user management provides just enough functionality for secure API access without unnecessary complexity.