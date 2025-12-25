# Multi-Tenant SaaS Backend

A production-ready multi-tenant SaaS backend built with Golang, Gin, PostgreSQL, and JWT authentication. Features include role-based access control, multi-tenancy using org_id scoping, and AI-powered issue summaries using Google's Gemini API.

## Tech Stack

- **Backend**: Golang 1.21+ with Gin framework
- **Database**: PostgreSQL 15+ (Supabase compatible)
- **Authentication**: JWT (access + refresh tokens)
- **Authorization**: Role-based access control (admin, manager, member)
- **AI Integration**: Google Gemini API for issue summaries
- **Deployment**: Docker, AWS EC2

## Features

- ✅ Clean layered architecture (handler → service → repository)
- ✅ Multi-tenancy with org_id scoping on all queries
- ✅ JWT authentication with access and refresh tokens
- ✅ Role-based access control (admin, manager, member)
- ✅ Password hashing with bcrypt
- ✅ PostgreSQL with proper indexes and relationships
- ✅ AI-powered issue summaries using Gemini API
- ✅ Audit logging for all operations
- ✅ CORS support
- ✅ Docker support
- ✅ Production-ready error handling

## Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── config/
│   └── config.go             # Configuration management
├── database/
│   ├── database.go           # Database connection
│   ├── schema.sql            # Database schema
│   └── seed.sql              # Sample data
├── internal/
│   ├── handler/              # HTTP handlers
│   ├── middleware/           # Middleware (auth, CORS, logger)
│   ├── models/               # Data models and DTOs
│   ├── repository/           # Database access layer
│   ├── router/               # Route definitions
│   ├── service/              # Business logic
│   └── utils/                # Utilities (JWT, crypto)
├── .env.example              # Environment variables template
├── docker-compose.yml        # Docker composition
├── Dockerfile                # Docker image definition
└── go.mod                    # Go dependencies
```

## Database Schema

### Tables
- **organizations**: Multi-tenant organization data
- **users**: User accounts with org_id scoping
- **refresh_tokens**: JWT refresh token storage
- **tasks**: Task management with assignment
- **issues**: Issue tracking with AI summaries
- **audit_logs**: Complete audit trail

All tables include `org_id` for multi-tenancy isolation.

## Quick Start

### Prerequisites
- Go 1.21+
- PostgreSQL 15+
- Docker (optional)

### Local Development

1. **Clone and setup**
```bash
git clone <repository-url>
cd portfolio
cp .env.example .env
```

2. **Configure environment variables**
Edit `.env` with your configuration:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=saas_db
JWT_ACCESS_SECRET=your-secret-key
JWT_REFRESH_SECRET=your-refresh-key
GEMINI_API_KEY=your-gemini-key
```

3. **Setup database**
```bash
# Create database
psql -U postgres -c "CREATE DATABASE saas_db;"

# Run migrations
psql -U postgres -d saas_db -f database/schema.sql
psql -U postgres -d saas_db -f database/seed.sql
```

4. **Install dependencies**
```bash
go mod download
```

5. **Run the server**
```bash
go run cmd/server/main.go
```

Server will start on `http://localhost:8080`

### Docker Deployment

1. **Using Docker Compose**
```bash
# Copy environment file
cp .env.docker .env

# Update .env with your secrets
# Then start services
docker-compose up -d
```

2. **Build and run manually**
```bash
# Build image
docker build -t saas-backend .

# Run container
docker run -p 8080:8080 --env-file .env saas-backend
```

## API Documentation

### Authentication

#### Register (Create Organization)
```bash
POST /api/v1/auth/register
Content-Type: application/json

{
  "org_name": "Acme Corp",
  "email": "admin@acme.com",
  "password": "password123",
  "first_name": "John",
  "last_name": "Doe"
}
```

#### Login
```bash
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "admin@acme.com",
  "password": "password123"
}
```

#### Refresh Token
```bash
POST /api/v1/auth/refresh
Content-Type: application/json

{
  "refresh_token": "your-refresh-token"
}
```

### Tasks

#### Create Task
```bash
POST /api/v1/tasks
Authorization: Bearer <access-token>
Content-Type: application/json

{
  "title": "Implement feature X",
  "description": "Add new feature",
  "priority": "high",
  "assigned_to": "user-uuid",
  "due_date": "2024-12-31T23:59:59Z"
}
```

#### List Tasks
```bash
GET /api/v1/tasks?status=todo
Authorization: Bearer <access-token>
```

