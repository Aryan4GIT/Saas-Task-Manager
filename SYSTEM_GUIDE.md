# 🚀 Enterprise Task Management System - Complete Setup Guide

## 📋 System Overview

This is an industry-grade task management platform with role-based workflows, AI-powered document analysis, and real-time collaboration features.

### ✨ Key Features

#### 1. **Role-Based Workflow System**
- **Members**: Start tasks, mark as done, upload completion documents
- **Managers**: Verify completed work, review AI summaries, reject if needed
- **Admins**: Final approval authority, generate AI reports, manage users

#### 2. **Document Management**
- Upload documents when marking tasks as done
- Automatic AI-generated summaries for faster review
- Support for PDF, DOC, TXT, MD, CSV, XLSX formats
- File size limit: 10MB

#### 3. **AI-Powered Features**
- Document summarization using Google Gemini AI
- Task analysis and risk assessment
- Weekly reports and insights
- RAG (Retrieval-Augmented Generation) for intelligent search

#### 4. **Complete Task Lifecycle**
```
TODO → IN PROGRESS → DONE (+ Document) → VERIFIED → APPROVED
```

---

## 🛠️ Setup Instructions

### 1. Database Migration

First, apply the document fields migration:

```bash
# Connect to PostgreSQL
psql -U postgres -d saas_db

# Run the migration
\i database/migration_004_task_documents.sql

# Verify columns were added
\d tasks
```

You should see these new columns:
- `document_filename` (VARCHAR 255)
- `document_path` (VARCHAR 500)
- `document_summary` (TEXT)

### 2. Backend Configuration

Ensure your `.env` or config file has:

```env
# Gemini AI Configuration
GEMINI_API_KEY=your_gemini_api_key_here
GEMINI_MODEL=gemini-2.5-flash
GEMINI_EMBEDDING_MODEL=text-embedding-004

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=saas_db
DB_USER=postgres
DB_PASSWORD=your_password

# Server
PORT=8080
JWT_SECRET=your_jwt_secret_here
```

### 3. Start Backend

```bash
cd cmd/server
go run main.go
```

Backend will start on http://localhost:8080

### 4. Start Frontend

```bash
cd frontend-react
npm install
npm run dev
```

Frontend will start on http://localhost:5173

---

## 🔄 Complete Workflow Guide

### For Members:

1. **View Assigned Tasks**
   - See only tasks assigned to you
   - Tasks marked with "My Task" badge

2. **Start Working**
   - Click "Start Work" button
   - Status changes to "In Progress"

3. **Mark as Done**
   - Click "Mark Done" button
   - Upload document modal appears
   - Select file (optional but recommended)
   - Click "Upload & Mark Done"
   - AI generates summary automatically

### For Managers:

1. **View Done Tasks**
   - Use "Review Queue (Done)" quick filter
   - See all tasks waiting for verification

2. **Review Work**
   - Check task description
   - Read AI-generated document summary
   - Review document if needed

3. **Verify or Reject**
   - Click "Verify" to approve work
   - Click "Reject" to send back (provide reason)

### For Admins:

1. **Final Approval**
   - Use "Approvals (Verified)" quick filter
   - Review verified tasks

2. **Approve or Reject**
   - Click "Approve" for final approval
   - Click "Reject" to send back for rework

3. **Generate AI Reports**
   - Click "Generate AI Report"
   - Get comprehensive analysis of all tasks
   - Identifies risks, bottlenecks, and recommendations

---

## 🎨 UI/UX Features

### Visual Indicators

- **Color-Coded Status Badges**
  - Yellow: Todo
  - Blue: In Progress
  - Green: Done
  - Purple: Verified
  - Emerald: Approved
  - Red: Blocked

- **Priority Levels**
  - Gray: Low
  - Orange: Medium
  - Red: High
  - Dark Red: Urgent

- **My Task Indicator**
  - Blue left border
  - "My Task" badge
  - Highlighted in list

### Document Display

- File icon with filename
- AI summary in expandable card
- Clear visual hierarchy
- Easy to scan information

### Workflow Actions

- Gradient buttons for actions
- Icon + label for clarity
- Hover effects for feedback
- Disabled states for safety

---

## 🔒 Permission System

### Member Permissions
- ✅ View own assigned tasks
- ✅ Start work on assigned tasks
- ✅ Mark tasks as done
- ✅ Upload documents
- ❌ Cannot edit/delete tasks
- ❌ Cannot assign tasks
- ❌ Cannot verify/approve

### Manager Permissions
- ✅ View all tasks
- ✅ Create new tasks
- ✅ Edit tasks
- ✅ Delete tasks
- ✅ Assign tasks to members
- ✅ Verify completed work
- ✅ Reject work
- ✅ Generate weekly summaries
- ❌ Cannot give final approval

### Admin Permissions
- ✅ All manager permissions
- ✅ Final approval authority
- ✅ Manage users
- ✅ Generate AI reports
- ✅ View audit logs
- ✅ System configuration

---

## 📊 Database Schema

### Tasks Table (Updated)

