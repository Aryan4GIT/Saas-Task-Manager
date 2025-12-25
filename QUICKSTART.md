# Quick Start Guide

## Getting Started in 5 Minutes

### Option 1: Docker (Recommended)

1. **Start the application**
```bash
# Clone the repository
git clone <repo-url>
cd portfolio

# Copy environment file
cp .env.docker .env

# Start services (PostgreSQL + API)
docker-compose up -d

# Check if services are running
docker-compose ps

# View logs
docker-compose logs -f app
```

The API will be available at `http://localhost:8080`

### Option 2: Local Development

1. **Setup PostgreSQL**
```bash
# Install PostgreSQL 15+
# Then create database
createdb saas_db

# Run migrations
psql -d saas_db -f database/schema.sql
psql -d saas_db -f database/seed.sql
```

2. **Configure environment**
```bash
cp .env.example .env
# Edit .env with your database credentials
```

3. **Run the application**
```bash
go mod download
go run cmd/server/main.go
```

## Test the API

### 1. Health Check
```bash
curl http://localhost:8080/health
```

### 2. Register New Organization
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "org_name": "My Company",
    "email": "admin@mycompany.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

Save the `access_token` from the response!

### 3. Login with Seeded Data
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@acme.com",
    "password": "password123"
  }'
```

### 4. Create a Task
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Build landing page",
    "description": "Create responsive landing page",
    "priority": "high"
  }'
```

### 5. Create an Issue (with AI Summary)
```bash
curl -X POST http://localhost:8080/api/v1/issues \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "App crashes on startup",
    "description": "The mobile app crashes immediately after opening on Android 14 devices",
    "severity": "critical"
  }'
```

### 6. List All Tasks
```bash
curl http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Default Test Accounts

After running the seed file, you'll have:

| Email | Password | Role | Organization |
|-------|----------|------|--------------|
| admin@acme.com | password123 | admin | Acme Corp |
| manager@acme.com | password123 | manager | Acme Corp |
| member@acme.com | password123 | member | Acme Corp |

## API Endpoints Overview

### Public Endpoints
- `POST /api/v1/auth/register` - Register new organization
- `POST /api/v1/auth/login` - Login
- `POST /api/v1/auth/refresh` - Refresh access token

### Protected Endpoints (Require Authentication)
- `GET /api/v1/auth/me` - Get current user
- `POST /api/v1/auth/logout` - Logout

#### Tasks
- `POST /api/v1/tasks` - Create task
- `GET /api/v1/tasks` - List tasks
- `GET /api/v1/tasks/my` - My assigned tasks
- `GET /api/v1/tasks/ai-report` - AI task report (admin only)
- `GET /api/v1/tasks/:id` - Get task
- `PATCH /api/v1/tasks/:id` - Update task
- `DELETE /api/v1/tasks/:id` - Delete task

#### Reports
- `GET /api/v1/reports/weekly-summary` - Weekly AI summary (admin/manager only)

#### Audit Logs
- `GET /api/v1/audit-logs?limit=50` - List audit logs (admin only)

#### Issues
- `POST /api/v1/issues` - Create issue (with AI summary)
- `GET /api/v1/issues` - List issues
- `GET /api/v1/issues/:id` - Get issue
- `PATCH /api/v1/issues/:id` - Update issue
- `DELETE /api/v1/issues/:id` - Delete issue

#### Users (Admin/Manager only)
- `POST /api/v1/users` - Create user
- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user
- `PATCH /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user (admin only)

## Role-Based Access

| Role | Permissions |
|------|-------------|
| **admin** | Full access to everything |
| **manager** | Can manage users, tasks, and issues |
| **member** | Can create/update own tasks and issues |

Notes:
- Members can only access their assigned tasks, and only issues they reported or are assigned to.
- Admin/Manager can access organization-wide tasks and issues.

## Multi-Tenancy

All data is automatically scoped to the user's organization (`org_id`). Users can only access data from their own organization.

## Gemini AI Integration

When creating an issue, the system automatically generates an AI-powered summary using Google's Gemini API. To enable:

1. Get API key from [Google AI Studio](https://makersuite.google.com/app/apikey)
2. Add to `.env`: `GEMINI_API_KEY=your-key-here`
3. (Optional) Pick a model: `GEMINI_MODEL=gemini-2.5-flash`
4. Create an issue - the `ai_summary` field will be automatically populated

## Stop the Application

### Docker
```bash
docker-compose down
```

### Local
Press `Ctrl+C` in the terminal

## Troubleshooting

### Database Connection Error
```bash
# Check if PostgreSQL is running
docker-compose ps
# Or for local:
pg_isready

# View database logs
docker-compose logs postgres
```

### Port Already in Use
```bash
# Change PORT in .env file
PORT=8081
```

### Can't Login
- Ensure database migrations have been run
- Check if user exists: `psql -d saas_db -c "SELECT * FROM users;"`
- Verify password is "password123" for seeded users

## Next Steps

1. **Frontend Integration**: Use the API with React, Vue, or Next.js
2. **Production Deployment**: Follow [README.md](README.md) for AWS EC2 setup
3. **Customize**: Modify models, add new endpoints, extend features
4. **Monitor**: Add logging, metrics, and monitoring

## Useful Commands

```bash
# View API logs
docker-compose logs -f app

# View database logs
docker-compose logs -f postgres

# Execute SQL query
docker-compose exec postgres psql -U postgres -d saas_db -c "SELECT * FROM organizations;"

# Restart services
docker-compose restart

# Rebuild after code changes
docker-compose up -d --build

# Access PostgreSQL shell
docker-compose exec postgres psql -U postgres -d saas_db
```

## Development Tips

1. **Use REST Client Extension**: Install REST Client in VS Code and use `api-tests.http`
2. **Watch Logs**: Keep logs open to see requests and errors
3. **Database GUI**: Use tools like DBeaver, pgAdmin, or TablePlus
4. **Hot Reload**: Use `air` for auto-reload during development

Happy coding! 🚀
