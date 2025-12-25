# SaaS Task Manager - React Frontend

A professional, production-ready React frontend for the SaaS Task Manager backend.

## 🚀 Features

- ✅ **Modern React 18** with Hooks and Context API
- ✅ **Vite** for lightning-fast development
- ✅ **Tailwind CSS** for beautiful, responsive UI
- ✅ **React Router** for seamless navigation
- ✅ **Axios** with interceptors for API calls
- ✅ **Token Refresh** automatic token renewal
- ✅ **Toast Notifications** user-friendly feedback
- ✅ **Full CRUD** for Tasks, Issues, and Users
- ✅ **Role-Based Access Control** UI adapts to user role
- ✅ **Professional Design** gradient backgrounds, smooth animations
- ✅ **Production Ready** optimized build, error handling

## 📦 Installation

```bash
cd frontend-react
npm install
```

## 🏃 Running the App

### Development Mode
```bash
npm run dev
```
The app will run on http://localhost:3000

### Production Build
```bash
npm run build
npm run preview
```

## 🔧 Configuration

Edit `.env` file to configure the API URL:
```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 📁 Project Structure

```
frontend-react/
├── src/
│   ├── components/         # Reusable components
│   │   ├── Layout.jsx     # Main layout with header & nav
│   │   ├── TaskModal.jsx  # Task create/edit modal
│   │   ├── IssueModal.jsx # Issue create/edit modal
│   │   └── UserModal.jsx  # User create modal
│   ├── pages/             # Page components
│   │   ├── Login.jsx      # Login page
│   │   ├── Register.jsx   # Registration page
│   │   ├── Dashboard.jsx  # Dashboard with stats
│   │   ├── Tasks.jsx      # Tasks management
│   │   ├── Issues.jsx     # Issues management
│   │   └── Users.jsx      # Users management
│   ├── services/          # API services
│   │   └── api.service.js # API calls (auth, tasks, issues, users)
│   ├── context/           # React Context
│   │   └── AuthContext.jsx # Authentication context
│   ├── lib/               # Utilities
│   │   └── api.js         # Axios instance with interceptors
│   ├── App.jsx            # Main app with routing
│   ├── main.jsx           # Entry point
│   └── index.css          # Global styles + Tailwind
├── index.html             # HTML template
├── vite.config.js         # Vite configuration
├── tailwind.config.js     # Tailwind configuration
├── package.json           # Dependencies
└── .env                   # Environment variables
```

## 🎨 Features Breakdown

### Authentication
- Login with email & password
- Register new organization
- Automatic token refresh
- Protected routes
- Persistent sessions

### Dashboard
- Overview of tasks and issues
- Quick stats cards
- Completion rate tracking
- Getting started guide

### Tasks Management
- Create, edit, delete tasks
- Filter by status and priority
- Due date tracking
- Status badges (pending, in progress, completed)
- Priority badges (low, medium, high)

### Issues Management
- Create, edit, delete issues
- Filter by status and severity
- AI-generated summaries (when available)
- Status badges (open, in progress, resolved, closed)
- Severity badges (low, medium, high, critical)

### Users Management
- View all organization users
- Create new users (admin/manager only)
- Delete users (admin only)
- Role-based UI (admin, manager, member)

## 🔐 Security

- JWT tokens stored in localStorage
- Automatic token refresh on 401 errors
- Protected routes with auth guards
- Role-based access control
- Secure API communication

## 🎯 User Roles

### Admin
- Full access to all features
- Can create and delete users
- Can manage all tasks and issues

### Manager
- Can create users
- Can manage tasks and issues
- Cannot delete users

### Member
- Can view and manage own tasks
- Can create issues
- Limited access to user management

## 📱 Responsive Design

- Mobile-first approach
- Tablet and desktop optimized
- Touch-friendly interface
- Collapsible navigation
- Adaptive layouts

## 🚀 Deployment

### Build for Production
```bash
npm run build
```

The `dist/` folder will contain the optimized production build.

### Deploy to Netlify/Vercel
1. Connect your repository
2. Build command: `npm run build`
3. Publish directory: `dist`
4. Set environment variable: `VITE_API_BASE_URL`

### Deploy with Docker
```bash
docker build -t saas-frontend .
docker run -p 3000:3000 saas-frontend
```

## 🛠️ Development Tips

### Adding New API Endpoints
1. Add service method in `src/services/api.service.js`
2. Create/update page component
3. Use the service method with try/catch
4. Show toast notifications for feedback

### Adding New Routes
1. Create page component in `src/pages/`
2. Add route in `src/App.jsx`
3. Add navigation link in `src/components/Layout.jsx`

### Styling Guidelines
- Use Tailwind utility classes
- Custom components in `@layer components` in index.css
- Follow color scheme: primary-500 (purple/blue gradient)
- Use provided button classes: btn-primary, btn-secondary, etc.

## 📚 Tech Stack

- **React 18.2** - UI library
- **Vite 5** - Build tool
- **React Router 6** - Routing
- **Axios 1.6** - HTTP client
- **Tailwind CSS 3.4** - Styling
- **Lucide React** - Icons
- **React Hot Toast** - Notifications
- **date-fns** - Date formatting

## 🐛 Troubleshooting

**Problem: API calls failing**
- Check if backend is running on correct port
- Verify VITE_API_BASE_URL in .env
- Check browser console for CORS errors

**Problem: Login not working**
- Clear localStorage
- Check network tab for API responses
- Verify credentials

**Problem: Build errors**
- Delete node_modules and package-lock.json
- Run `npm install` again
- Check Node.js version (requires 16+)

## 📖 API Documentation

See backend README for complete API documentation.

## 🎉 Getting Started

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Start backend server** (if not already running):
   ```bash
   cd ../
   go run cmd/server/main.go
   ```

3. **Start frontend:**
   ```bash
   npm run dev
   ```

4. **Open browser:**
   Navigate to http://localhost:3000

5. **Register an account:**
   - Click "Create one now"
   - Fill in organization details
   - You'll be logged in automatically

6. **Start using the app:**
   - Create tasks
   - Report issues
   - Invite team members

Enjoy your professional SaaS Task Manager! 🚀
