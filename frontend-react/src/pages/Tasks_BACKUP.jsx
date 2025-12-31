import { useEffect, useState } from 'react';
import { taskService } from '../services/api.service';
import toast from 'react-hot-toast';
import { Plus, Edit2, Trash2, Calendar, Filter, Check, CheckCircle, XCircle, Shield, Clock, FileText, Download } from 'lucide-react';
import TaskModal from '../components/TaskModal';
import MarkDoneModal from '../components/MarkDoneModal';
import { useAuth } from '../context/AuthContext';

export default function Tasks() {
  const { user } = useAuth();
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [aiReport, setAIReport] = useState('');
  const [aiReportLoading, setAIReportLoading] = useState(false);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [isMarkDoneModalOpen, setIsMarkDoneModalOpen] = useState(false);
  const [selectedTask, setSelectedTask] = useState(null);
  const [filters, setFilters] = useState({
    status: '',
    priority: '',
  });

  const userRole = user?.role || 'member';
  const currentUserId = user?.id || user?.user_id;

  useEffect(() => {
    loadTasks();
  }, [filters]);

  const loadTasks = async () => {
    try {
      setLoading(true);
      const params = {};
      if (filters.status) params.status = filters.status;
      if (filters.priority) params.priority = filters.priority;

      const response = await taskService.getTasks(params);
      // Handle both array and object with tasks property
      const tasksData = Array.isArray(response.data) ? response.data : (response.data?.tasks || response.data?.data || []);
      setTasks(tasksData);
      console.log('[Tasks] Loaded', tasksData.length, 'tasks');
    } catch (error) {
      console.error('[Tasks] Failed to load:', error);
      toast.error('Failed to load tasks - ' + (error.response?.data?.message || 'Server error'));
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setSelectedTask(null);
    setIsModalOpen(true);
  };

  const handleEdit = (task) => {
    setSelectedTask(task);
    setIsModalOpen(true);
  };

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this task?')) return;

    try {
      await taskService.deleteTask(id);
      toast.success('Task deleted successfully');
      loadTasks();
    } catch (error) {
      toast.error('Failed to delete task');
    }
  };

  // Workflow actions
  const handleStartWork = async (id) => {
    try {
      await taskService.updateTask(id, { status: 'in_progress' });
      toast.success('Task started!');
      loadTasks();
    } catch (error) {
      toast.error('Failed to start task');
    }
  };

  const handleMarkDone = async (task) => {
    setSelectedTask(task);
    setIsMarkDoneModalOpen(true);
  };

  const handleVerify = async (id) => {
    try {
      await taskService.verifyTask(id);
      toast.success('Task verified!');
      loadTasks();
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to verify task');
    }
  };

  const handleApprove = async (id) => {
    try {
      await taskService.approveTask(id);
      toast.success('Task approved!');
      loadTasks();
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to approve task');
    }
  };

  const handleReject = async (id) => {
    if (!confirm('Reject this task? It will be sent back to in progress.')) return;
    try {
      await taskService.rejectTask(id);
      toast.success('Task rejected and sent back');
      loadTasks();
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to reject task');
    }
  };

  const handleGenerateAIReport = async () => {
    try {
      setIsModalOpen(false);
      setAIReportLoading(true);
      const response = await taskService.getAIReport();
      const reportText = response.data?.report || '';
      setAIReport(reportText);
      if (!reportText) {
        toast.error('No report generated');
      }
    } catch (error) {
      const message = error.response?.data?.message || error.response?.data?.error || 'Failed to generate AI report';
      toast.error(message);
    } finally {
      setAIReportLoading(false);
    }
  };

  // Get workflow actions based on role and status
  const getWorkflowActions = (task) => {
    const actions = [];

    const isAssignedToMe = Boolean(
      currentUserId && task?.assigned_to && String(task.assigned_to) === String(currentUserId)
    );
    
    // Member workflow actions (only for tasks assigned to the current user)
    if (userRole === 'member' && isAssignedToMe && task.status === 'todo') {
      actions.push({ label: 'Start', onClick: () => handleStartWork(task.id), className: 'bg-gradient-to-r from-blue-500 to-cyan-500 text-white shadow-md shadow-blue-500/50 hover:shadow-lg hover:shadow-blue-500/60 transform hover:-translate-y-0.5 transition', icon: <Clock className="w-4 h-4" /> });
    }
    if (userRole === 'member' && isAssignedToMe && (task.status === 'todo' || task.status === 'in_progress')) {
      actions.push({ label: 'Mark Done', onClick: () => handleMarkDone(task), className: 'bg-gradient-to-r from-teal-500 to-green-500 text-white shadow-md shadow-teal-500/50 hover:shadow-lg hover:shadow-teal-500/60 transform hover:-translate-y-0.5 transition', icon: <Check className="w-4 h-4" /> });
    }
    if (task.status === 'done' && (userRole === 'manager' || userRole === 'admin')) {
      actions.push({ label: 'Verify', onClick: () => handleVerify(task.id), className: 'bg-gradient-to-r from-purple-500 to-pink-500 text-white shadow-md shadow-purple-500/50 hover:shadow-lg hover:shadow-purple-500/60 transform hover:-translate-y-0.5 transition', icon: <CheckCircle className="w-4 h-4" /> });
      actions.push({ label: 'Reject', onClick: () => handleReject(task.id), className: 'bg-gradient-to-r from-red-500 to-pink-500 text-white shadow-md shadow-red-500/50 hover:shadow-lg hover:shadow-red-500/60 transform hover:-translate-y-0.5 transition', icon: <XCircle className="w-4 h-4" /> });
    }
    if (task.status === 'verified' && userRole === 'admin') {
      actions.push({ label: 'Approve', onClick: () => handleApprove(task.id), className: 'bg-gradient-to-r from-emerald-500 to-teal-500 text-white shadow-md shadow-emerald-500/50 hover:shadow-lg hover:shadow-emerald-500/60 transform hover:-translate-y-0.5 transition', icon: <Shield className="w-4 h-4" /> });
      actions.push({ label: 'Reject', onClick: () => handleReject(task.id), className: 'bg-gradient-to-r from-red-500 to-pink-500 text-white shadow-md shadow-red-500/50 hover:shadow-lg hover:shadow-red-500/60 transform hover:-translate-y-0.5 transition', icon: <XCircle className="w-4 h-4" /> });
    }
    
    return actions;
  };

  const canEdit = () => userRole === 'admin' || userRole === 'manager';
  const canDelete = () => userRole === 'admin' || userRole === 'manager';
  const canCreate = () => userRole === 'admin' || userRole === 'manager';

  const handleModalClose = (shouldRefresh) => {
    setIsModalOpen(false);
    setSelectedTask(null);
    if (shouldRefresh) {
      loadTasks();
    }
  };

  const getStatusColor = (status) => {
    const colors = {
      todo: 'bg-yellow-100 text-yellow-800',
      in_progress: 'bg-blue-100 text-blue-800',
      done: 'bg-green-100 text-green-800',
      verified: 'bg-purple-100 text-purple-800',
      approved: 'bg-emerald-100 text-emerald-800',
      blocked: 'bg-red-100 text-red-800',
    };
    return colors[status] || 'bg-white-100 text-white-800';
  };

  const getPriorityColor = (priority) => {
    const colors = {
      low: 'bg-white-100 text-white-800',
      medium: 'bg-orange-100 text-orange-800',
      high: 'bg-red-100 text-red-800',
      urgent: 'bg-red-600 text-white',
    };
    return colors[priority] || 'bg-white-100 text-white-800';
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'No due date';
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    });
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="page-header">
        <div>
          <h2 className="text-3xl font-bold text-white">Tasks</h2>
          <p className="text-white/80 text-sm mt-1">Role: <span className="font-semibold capitalize">{userRole}</span></p>
        </div>
        {canCreate() && (
          <button onClick={handleCreate} className="btn btn-primary flex items-center gap-2">
            <Plus className="w-5 h-5" />
            New Task
          </button>
        )}
      </div>

      {/* Workflow Legend */}
      <div className="card mb-4">
        <h4 className="font-semibold text-white-700 mb-2">📋 Task Workflow</h4>
        <div className="flex flex-wrap gap-2 items-center text-sm">
          <span className="px-2 py-1 bg-yellow-100 text-yellow-800 rounded">Todo</span>
          <span className="text-gray-400">→</span>
          <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded">In Progress</span>
          <span className="text-gray-400">→</span>
          <span className="px-2 py-1 bg-green-100 text-green-800 rounded">Done (Member)</span>
          <span className="text-gray-400">→</span>
          <span className="px-2 py-1 bg-purple-100 text-purple-800 rounded">Verified (Manager)</span>
          <span className="text-gray-400">→</span>
          <span className="px-2 py-1 bg-emerald-100 text-emerald-800 rounded">Approved (Admin)</span>
        </div>
      </div>

      {/* Admin AI Report */}
      {userRole === 'admin' && (
        <div className="card mb-6">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h4 className="font-semibold text-gray-500">AI Report</h4>
              <p className="text-sm text-gray-500">Summarizes the organization’s current tasks for admin review.</p>
            </div>

            <button
              onClick={handleGenerateAIReport}
              disabled={aiReportLoading}
              className="btn btn-primary"
            >
              {aiReportLoading ? 'Generating…' : 'Generate AI Report'}
            </button>
          </div>

          {aiReport && (
            <div className="mt-4">
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 whitespace-pre-wrap text-sm text-gray-700">
                {aiReport}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Filters */}
      <div className="card mb-6">
        <div className="toolbar">
          <div className="flex items-center gap-2">
            <Filter className="w-5 h-5 text-white-600" />
            <span className="text-sm font-medium text-white-700">Filters:</span>
          </div>

          <select
            value={filters.status}
            onChange={(e) => setFilters({ ...filters, status: e.target.value })}
            className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="">All Status</option>
            <option value="todo">To Do</option>
            <option value="in_progress">In Progress</option>
            <option value="done">Done</option>
            <option value="verified">Verified</option>
            <option value="approved">Approved</option>
            <option value="completed">Completed (Done/Verified/Approved)</option>
            <option value="blocked">Blocked</option>
          </select>

          <select
            value={filters.priority}
            onChange={(e) => setFilters({ ...filters, priority: e.target.value })}
            className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="">All Priority</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="urgent">Urgent</option>
          </select>

          {(filters.status || filters.priority) && (
            <button
              onClick={() => setFilters({ status: '', priority: '' })}
              className="text-sm text-primary-600 hover:text-primary-700 font-medium"
            >
              Clear Filters
            </button>
          )}

          {/* Quick Filters by Role */}
          <div className="flex items-center gap-2 ml-auto">
            <span className="text-xs text-white-500">Quick:</span>
            {(userRole === 'manager' || userRole === 'admin') && (
              <button
                onClick={() => setFilters({ ...filters, status: 'done' })}
                className="px-3 py-1.5 rounded-lg text-xs bg-purple-50 text-purple-700 hover:bg-purple-100"
              >
                Review Queue (Done)
              </button>
            )}
            {userRole === 'admin' && (
              <button
                onClick={() => setFilters({ ...filters, status: 'verified' })}
                className="px-3 py-1.5 rounded-lg text-xs bg-emerald-50 text-emerald-700 hover:bg-emerald-100"
              >
                Approvals (Verified)
              </button>
            )}
            <button
              onClick={() => setFilters({ ...filters, status: 'completed' })}
              className="px-3 py-1.5 rounded-lg text-xs bg-green-50 text-green-700 hover:bg-green-100"
            >
              Completed
            </button>
          </div>
        </div>
      </div>

      {/* Tasks List */}
      {tasks.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-100 mb-4">No tasks found</p>
          {canCreate() && (
            <button onClick={handleCreate} className="btn btn-primary">
              Create your first task
            </button>
          )}
        </div>
      ) : (
        <div className="grid gap-4">
          {tasks.map((task) => {
            const workflowActions = getWorkflowActions(task);
            return (
              <div key={task.id} className="card hover:shadow-xl transition-shadow">
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <h3 className="text-lg font-semibold text-gray-300 mb-2">{task.title}</h3>
                    <p className="text-gray-400 mb-4">{task.description || 'No description'}</p>
                    
                    <div className="flex items-center gap-4 flex-wrap mb-3">
                      <span className={`badge ${getStatusColor(task.status)}`}>
                        {task.status?.replace('_', ' ') || 'Unknown'}
                      </span>
                      <span className={`badge ${getPriorityColor(task.priority)}`}>
                        {task.priority || 'No priority'}
                      </span>
                      <span className="flex items-center gap-1 text-sm text-gray-500">
                        <Calendar className="w-4 h-4" />
                        {formatDate(task.due_date)}
                      </span>
                      {task.assigned_to_name && (
                        <span className="text-sm text-purple-600">👤 {task.assigned_to_name}</span>
                      )}
                    </div>

                    {/* Workflow info */}
                    {(task.verified_by_name || task.approved_by_name) && (
                      <div className="text-xs text-gray-500 mb-2">
                        {task.verified_by_name && <span className="mr-3">✓ Verified by {task.verified_by_name}</span>}
                        {task.approved_by_name && <span>✓ Approved by {task.approved_by_name}</span>}
                      </div>
                    )}

                    {/* Document Info */}
                    {task.document_filename && (
                      <div className="mt-3 p-3 bg-blue-50 border border-blue-200 rounded-lg">
                        <div className="flex items-start gap-3">
                          <FileText className="w-5 h-5 text-blue-600 flex-shrink-0 mt-0.5" />
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-blue-900 flex items-center gap-2">
                              📎 {task.document_filename}
                            </p>
                            {task.document_summary && (
                              <div className="mt-2 p-2 bg-white rounded border border-blue-100">
                                <p className="text-xs font-semibold text-blue-800 mb-1">🤖 AI Summary:</p>
                                <p className="text-xs text-gray-700 whitespace-pre-wrap">{task.document_summary}</p>
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Workflow Actions */}
                    {workflowActions.length > 0 && (
                      <div className="flex gap-2 pt-3 border-t border-gray-100">
                        {workflowActions.map((action, idx) => (
                          <button
                            key={idx}
                            onClick={action.onClick}
                            className={`px-3 py-1.5 rounded-lg text-sm font-medium flex items-center gap-1.5 transition-colors ${action.className}`}
                          >
                            {action.icon}
                            {action.label}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>

                  <div className="flex items-center gap-2 ml-4">
                    {canEdit() && (
                      <button
                        onClick={() => handleEdit(task)}
                        className="p-2 text-gray-600 hover:text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
                      >
                        <Edit2 className="w-5 h-5" />
                      </button>
                    )}
                    {canDelete() && (
                      <button
                        onClick={() => handleDelete(task.id)}
                        className="p-2 text-gray-600 hover:text-red-600 hover:bg-red-50 rounded-lg transition-colors"
                      >
                        <Trash2 className="w-5 h-5" />
                      </button>
                    )}
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {isModalOpen && (
        <TaskModal task={selectedTask} onClose={handleModalClose} />
      )}

      {isMarkDoneModalOpen && selectedTask && (
        <MarkDoneModal
          task={selectedTask}
          onClose={(shouldRefresh) => {
            setIsMarkDoneModalOpen(false);
            setSelectedTask(null);
            if (shouldRefresh) loadTasks();
          }}
        />
      )}
    </div>
  );
}
