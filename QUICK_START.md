# 🚀 SaaS Task Manager - Quick Start Guide

## ✅ Current Status
- **Backend API**: Running on http://localhost:8080 ✓
- **React Frontend**: Running on http://localhost:3001 ✓
- **Database**: Connected to Supabase ✓

## 🔐 Working Credentials

### Account 1: test@test.com
- **Email**: test@test.com
- **Password**: password123
- **Organization**: test company
- **Status**: ✅ VERIFIED WORKING

## 🎯 How to Use the App

### Step 1: Open the App
Open your browser and go to: **http://localhost:3001**

### Step 2: Login
1. Click "Sign in instead" (if on register page)
2. Enter:
   - Email: `test@test.com`
   - Password: `password123`
3. Click "Sign In"

### Step 3: Explore Features

#### Dashboard
- View your task statistics
- See completion rates
- Quick overview of issues

#### Tasks
- Click "Tasks" in the navigation
- Click "+ New Task" to create
- Fill in title, description, priority, status
- Click "Save"

#### Issues  
- Click "Issues" in the navigation
- Click "+ New Issue" to create
- Report bugs with severity levels
- AI summaries (if Gemini API configured)

#### Users
- Click "Users" in the navigation
- Click "+ New User" to invite team members
- Manage user roles (admin, manager, member)

## ⚠️ Common Issues & Solutions

### Issue: "Registration failed"
**Problem**: Organization name already exists
**Solution**: 
- Use a unique organization name
- OR use existing account (login instead)

### Issue: "Invalid email or password"
**Problem**: Wrong credentials
**Solution**: Use working credentials:
- Email: test@test.com
- Password: password123

### Issue: Frontend won't load
**Problem**: React server not running
**Solution**: 
```bash
cd frontend-react
npm run dev
```
Frontend will be on http://localhost:3001

### Issue: API errors
**Problem**: Backend not running
**Solution**:
```bash
cd portfolio
go run cmd/server/main.go
```
Backend will be on http://localhost:8080

## 🧪 API Testing (PowerShell)

### Test Health Check
```powershell
curl http://localhost:8080/health
```

### Test Login
```powershell
$body = @{
    email = "test@test.com"
    password = "password123"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
    -Method POST `
    -ContentType "application/json" `
    -Body $body
```

### Create Task
```powershell
$token = "YOUR_ACCESS_TOKEN_FROM_LOGIN"

$task = @{
    title = "Complete documentation"
    description = "Write API docs"
    status = "pending"
    priority = "high"
    due_date = "2025-12-30T23:59:59Z"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:8080/api/v1/tasks" `
    -Method POST `
    -ContentType "application/json" `
    -Headers @{Authorization="Bearer $token"} `
    -Body $task
```

## 📋 Features Checklist

- ✅ User Registration
- ✅ User Login
- ✅ JWT Authentication
- ✅ Task Management (CRUD)
- ✅ Issue Tracking (CRUD)
- ✅ User Management
- ✅ Role-Based Access Control
- ✅ Multi-tenancy (Organization isolation)
- ✅ Filtering (by status, priority, severity)
- ✅ Beautiful UI with Tailwind CSS
- ✅ Responsive Design
- ✅ Toast Notifications
- ✅ Auto Token Refresh

## 🎉 Everything is Working!

Your full-stack SaaS application is ready to use. Both the backend and frontend are communicating perfectly. Enjoy!

## 📞 Quick Commands

**Start Backend:**
```bash
cd portfolio
go run cmd/server/main.go
```

**Start Frontend:**
```bash
cd frontend-react
npm run dev
```

**Access App:**
- Frontend: http://localhost:3001
- Backend API: http://localhost:8080
- Health Check: http://localhost:8080/health
