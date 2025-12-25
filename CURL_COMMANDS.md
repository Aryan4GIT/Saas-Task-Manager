# 🔥 cURL Testing Commands - Copy & Paste Ready

## 🚀 Quick Start

### 1. Health Check
```bash
curl http://localhost:8080/health
```

---

## 🔐 Authentication

### 2. Login (Save the access_token!)
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"admin@acme.com\", \"password\": \"password123\"}"
```

**⚠️ Copy the `access_token` from response and replace `YOUR_TOKEN` below!**

### 3. Register New Organization
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d "{\"org_name\": \"My Company\", \"email\": \"admin@mycompany.com\", \"password\": \"password123\", \"first_name\": \"John\", \"last_name\": \"Doe\"}"
```

### 4. Get Current User
```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 5. Refresh Token
```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\": \"YOUR_REFRESH_TOKEN\"}"
```

### 6. Logout
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 📋 Tasks

### 7. Create Task
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"Build user dashboard\", \"description\": \"Create responsive dashboard\", \"priority\": \"high\"}"
```

### 8. List All Tasks
```bash
curl http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 9. List Tasks by Status
```bash
curl "http://localhost:8080/api/v1/tasks?status=todo" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 10. List My Assigned Tasks
```bash
curl http://localhost:8080/api/v1/tasks/my \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 11. Get Single Task
```bash
curl http://localhost:8080/api/v1/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 12. Update Task
```bash
curl -X PATCH http://localhost:8080/api/v1/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"status\": \"in_progress\"}"
```

### 13. Delete Task
```bash
curl -X DELETE http://localhost:8080/api/v1/tasks/TASK_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 🐛 Issues

### 14. Create Issue (with AI Summary!)
```bash
curl -X POST http://localhost:8080/api/v1/issues \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"title\": \"Login page crashes on Safari\", \"description\": \"Users on Safari browser experience crashes when trying to login. This affects around 15% of our user base.\", \"severity\": \"high\"}"
```

### 15. List All Issues
```bash
curl http://localhost:8080/api/v1/issues \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 16. List Issues by Status
```bash
curl "http://localhost:8080/api/v1/issues?status=open" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 17. Get Single Issue
```bash
curl http://localhost:8080/api/v1/issues/ISSUE_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 18. Update Issue
```bash
curl -X PATCH http://localhost:8080/api/v1/issues/ISSUE_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"status\": \"in_progress\", \"severity\": \"critical\"}"
```

### 19. Resolve Issue
```bash
curl -X PATCH http://localhost:8080/api/v1/issues/ISSUE_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"status\": \"resolved\"}"
```

### 20. Delete Issue
```bash
curl -X DELETE http://localhost:8080/api/v1/issues/ISSUE_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 👥 Users (Admin/Manager Only)

### 21. Create User
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"newuser@acme.com\", \"password\": \"password123\", \"first_name\": \"New\", \"last_name\": \"User\", \"role\": \"member\"}"
```

### 22. List All Users
```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 23. Get User by ID
```bash
curl http://localhost:8080/api/v1/users/USER_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 24. Update User
```bash
curl -X PATCH http://localhost:8080/api/v1/users/USER_ID \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"first_name\": \"Updated\", \"role\": \"manager\"}"
```

### 25. Delete User (Admin Only)
```bash
curl -X DELETE http://localhost:8080/api/v1/users/USER_ID \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## 📝 Windows PowerShell Version

If you're using PowerShell, use this format:

### Login (PowerShell)
```powershell
$body = @{
    email = "admin@acme.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri "http://localhost:8080/api/v1/auth/login" -Body $body -ContentType "application/json"
```

### Create Task (PowerShell)
```powershell
$token = "YOUR_TOKEN"
$body = @{
    title = "Build dashboard"
    description = "Create responsive dashboard"
    priority = "high"
} | ConvertTo-Json

Invoke-RestMethod -Method Post -Uri "http://localhost:8080/api/v1/tasks" -Headers @{Authorization = "Bearer $token"} -Body $body -ContentType "application/json"
```

### List Tasks (PowerShell)
```powershell
$token = "YOUR_TOKEN"
Invoke-RestMethod -Method Get -Uri "http://localhost:8080/api/v1/tasks" -Headers @{Authorization = "Bearer $token"}
```

---

## 🎯 Quick Testing Flow

```bash
# 1. Check health
curl http://localhost:8080/health

# 2. Login and copy token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d "{\"email\": \"admin@acme.com\", \"password\": \"password123\"}"

# 3. Export token (Linux/Mac)
export TOKEN="paste_your_token_here"

# 3. Or set variable (Windows CMD)
set TOKEN=paste_your_token_here

# 3. Or set variable (Windows PowerShell)
$TOKEN="paste_your_token_here"

# 4. Now use $TOKEN in requests (Linux/Mac/PowerShell)
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"

# Or %TOKEN% (Windows CMD)
curl http://localhost:8080/api/v1/auth/me ^
  -H "Authorization: Bearer %TOKEN%"
```

---

## 🔑 Test Accounts

| Email | Password | Role |
|-------|----------|------|
| admin@acme.com | password123 | admin |
| manager@acme.com | password123 | manager |
| member@acme.com | password123 | member |

---

## 💡 Pro Tips

1. **Save token as variable** to avoid copy-pasting
2. **Use `-v` flag** for verbose output: `curl -v http://...`
3. **Pretty print JSON**: Add `| jq` at the end (requires jq installed)
4. **Save response**: Add `-o response.json`
5. **See just status code**: Add `-w "%{http_code}\n" -o /dev/null -s`

---

**Ready to test! Just replace `YOUR_TOKEN` with your actual token from login response.** 🚀
