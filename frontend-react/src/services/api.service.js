import api from '../lib/api';

export const authService = {
  async register(data) {
    const response = await api.post('/auth/register', {
      org_name: data.orgName,
      email: data.email,
      password: data.password,
      first_name: data.firstName,
      last_name: data.lastName,
    });
    return response;
  },

  async login(email, password) {
    const response = await api.post('/auth/login', { email, password });
    return response;
  },

  async logout() {
    const response = await api.post('/auth/logout');
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    return response;
  },

  async getMe() {
    const response = await api.get('/auth/me');
    return response;
  },

  async refreshToken(refreshToken) {
    const response = await api.post('/auth/refresh', { refresh_token: refreshToken });
    return response;
  },
};

export const taskService = {
  async getTasks(params = {}) {
    const response = await api.get('/tasks', { params });
    return response;
  },

  async getAIReport() {
    const response = await api.get('/tasks/ai-report');
    return response;
  },

  async getTask(id) {
    const response = await api.get(`/tasks/${id}`);
    return response;
  },

  async createTask(data) {
    const response = await api.post('/tasks', data);
    return response;
  },

  async updateTask(id, data) {
    const response = await api.patch(`/tasks/${id}`, data);
    return response;
  },

  async deleteTask(id) {
    const response = await api.delete(`/tasks/${id}`);
    return response;
  },

  async getMyTasks() {
    const response = await api.get('/tasks/my');
    return response;
  },

  // Workflow actions
  async markDone(id) {
    const response = await api.post(`/tasks/${id}/done`);
    return response;
  },

  async verifyTask(id) {
    const response = await api.post(`/tasks/${id}/verify`);
    return response;
  },

  async approveTask(id) {
    const response = await api.post(`/tasks/${id}/approve`);
    return response;
  },

  async rejectTask(id) {
    const response = await api.post(`/tasks/${id}/reject`);
    return response;
  },
};

export const issueService = {
  async getIssues(params = {}) {
    const response = await api.get('/issues', { params });
    return response;
  },

  async getIssue(id) {
    const response = await api.get(`/issues/${id}`);
    return response;
  },

  async createIssue(data) {
    const response = await api.post('/issues', data);
    return response;
  },

  async updateIssue(id, data) {
    const response = await api.patch(`/issues/${id}`, data);
    return response;
  },

  async deleteIssue(id) {
    const response = await api.delete(`/issues/${id}`);
    return response;
  },
};

export const userService = {
  async getUsers(params = {}) {
    const response = await api.get('/users', { params });
    return response;
  },

  async getUser(id) {
    const response = await api.get(`/users/${id}`);
    return response;
  },

  async createUser(data) {
    const response = await api.post('/users', data);
    return response;
  },

  async updateUser(id, data) {
    const response = await api.patch(`/users/${id}`, data);
    return response;
  },

  async deleteUser(id) {
    const response = await api.delete(`/users/${id}`);
    return response;
  },
};

export const reportService = {
  async getWeeklySummary() {
    const response = await api.get('/reports/weekly-summary');
    return response;
  },
};

export const auditLogService = {
  async list(limit = 50) {
    const response = await api.get('/audit-logs', { params: { limit } });
    return response;
  },
};