```sql
CREATE TABLE tasks (
    id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'todo',
    priority VARCHAR(50) DEFAULT 'medium',
    assigned_to UUID REFERENCES users(id),
    created_by UUID NOT NULL,
    verified_by UUID,
    verified_at TIMESTAMP,
    approved_by UUID,
    approved_at TIMESTAMP,
    document_filename VARCHAR(255),      -- NEW
    document_path VARCHAR(500),          -- NEW
    document_summary TEXT,               -- NEW
    due_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## 🔌 API Endpoints

### Task Endpoints

```
POST   /api/v1/tasks              - Create task (manager/admin)
GET    /api/v1/tasks              - List all tasks
GET    /api/v1/tasks/my           - List my assigned tasks
GET    /api/v1/tasks/:id          - Get task details
PATCH  /api/v1/tasks/:id          - Update task (manager/admin)
DELETE /api/v1/tasks/:id          - Delete task (manager/admin)

POST   /api/v1/tasks/:id/done     - Mark done (with optional file upload)
POST   /api/v1/tasks/:id/verify   - Verify task (manager/admin)
POST   /api/v1/tasks/:id/approve  - Approve task (admin only)
POST   /api/v1/tasks/:id/reject   - Reject task (manager/admin)

GET    /api/v1/tasks/ai-report    - Generate AI report (admin only)
```

### Document Upload

When marking a task as done:

```javascript
// With document
const formData = new FormData();
formData.append('document', file);
await taskService.markDone(taskId, formData);

// Without document
await taskService.markDone(taskId);
```

---

## 🤖 AI Features

### Document Summarization

When a document is uploaded:
1. File is saved to `cmd/server/uploads/tasks/`
2. Text content is extracted (for TXT, MD, LOG files)
3. AI generates 3-5 sentence summary covering:
   - Main deliverables
   - Key findings
   - Notable points for review

### Task Analysis Report

Admin AI Report includes:
- Overall status summary
- Key risks and bottlenecks
- Recommended next actions
- Overdue task analysis
- Resource allocation insights

---

## 🚨 Troubleshooting

### Backend Not Starting

1. Check database connection
2. Verify Gemini API key is set
3. Ensure port 8080 is available

### File Upload Not Working

1. Check `cmd/server/uploads/tasks/` directory exists
2. Verify write permissions
3. Check file size (max 10MB)
4. Confirm Content-Type in API call

### AI Summary Not Generating

1. Verify GEMINI_API_KEY is set
2. Check API quota/billing
3. Ensure text extraction worked
4. Check backend logs for errors

### Tasks Not Showing

1. Check user role and permissions
2. Verify tasks are assigned correctly
3. Clear filters
4. Check network tab for API errors

---

## 📈 Performance Tips

1. **Database Indexes**
   - Status and priority are indexed
   - Assigned_to is indexed
   - Document filename has conditional index

2. **Query Optimization**
   - Use status filters for large datasets
   - Limit results per page
   - Use quick filters for common queries

3. **File Uploads**
   - Compress large files before upload
   - Use text formats for better AI analysis
   - Keep documents focused and relevant

---

## 🔐 Security Best Practices

1. **Authentication**
   - JWT tokens with expiration
   - Refresh token rotation
   - Secure password hashing

2. **Authorization**
   - Role-based access control
   - Permission checks on every endpoint
   - Task ownership validation

3. **File Upload**
   - File type validation
   - Size limits enforced
   - Virus scanning recommended (external)
   - Unique filenames to prevent collisions

4. **Data Protection**
   - Input validation on all endpoints
   - SQL injection prevention
   - XSS protection
   - CORS properly configured

---

## 📝 Testing Checklist

### Member Workflow
- [ ] Can view only assigned tasks
- [ ] Can start work on task
- [ ] Can mark task as done
- [ ] Can upload document
- [ ] Can skip document upload
- [ ] Cannot edit/delete tasks
- [ ] Cannot verify/approve

### Manager Workflow
- [ ] Can view all tasks
- [ ] Can create/edit/delete tasks
- [ ] Can assign tasks to members
- [ ] Can verify done tasks
- [ ] Can reject tasks
- [ ] Can see AI summaries
- [ ] Cannot give final approval

### Admin Workflow
- [ ] All manager permissions work
- [ ] Can approve verified tasks
- [ ] Can generate AI reports
- [ ] Can manage users
- [ ] Has system overview

### Document Features
- [ ] File upload works
- [ ] AI summary generates
- [ ] Document info displays
- [ ] File size validation
- [ ] File type validation

---

## 🎯 Best Practices

### For Members
- Always upload completion documents
- Provide clear task descriptions
- Update status regularly
- Add comments when needed

### For Managers
- Review AI summaries first
- Provide feedback when rejecting
- Keep tasks properly assigned
- Use filters effectively

### For Admins
- Generate regular AI reports
- Monitor system health
- Review user permissions
- Backup database regularly

---

## 📞 Support

If you encounter issues:

1. Check backend logs in terminal
2. Check browser console for errors
3. Verify database schema is updated
4. Ensure all environment variables are set
5. Review this guide thoroughly

---

## 🔄 Version History

**v2.0.0** - Current Version
- ✅ Document upload integration
- ✅ AI-powered summaries
- ✅ Improved workflow logic
- ✅ Enhanced UI/UX
- ✅ Better error handling
- ✅ Performance optimizations

---

## 🎉 Success Criteria

Your system is working correctly when:

1. ✅ Members can mark tasks as done with documents
2. ✅ AI summaries appear on done tasks
3. ✅ Managers can verify/reject based on summaries
4. ✅ Admins can approve and generate reports
5. ✅ All role permissions work correctly
6. ✅ No console errors in browser or backend
7. ✅ Database queries are fast (< 1 second)

---

**System Status: Production Ready** ✅

This is now an industry-level task management platform with proper workflows, AI integration, and enterprise-grade features!
