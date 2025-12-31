# 🧪 API Testing Guide - Thunder Client Free Version

## Quick Setup

### 1. Start the Server
```bash
go run cmd/server/main.go
```

### 2. Test with Thunder Client

Open Thunder Client in VS Code sidebar, then create requests manually:

---

## 📋 Step-by-Step Testing

### ✅ STEP 1: Health Check
```
Method: GET
URL: http://localhost:8080/health
```
**Expected**: `{"status": "healthy"}`

---

### ✅ STEP 2: Login (Get Access Token)
```
Method: POST
URL: http://localhost:8080/api/v1/auth/login
Headers:
  Content-Type: application/json
Body (JSON):
{
  "email": "admin@acme.com",
  "password": "password123"
}
```

**Copy the `access_token` from response!** You'll need it for all other requests.

**Expected Response**:
```json
{
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "user": {...}
}
```

---

### ✅ STEP 3: Get Current User
```
Method: GET
URL: http://localhost:8080/api/v1/auth/me
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
```

---

### ✅ STEP 4: Create Task
```
Method: POST
URL: http://localhost:8080/api/v1/tasks
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
  Content-Type: application/json
Body (JSON):
{
  "title": "Build user dashboard",
  "description": "Create responsive dashboard",
  "priority": "high"
}
```

**Copy the `id` from response** to use in next steps.

---

### ✅ STEP 5: List All Tasks
```
Method: GET
URL: http://localhost:8080/api/v1/tasks
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
```

---

### ✅ STEP 6: List Tasks by Status
```
Method: GET
URL: http://localhost:8080/api/v1/tasks?status=todo
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
```

---

### ✅ STEP 7: Update Task
```
Method: PATCH
URL: http://localhost:8080/api/v1/tasks/TASK_ID_HERE
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
  Content-Type: application/json
Body (JSON):
{
  "status": "in_progress"
}
```

---

### ✅ STEP 8: Create Issue (with AI Summary)
```
Method: POST
URL: http://localhost:8080/api/v1/issues
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
  Content-Type: application/json
Body (JSON):
{
  "title": "Login page crashes on Safari",
  "description": "Users on Safari browser experience crashes when trying to login. This affects around 15% of our user base on macOS.",
  "severity": "high"
}
```

**Look for `ai_summary` in response!**

---

### ✅ STEP 9: List Issues
```
Method: GET
URL: http://localhost:8080/api/v1/issues
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
```

---

### ✅ STEP 10: Create User (Admin only)
```
Method: POST
URL: http://localhost:8080/api/v1/users
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
  Content-Type: application/json
Body (JSON):
{
  "email": "newuser@acme.com",
  "password": "password123",
  "first_name": "New",
  "last_name": "User",
  "role": "member"
}
```

---

### ✅ STEP 11: List Users
```
Method: GET
URL: http://localhost:8080/api/v1/users
Headers:
  Authorization: Bearer YOUR_ACCESS_TOKEN_HERE
```

---

## 🔑 Test Accounts (from seed data)

| Email | Password | Role |
|-------|----------|------|
| admin@acme.com | password123 | admin |
| manager@acme.com | password123 | manager |
| member@acme.com | password123 | member |

---

## 📝 Tips for Thunder Client Free Version

1. **Save your access token**: Copy it to a notepad after login
2. **Replace placeholders**:
   - `YOUR_ACCESS_TOKEN_HERE` → paste your actual token
   - `TASK_ID_HERE` → paste task ID from create response
   - `ISSUE_ID_HERE` → paste issue ID from create response

3. **Create reusable requests**:
   - Save commonly used requests in Thunder Client
   - Name them clearly (e.g., "Login", "Create Task")

4. **Test different roles**:
   - Login as manager@acme.com to test manager permissions
   - Login as member@acme.com to test member permissions

---

## 🚨 Common Issues

### "Unauthorized" error
- Check if token is included in Authorization header
- Format: `Bearer YOUR_TOKEN` (with space after Bearer)
- Token might be expired (tokens expire after 15 minutes)

### "Invalid token"
- Login again to get a new token
- Make sure you copied the complete token

### "Forbidden" error
- User doesn't have permission
- Try logging in as admin@acme.com

### Can't create users
- Make sure you're logged in as admin or manager
- Members can't create users

---

## 🎯 Quick Workflow

1. ✅ Health check
2. ✅ Login → copy access_token
3. ✅ Create task → copy task id
4. ✅ List tasks
5. ✅ Update task status
6. ✅ Create issue → check AI summary
7. ✅ List issues

**That's it! You're testing a production-ready API!** 🎉