#### Update Task
```bash
PATCH /api/v1/tasks/:id
Authorization: Bearer <access-token>
Content-Type: application/json

{
  "status": "in_progress"
}
```

### Issues

#### Create Issue (with AI Summary)
```bash
POST /api/v1/issues
Authorization: Bearer <access-token>
Content-Type: application/json

{
  "title": "Login page not responsive",
  "description": "The login page doesn't work on mobile devices",
  "severity": "high",
  "assigned_to": "user-uuid"
}
```

#### List Issues
```bash
GET /api/v1/issues?status=open
Authorization: Bearer <access-token>
```

### Users (Admin/Manager only)

#### Create User
```bash
POST /api/v1/users
Authorization: Bearer <access-token>
Content-Type: application/json

{
  "email": "user@acme.com",
  "password": "password123",
  "first_name": "Jane",
  "last_name": "Smith",
  "role": "member"
}
```

#### List Users
```bash
GET /api/v1/users
Authorization: Bearer <access-token>
```

## AWS EC2 Deployment

### 1. Launch EC2 Instance
- Instance type: t2.micro (free tier)
- AMI: Ubuntu 22.04 LTS
- Security group: Allow ports 22 (SSH), 80 (HTTP), 443 (HTTPS), 8080

### 2. Install Dependencies
```bash
# SSH into EC2 instance
ssh -i your-key.pem ubuntu@your-ec2-ip

# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker ubuntu

# Install Docker Compose
sudo apt install docker-compose -y
```

### 3. Deploy Application
```bash
# Clone repository
git clone <repository-url>
cd portfolio

# Configure environment
cp .env.example .env
nano .env  # Update with production values

# Start services
docker-compose up -d

# Check logs
docker-compose logs -f
```

### 4. Setup Nginx (Optional)
```bash
# Install Nginx
sudo apt install nginx -y

# Configure reverse proxy
sudo nano /etc/nginx/sites-available/saas-backend
```

Add configuration:
```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/saas-backend /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### 5. Setup SSL with Let's Encrypt
```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your-domain.com
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `ENV` | Environment (development/production) | `development` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_USER` | Database user | `postgres` |
| `DB_PASSWORD` | Database password | - |
| `DB_NAME` | Database name | `saas_db` |
| `JWT_ACCESS_SECRET` | JWT access token secret | - |
| `JWT_REFRESH_SECRET` | JWT refresh token secret | - |
| `JWT_ACCESS_EXPIRY` | Access token expiry | `15m` |
| `JWT_REFRESH_EXPIRY` | Refresh token expiry | `168h` |
| `GEMINI_API_KEY` | Google Gemini API key | - |
| `GEMINI_MODEL` | Gemini model name (e.g. `gemini-2.5-flash`) | `gemini-2.5-flash` |
| `ALLOWED_ORIGINS` | CORS allowed origins | `http://localhost:3000` |

## Security Features

- **Password Hashing**: bcrypt with default cost
- **JWT Tokens**: Separate access and refresh tokens
- **Token Rotation**: Refresh tokens are rotated on use
- **Multi-tenancy**: All queries scoped by org_id
- **RBAC**: Role-based permissions (admin, manager, member)
- **Audit Logging**: Complete audit trail of all operations
- **SQL Injection Protection**: Parameterized queries

## Testing

### Sample Test Users
Default password for all users: `password123`

- **Admin**: admin@acme.com (full access)
- **Manager**: manager@acme.com (user management + tasks/issues)
- **Member**: member@acme.com (tasks/issues only)

### Health Check
```bash
curl http://localhost:8080/health
```

## Development

### Run Tests
```bash
go test ./...
```

### Build Binary
```bash
go build -o saas-backend ./cmd/server
```

### Database Migrations
To create a new migration, add SQL files to the `database/` directory and run them manually or use a migration tool.

## Production Checklist

- [ ] Change all default passwords and secrets
- [ ] Enable HTTPS/TLS
- [ ] Configure proper CORS origins
- [ ] Setup database backups
- [ ] Enable application logging
- [ ] Configure monitoring (CloudWatch, Prometheus, etc.)
- [ ] Setup rate limiting
- [ ] Configure proper security groups
- [ ] Use AWS RDS for database in production
- [ ] Setup CI/CD pipeline
- [ ] Configure environment-specific configs

## License

MIT

## Support

For issues and questions, please open an issue on GitHub.

---

Built with ❤️ using Golang, Gin, PostgreSQL, and Gemini AI
