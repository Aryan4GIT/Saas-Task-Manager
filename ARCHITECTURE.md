# System Architecture

## Overview

This is a multi-tenant SaaS backend built with clean architecture principles, featuring JWT authentication, role-based access control, and AI-powered features.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         Client Layer                         │
│  (Web App, Mobile App, Third-party Integrations)            │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │ HTTPS/JSON
                     │
┌────────────────────▼────────────────────────────────────────┐
│                      API Gateway / Nginx                     │
│              (Load Balancing, SSL Termination)              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     │
┌────────────────────▼────────────────────────────────────────┐
│                   Go Gin Web Server                          │
│                    (Port 8080)                               │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Middleware Layer                         │  │
│  │  • CORS         • Logger      • Recovery             │  │
│  │  • JWT Auth     • Rate Limit  • Request ID           │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Handler Layer                            │  │
│  │  • AuthHandler    • TaskHandler                       │  │
│  │  • UserHandler    • IssueHandler                      │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Service Layer                            │  │
│  │  (Business Logic)                                     │  │
│  │  • AuthService    • TaskService                       │  │
│  │  • UserService    • IssueService                      │  │
│  │  • GeminiService  (AI Integration)                    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │            Repository Layer                           │  │
│  │  (Data Access)                                        │  │
│  │  • UserRepo       • TaskRepo                          │  │
│  │  • OrgRepo        • IssueRepo                         │  │
│  │  • TokenRepo      • AuditLogRepo                      │  │
│  └──────────────────────────────────────────────────────┘  │
└────────────────────┬────────────────────────────────────────┘
                     │
         ┌───────────┴───────────┬─────────────────┐
         │                       │                  │
┌────────▼────────┐   ┌──────────▼──────┐   ┌─────▼─────┐
│   PostgreSQL    │   │  Gemini API     │   │  AWS S3   │
│   (Supabase)    │   │  (Google AI)    │   │ (Future)  │
│                 │   │                 │   │           │
│ • Multi-tenant  │   │ • AI Summaries  │   │ • Files   │
│ • Indexes       │   │ • Text Analysis │   │ • Images  │
│ • Constraints   │   │                 │   │           │
└─────────────────┘   └─────────────────┘   └───────────┘
```

## Request Flow

### Authentication Flow

```
1. Client Registration/Login
   ↓
2. Handler validates request
   ↓
3. Service layer:
   - Validates credentials
   - Hashes password (bcrypt)
   - Generates JWT tokens
   ↓
4. Repository layer:
   - Stores user in database
   - Stores refresh token hash
   ↓
5. Returns access + refresh tokens
```

### Protected Resource Access Flow

```
1. Client sends request with JWT
   ↓
2. Auth Middleware:
   - Extracts token from header
   - Validates JWT signature
   - Checks expiration
   - Extracts user_id, org_id, role
   ↓
3. Role Middleware (if required):
   - Checks user role
   - Denies if insufficient permissions
   ↓
4. Handler:
   - Extracts context (user_id, org_id)
   - Passes to service layer
   ↓
5. Service layer:
   - Implements business logic
   - Calls repository
   ↓
6. Repository layer:
   - Builds SQL query with org_id filter
   - Executes parameterized query
   - Returns data scoped to organization
   ↓
7. Response sent back to client
```

## Layer Responsibilities

### Handler Layer (`internal/handler/`)
**Responsibility**: HTTP request/response handling

- Bind and validate JSON requests
- Extract authentication context
- Call appropriate service methods
- Format HTTP responses
- Handle HTTP errors

**Example**: `auth_handler.go`, `task_handler.go`

### Service Layer (`internal/service/`)
**Responsibility**: Business logic

- Validate business rules
- Coordinate between repositories
- Handle transactions
- Implement workflows
- Call external APIs (e.g., Gemini)
- Create audit logs

**Example**: `auth_service.go`, `task_service.go`

### Repository Layer (`internal/repository/`)
**Responsibility**: Data persistence

- Execute SQL queries
- Map database rows to models
- Handle database errors
- Ensure org_id scoping on all queries
- No business logic

**Example**: `user_repository.go`, `task_repository.go`

### Middleware Layer (`internal/middleware/`)
**Responsibility**: Cross-cutting concerns

- Authentication (JWT validation)
- Authorization (role checking)
- CORS handling
- Request logging
- Error recovery

**Example**: `auth.go`, `cors.go`, `logger.go`

## Data Model

```
┌─────────────────────┐
│   organizations     │
│ ─────────────────── │
│ id (PK)             │
│ name                │
│ slug (unique)       │
│ created_at          │
│ updated_at          │
└──────────┬──────────┘
           │
           │ 1:N
           │
┌──────────▼──────────┐
│       users         │
│ ─────────────────── │
│ id (PK)             │
│ org_id (FK)         │◄───────┐
│ email               │        │
│ password_hash       │        │
│ first_name          │        │
│ last_name           │        │
│ role                │        │
│ is_active           │        │
│ created_at          │        │
│ updated_at          │        │
└──────────┬──────────┘        │
           │                   │
           │ 1:N               │
           │                   │
     ┌─────┴────────┐          │
     │              │          │
┌────▼─────┐   ┌───▼──────┐   │
│  tasks   │   │  issues  │   │
│ ────────│   │ ──────── │   │
│ id (PK)  │   │ id (PK)  │   │
│ org_id   │   │ org_id   │   │
│ title    │   │ title    │   │
│ status   │   │ status   │   │
│ assigned │   │ severity │   │
│ created  │───┤ reported │───┘
└──────────┘   │ ai_summary
               └──────────┘

┌──────────────────────┐
│   refresh_tokens     │
│ ──────────────────── │
│ id (PK)              │
│ user_id (FK)         │
│ org_id (FK)          │
│ token_hash           │
│ expires_at           │
│ revoked_at           │
└──────────────────────┘

┌──────────────────────┐
│    audit_logs        │
│ ──────────────────── │
│ id (PK)              │
│ org_id (FK)          │
│ user_id (FK)         │
│ action               │
│ entity_type          │
│ entity_id            │
│ details (JSONB)      │
│ ip_address           │
│ created_at           │
└──────────────────────┘
```

## Multi-Tenancy Strategy

### Org-ID Scoping
Every query includes `WHERE org_id = $1` to ensure data isolation:

```go
// Good - Scoped by org_id
SELECT * FROM tasks WHERE org_id = $1 AND status = $2

// Bad - Missing org_id filter
SELECT * FROM tasks WHERE status = $1  // ❌ Security risk!
```

### JWT Token Structure
```json
{
  "user_id": "uuid",
  "org_id": "uuid",
  "role": "admin|manager|member",
  "exp": 1234567890,
  "iat": 1234567890
}
```

### Context Propagation
```
Request → Middleware → Context → Handler → Service → Repository
                ↓
        Sets: user_id, org_id, role
```

## Security Layers

### 1. Authentication Layer
- JWT token validation
- Token expiration checking
- Refresh token rotation
- Password hashing (bcrypt)

### 2. Authorization Layer
- Role-based access control (RBAC)
- Route-level permission checking
- Admin/Manager/Member roles

### 3. Data Isolation Layer
- Mandatory org_id on all queries
- No cross-organization data access
- Database-level constraints

### 4. Input Validation Layer
- JSON schema validation
- SQL injection prevention (parameterized queries)
- XSS prevention
- CORS configuration

## Scalability Considerations

### Current Architecture
- Single server deployment
- Connection pooling (25 max connections)
- Suitable for 1,000-10,000 users

### Future Scaling Options

1. **Horizontal Scaling**
   ```
   Load Balancer
        ↓
   [API Server 1] [API Server 2] [API Server 3]
        ↓              ↓              ↓
        └──────────────┴──────────────┘
                      ↓
              PostgreSQL (RDS)
   ```

2. **Caching Layer**
   - Add Redis for session management
   - Cache frequently accessed data
   - Reduce database load

3. **Database Optimization**
   - Read replicas for queries
   - Write to primary, read from replicas
   - Connection pooling per instance

4. **Microservices (Future)**
   - Auth Service
   - Task Service
   - Issue Service
   - Notification Service

## Deployment Architecture (AWS)

```
Internet
    ↓
┌───────────────────────┐
│  Route 53 (DNS)       │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│  CloudFront (CDN)     │
└──────────┬────────────┘
           ↓
┌───────────────────────┐
│  ALB (Load Balancer)  │
└──────────┬────────────┘
           ↓
┌─────────────────────────────────┐
│      VPC (Virtual Network)       │
│                                  │
│  ┌────────────────────────────┐ │
│  │  Public Subnet             │ │
│  │  ┌──────────────────────┐  │ │
│  │  │  EC2 (App Server)    │  │ │
│  │  │  - Docker           │  │ │
│  │  │  - Gin App          │  │ │
│  │  └──────────────────────┘  │ │
│  └────────────────────────────┘ │
│                                  │
│  ┌────────────────────────────┐ │
│  │  Private Subnet            │ │
│  │  ┌──────────────────────┐  │ │
│  │  │  RDS PostgreSQL      │  │ │
│  │  │  - Multi-AZ          │  │ │
│  │  │  - Automated backups │  │ │
│  │  └──────────────────────┘  │ │
│  └────────────────────────────┘ │
└─────────────────────────────────┘
```

## Technology Choices Explained

| Technology | Purpose | Why? |
|------------|---------|------|
| **Golang** | Backend language | Fast, concurrent, simple, great for APIs |
| **Gin** | Web framework | Lightweight, fast, good middleware support |
| **PostgreSQL** | Database | ACID compliant, great for relational data, JSONB support |
| **JWT** | Authentication | Stateless, scalable, standard |
| **bcrypt** | Password hashing | Secure, adaptive cost factor |
| **Gemini API** | AI summaries | Powerful, cost-effective, easy integration |
| **Docker** | Containerization | Consistent environments, easy deployment |

## Interview Talking Points

1. **Clean Architecture**: Separation of concerns, easy to test and maintain
2. **Multi-tenancy**: Data isolation using org_id scoping
3. **Security**: Multiple layers (auth, authz, input validation)
4. **Scalability**: Stateless design, can scale horizontally
5. **Production-ready**: Error handling, logging, audit trails
6. **Simple Code**: Easy to understand and explain, no over-engineering

## Monitoring & Observability (Future)

```
┌─────────────────────────────────────────┐
│          Application Metrics            │
│  • Request rate                          │
│  • Response times                        │
│  • Error rates                           │
│  • Database query performance            │
└──────────────┬──────────────────────────┘
               │
               ↓
┌──────────────────────────────────────────┐
│      CloudWatch / Prometheus             │
│  • Dashboards                             │
│  • Alerts                                 │
│  • Log aggregation                        │
└──────────────────────────────────────────┘
```

---

This architecture provides a solid foundation for a production-ready SaaS application while remaining simple enough to explain in technical interviews.
