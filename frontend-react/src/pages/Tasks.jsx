import { useEffect, useState } from 'react';
import { taskService } from '../services/api.service';
import toast from 'react-hot-toast';
import {
  Plus,
  Edit2,
  Trash2,
  Calendar,
  Filter,
  Check,
  CheckCircle,
  XCircle,
  Shield,
  Clock,
  FileText,
  Upload,
  X,
  AlertCircle,
  User
} from 'lucide-react';
import TaskModal from '../components/TaskModal';
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
  const [uploadFile, setUploadFile] = useState(null);
  const [uploadingTaskId, setUploadingTaskId] = useState(null);
  const [filters, setFilters] = useState({
    status: '',
    priority: '',
  });

  const userRole = user?.role || 'member';
  const currentUserId = user?.id || user?.user_id;

  useEffect(() => {
    loadTasks();
  }, [filters]);

  // Auto-refresh when AI summaries are being processed
  useEffect(() => {
    const hasProcessingSummaries = tasks.some(
      task => task.document_summary && 
      (task.document_summary.includes('⏳') || task.document_summary.includes('generating'))
    );

    if (hasProcessingSummaries) {
      const intervalId = setInterval(() => {
        console.log('Auto-refreshing for AI summary updates...');
        loadTasks();
      }, 5000); // Check every 5 seconds

      return () => clearInterval(intervalId);
    }
  }, [tasks]);

  const loadTasks = async () => {
    try {
      setLoading(true);
      const params = {};
      if (filters.status) params.status = filters.status;
      if (filters.priority) params.priority = filters.priority;

      const response = await taskService.getTasks(params);
      const tasksData = Array.isArray(response.data)
        ? response.data
        : response.data?.tasks || response.data?.data || [];
      setTasks(tasksData);
    } catch (error) {
      console.error('[Tasks] Failed to load:', error);
      toast.error(
        'Failed to load tasks - ' + (error.response?.data?.message || 'Server error')
      );
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

  const handleReassign = (task) => {
    // Manager can reassign tasks assigned to them to members or keep for themselves
    setSelectedTask(task);
    setIsModalOpen(true);
    toast.info('💡 You can reassign this task to a team member or keep it for yourself', {
      duration: 4000,
    });
  };

  const handleStartWork = async (id) => {
    try {
      await taskService.updateTask(id, { status: 'in_progress' });
      toast.success('Task started!');
      loadTasks();
    } catch (error) {
      toast.error('Failed to start task');
    }
  };

  const handleMarkDone = (task) => {
    setSelectedTask(task);
    setIsMarkDoneModalOpen(true);
    setUploadFile(null);
  };

  const handleMarkDoneWithFile = async () => {
    if (!selectedTask) return;

    try {
      setUploadingTaskId(selectedTask.id);

      if (uploadFile) {
        const formData = new FormData();
        formData.append('document', uploadFile);
        await taskService.markDone(selectedTask.id, formData);
        toast.success('Task marked as done with document!');
      } else {
        if (!confirm('Mark as done without uploading a document?')) {
          setUploadingTaskId(null);
          return;
        }
        await taskService.markDone(selectedTask.id);
        toast.success('Task marked as done!');
      }

      setIsMarkDoneModalOpen(false);
      setSelectedTask(null);
      setUploadFile(null);
      loadTasks();
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to mark task as done');
    } finally {
      setUploadingTaskId(null);
    }
  };

  const handleVerify = async (id) => {
    if (!confirm('Verify this completed task?')) return;

    try {
      await taskService.verifyTask(id);
      toast.success('Task verified!');
      loadTasks();
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to verify task');
    }
  };

  const handleApprove = async (id) => {
    if (!confirm('Approve this verified task?')) return;

    try {
      await taskService.approveTask(id);
      toast.success('Task approved!');
      loadTasks();
    } catch (error) {
      toast.error(error.response?.data?.message || 'Failed to approve task');
    }
  };

  const handleReject = async (id) => {
    const reason = prompt('Why are you rejecting this task?');
    if (!reason) return;

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
      setAIReportLoading(true);
      const response = await taskService.getAIReport();
      const reportText = response.data?.report || '';
      setAIReport(reportText);
      if (!reportText) {
        toast.error('No report generated');
      } else {
        toast.success('AI Report generated!');
      }
    } catch (error) {
      const message =
        error.response?.data?.message ||
        error.response?.data?.error ||
        'Failed to generate AI report';
      toast.error(message);
    } finally {
      setAIReportLoading(false);
    }
  };

  const getWorkflowActions = (task) => {
    const actions = [];

    const isAssignedToMe = Boolean(
      currentUserId &&
        task?.assigned_to &&
        String(task.assigned_to) === String(currentUserId)
    );

    // Managers can reassign tasks assigned to them to members OR start work themselves
    if (userRole === 'manager' && isAssignedToMe) {
      if (task.status === 'todo') {
        actions.push({
          label: 'Start Work',
          onClick: () => handleStartWork(task.id),
          className:
            'bg-gradient-to-r from-blue-500 to-cyan-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <Clock className="w-4 h-4" />,
        });
      }
      
      if (task.status === 'todo' || task.status === 'in_progress') {
        actions.push({
          label: 'Mark Done',
          onClick: () => handleMarkDone(task),
          className:
            'bg-gradient-to-r from-teal-500 to-green-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <Check className="w-4 h-4" />,
        });
        actions.push({
          label: 'Reassign to Member',
          onClick: () => handleReassign(task),
          className:
            'bg-gradient-to-r from-indigo-500 to-purple-600 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <User className="w-4 h-4" />,
        });
      }
    }

    // Member workflow - can only act on their own tasks
    if (userRole === 'member') {
      if (!isAssignedToMe) return actions; // Members can only act on their assigned tasks

      if (task.status === 'todo') {
        actions.push({
          label: 'Start Work',
          onClick: () => handleStartWork(task.id),
          className:
            'bg-gradient-to-r from-blue-500 to-cyan-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <Clock className="w-4 h-4" />,
        });
      }

      if (task.status === 'todo' || task.status === 'in_progress') {
        actions.push({
          label: 'Mark Done',
          onClick: () => handleMarkDone(task),
          className:
            'bg-gradient-to-r from-teal-500 to-green-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <Check className="w-4 h-4" />,
        });
      }
    }

    // Manager workflow - can verify done tasks
    if (userRole === 'manager' || userRole === 'admin') {
      if (task.status === 'done') {
        actions.push({
          label: 'Verify',
          onClick: () => handleVerify(task.id),
          className:
            'bg-gradient-to-r from-purple-500 to-pink-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <CheckCircle className="w-4 h-4" />,
        });
        actions.push({
          label: 'Reject',
          onClick: () => handleReject(task.id),
          className:
            'bg-gradient-to-r from-red-500 to-pink-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <XCircle className="w-4 h-4" />,
        });
      }
    }

    // Admin workflow - can approve verified tasks
    if (userRole === 'admin') {
      if (task.status === 'verified') {
        actions.push({
          label: 'Approve',
          onClick: () => handleApprove(task.id),
          className:
            'bg-gradient-to-r from-emerald-500 to-teal-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <Shield className="w-4 h-4" />,
        });
        actions.push({
          label: 'Reject',
          onClick: () => handleReject(task.id),
          className:
            'bg-gradient-to-r from-red-500 to-pink-500 text-white shadow-md hover:shadow-lg transform hover:-translate-y-0.5 transition',
          icon: <XCircle className="w-4 h-4" />,
        });
      }
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
      todo: 'bg-yellow-100 text-yellow-800 border-yellow-200',
      in_progress: 'bg-blue-100 text-blue-800 border-blue-200',
      done: 'bg-green-100 text-green-800 border-green-200',
      verified: 'bg-purple-100 text-purple-800 border-purple-200',
      approved: 'bg-emerald-100 text-emerald-800 border-emerald-200',
      blocked: 'bg-red-100 text-red-800 border-red-200',
    };
    return colors[status] || 'bg-gray-100 text-gray-800 border-gray-200';
  };

  const getPriorityColor = (priority) => {
    const colors = {
      low: 'bg-gray-100 text-gray-700 border-gray-200',
      medium: 'bg-orange-100 text-orange-800 border-orange-200',
      high: 'bg-red-100 text-red-800 border-red-200',
      urgent: 'bg-red-600 text-white border-red-700',
    };
    return colors[priority] || 'bg-gray-100 text-gray-700 border-gray-200';
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
          <p className="text-white/80 text-sm mt-1">
            Role: <span className="font-semibold capitalize">{userRole}</span> •{' '}
            <span className="text-white/60">{tasks.length} tasks</span>
          </p>
        </div>
        {canCreate() && (
          <button
            onClick={handleCreate}
            className="btn btn-primary flex items-center gap-2"
          >
            <Plus className="w-5 h-5" />
            New Task
          </button>
        )}
      </div>

      {/* Workflow Legend */}
      <div className="card mb-6 bg-gradient-to-r from-slate-900 to-slate-800 border-slate-700">
        <div className="flex items-center gap-2 mb-3">
          <AlertCircle className="w-5 h-5 text-blue-400" />
          <h4 className="font-semibold text-white">📋 Task Workflow</h4>
        </div>
        <div className="flex flex-wrap gap-2 items-center text-sm">
          <span className="px-3 py-1.5 bg-yellow-100 text-yellow-800 rounded-lg font-medium border border-yellow-200">
            Todo
          </span>
          <span className="text-gray-400">→</span>
          <span className="px-3 py-1.5 bg-blue-100 text-blue-800 rounded-lg font-medium border border-blue-200">
            In Progress
          </span>
          <span className="text-gray-400">→</span>
          <span className="px-3 py-1.5 bg-green-100 text-green-800 rounded-lg font-medium border border-green-200">
            Done (Member + Document)
          </span>
          <span className="text-gray-400">→</span>
          <span className="px-3 py-1.5 bg-purple-100 text-purple-800 rounded-lg font-medium border border-purple-200">
            Verified (Manager)
          </span>
          <span className="text-gray-400">→</span>
          <span className="px-3 py-1.5 bg-emerald-100 text-emerald-800 rounded-lg font-medium border border-emerald-200">
            Approved (Admin)
          </span>
        </div>
      </div>

      {/* Admin AI Report */}
      {userRole === 'admin' && (
        <div className="card mb-6 bg-gradient-to-r from-indigo-900/20 to-purple-900/20 border-indigo-700/50">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h4 className="font-semibold text-white flex items-center gap-2">
                <span className="text-2xl">🤖</span> AI Task Report
              </h4>
              <p className="text-sm text-gray-400 mt-1">
                Generate comprehensive analysis of all organizational tasks
              </p>
            </div>

            <button
              onClick={handleGenerateAIReport}
              disabled={aiReportLoading}
              className="btn btn-primary disabled:opacity-50"
            >
              {aiReportLoading ? 'Generating…' : 'Generate AI Report'}
            </button>
          </div>

          {aiReport && (
            <div className="mt-4">
              <div className="bg-slate-800 border border-slate-700 rounded-lg p-4 whitespace-pre-wrap text-sm text-gray-300">
                {aiReport}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Filters */}
      <div className="card mb-6">
        <div className="toolbar flex-wrap">
          <div className="flex items-center gap-2">
            <Filter className="w-5 h-5 text-white-600" />
            <span className="text-sm font-medium text-white-700">Filters:</span>
          </div>

          <select
            value={filters.status}
            onChange={(e) => setFilters({ ...filters, status: e.target.value })}
            className="px-4 py-2 border border-gray-600 bg-slate-800 rounded-lg text-sm text-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
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
            className="px-4 py-2 border border-gray-600 bg-slate-800 rounded-lg text-sm text-gray-300 focus:outline-none focus:ring-2 focus:ring-primary-500"
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
              className="text-sm text-primary-400 hover:text-primary-300 font-medium"
            >
              Clear Filters
            </button>
          )}

          {/* Quick Filters by Role */}
          <div className="flex items-center gap-2 ml-auto">
            <span className="text-xs text-gray-500">Quick:</span>
            {(userRole === 'manager' || userRole === 'admin') && (
              <button
                onClick={() => setFilters({ ...filters, status: 'done' })}
                className="px-3 py-1.5 rounded-lg text-xs bg-purple-100 text-purple-700 hover:bg-purple-200 border border-purple-200"
              >
                Review Queue (Done)
              </button>
            )}
            {userRole === 'admin' && (
              <button
                onClick={() => setFilters({ ...filters, status: 'verified' })}
                className="px-3 py-1.5 rounded-lg text-xs bg-emerald-100 text-emerald-700 hover:bg-emerald-200 border border-emerald-200"
              >
                Approvals (Verified)
              </button>
            )}
            <button
              onClick={() => setFilters({ ...filters, status: 'completed' })}
              className="px-3 py-1.5 rounded-lg text-xs bg-green-100 text-green-700 hover:bg-green-200 border border-green-200"
            >
              Completed
            </button>
          </div>
        </div>
      </div>

      {/* Tasks List */}
      {tasks.length === 0 ? (
        <div className="card text-center py-12 bg-slate-900 border-slate-800">
          <div className="text-6xl mb-4">📋</div>
          <p className="text-gray-400 mb-4 text-lg">No tasks found</p>
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
            const isMyTask =
              currentUserId &&
              task?.assigned_to &&
              String(task.assigned_to) === String(currentUserId);

            return (
              <div
                key={task.id}
                className={`card hover:shadow-2xl transition-all ${
                  isMyTask ? 'border-l-4 border-l-blue-500' : ''
                } ${task.status === 'done' ? 'bg-green-900/10' : ''} ${
                  task.status === 'verified' ? 'bg-purple-900/10' : ''
                } ${task.status === 'approved' ? 'bg-emerald-900/10' : ''}`}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-start gap-3 mb-2">
                      <h3 className="text-lg font-semibold text-gray-200 flex-1">
                        {task.title}
                      </h3>
                      {isMyTask && (
                        <span className="px-2 py-1 bg-blue-500/20 text-blue-300 text-xs rounded-full border border-blue-500/30">
                          My Task
                        </span>
                      )}
                    </div>

                    <p className="text-gray-400 mb-4 text-sm">
                      {task.description || 'No description'}
                    </p>

                    <div className="flex items-center gap-3 flex-wrap mb-3">
                      <span className={`badge border ${getStatusColor(task.status)}`}>
                        {task.status?.replace('_', ' ').toUpperCase() || 'UNKNOWN'}
                      </span>
                      <span className={`badge border ${getPriorityColor(task.priority)}`}>
                        {task.priority?.toUpperCase() || 'NO PRIORITY'}
                      </span>
                      <span className="flex items-center gap-1 text-sm text-gray-500">
                        <Calendar className="w-4 h-4" />
                        {formatDate(task.due_date)}
                      </span>
                      {task.assigned_to_name && (
                        <span className="flex items-center gap-1 text-sm text-purple-400">
                          <User className="w-4 h-4" />
                          {task.assigned_to_name}
                        </span>
                      )}
                    </div>

                    {/* Workflow info */}
                    {(task.verified_by_name || task.approved_by_name) && (
                      <div className="text-xs text-gray-500 mb-3 flex gap-4">
                        {task.verified_by_name && (
                          <span className="flex items-center gap-1">
                            <CheckCircle className="w-3 h-3" />
                            Verified by {task.verified_by_name}
                          </span>
                        )}
                        {task.approved_by_name && (
                          <span className="flex items-center gap-1">
                            <Shield className="w-3 h-3" />
                            Approved by {task.approved_by_name}
                          </span>
                        )}
                      </div>
                    )}

                    {/* Document Info - Enhanced for Manager/Admin Review */}
                    {task.document_filename && (
                      <div className="mt-3 p-4 bg-gradient-to-r from-blue-900/30 to-purple-900/30 border-2 border-blue-500/40 rounded-xl shadow-lg">
                        <div className="flex items-start gap-3">
                          <div className="p-2 bg-blue-500/20 rounded-lg">
                            <FileText className="w-6 h-6 text-blue-400 flex-shrink-0" />
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-base font-bold text-blue-200 flex items-center gap-2 mb-3">
                              📎 {task.document_filename}
                            </p>
                            {task.document_summary && (
                              <div className="mt-2 p-4 bg-slate-800/80 rounded-xl border-2 border-emerald-500/30 shadow-inner">
                                {task.document_summary.includes('⏳') || task.document_summary.includes('generating') ? (
                                  // Processing state
                                  <div className="flex items-center gap-3">
                                    <div className="animate-spin h-5 w-5 border-2 border-blue-400 border-t-transparent rounded-full"></div>
                                    <div>
                                      <p className="text-sm font-medium text-blue-300">
                                        {task.document_summary}
                                      </p>
                                      <p className="text-xs text-gray-400 mt-1">
                                        This usually takes 5-15 seconds. Page will auto-refresh.
                                      </p>
                                    </div>
                                  </div>
                                ) : (
                                  // Completed state
                                  <>
                                    <div className="flex items-center gap-2 mb-3 pb-2 border-b border-emerald-500/20">
                                      <span className="text-2xl">🤖</span>
                                      <p className="text-sm font-bold text-emerald-400 uppercase tracking-wide">
                                        AI Summary for Review
                                      </p>
                                      {(userRole === 'manager' || userRole === 'admin') && (
                                        <span className="ml-auto px-2 py-0.5 bg-emerald-500/20 text-emerald-300 text-xs rounded-full">
                                          Review Ready
                                        </span>
                                      )}
                                    </div>
                                    <p className="text-sm text-gray-200 leading-relaxed whitespace-pre-wrap font-medium">
                                      {task.document_summary}
                                    </p>
                                  </>
                                )}
                              </div>
                            )}
                            {!task.document_summary && (
                              <div className="mt-2 p-3 bg-yellow-900/20 border border-yellow-700/30 rounded-lg">
                                <div className="flex items-center gap-2">
                                  <AlertCircle className="w-4 h-4 text-yellow-500 flex-shrink-0" />
                                  <p className="text-xs text-yellow-200">
                                    AI summary is being generated. This may take a moment. Refresh the page to see the updated summary.
                                  </p>
                                </div>
                              </div>
                            )}
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Workflow Actions */}
                    {workflowActions.length > 0 && (
                      <div className="flex gap-2 pt-4 border-t border-gray-700 mt-4">
                        {workflowActions.map((action, idx) => (
                          <button
                            key={idx}
                            onClick={action.onClick}
                            className={`px-4 py-2 rounded-lg text-sm font-medium flex items-center gap-2 ${action.className}`}
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
                        className="p-2 text-gray-400 hover:text-primary-400 hover:bg-primary-500/10 rounded-lg transition-colors"
                        title="Edit Task"
                      >
                        <Edit2 className="w-5 h-5" />
                      </button>
                    )}
                    {canDelete() && (
                      <button
                        onClick={() => handleDelete(task.id)}
                        className="p-2 text-gray-400 hover:text-red-400 hover:bg-red-500/10 rounded-lg transition-colors"
                        title="Delete Task"
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

      {/* Task Modal */}
      {isModalOpen && (
        <TaskModal task={selectedTask} onClose={handleModalClose} />
      )}

      {/* Mark Done Modal */}
      {isMarkDoneModalOpen && selectedTask && (
        <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
          <div className="rounded-2xl w-full max-w-md bg-slate-900 border border-slate-800 shadow-2xl">
            <div className="px-6 py-4 flex justify-between items-center bg-slate-800 border-b border-slate-700">
              <h2 className="text-2xl font-bold text-white">Mark Task as Done</h2>
              <button
                onClick={() => {
                  setIsMarkDoneModalOpen(false);
                  setSelectedTask(null);
                  setUploadFile(null);
                }}
                className="p-2 hover:bg-slate-700 rounded-lg transition-colors"
              >
                <X className="w-5 h-5 text-gray-400" />
              </button>
            </div>

            <div className="p-6 space-y-5">
              <div className="bg-blue-900/20 border border-blue-700/30 rounded-lg p-4">
                <p className="text-sm text-blue-300">
                  <strong className="text-blue-200">📄 Upload your completed work document</strong>
                  <br />
                  This helps managers and admins review your work faster with AI-generated
                  summaries.
                </p>
              </div>

              <div className="space-y-3">
                <label className="block text-sm font-medium text-gray-300">
                  Document{' '}
                  <span className="text-gray-500">(PDF, DOC, TXT, etc.)</span>
                </label>

                <div className="relative">
                  <input
                    type="file"
                    id="document"
                    onChange={(e) => {
                      const file = e.target.files[0];
                      if (file) {
                        if (file.size > 10 * 1024 * 1024) {
                          toast.error('File size must be less than 10MB');
                          return;
                        }
                        setUploadFile(file);
                      }
                    }}
                    className="hidden"
                    accept=".pdf,.doc,.docx,.txt,.md,.log,.csv,.xlsx,.xls"
                  />
                  <label
                    htmlFor="document"
                    className="flex items-center justify-center gap-3 px-4 py-8 border-2 border-dashed border-gray-600 rounded-lg cursor-pointer hover:border-primary-500 hover:bg-slate-800 transition-colors"
                  >
                    <Upload className="w-8 h-8 text-gray-400" />
                    <div className="text-center">
                      <p className="text-sm font-medium text-gray-300">
                        {uploadFile ? uploadFile.name : 'Click to upload document'}
                      </p>
                      <p className="text-xs text-gray-500 mt-1">
                        {uploadFile
                          ? `${(uploadFile.size / 1024).toFixed(2)} KB`
                          : 'Max 10MB'}
                      </p>
                    </div>
                  </label>
                </div>

                {uploadFile && (
                  <div className="flex items-center gap-2 p-3 bg-green-900/20 border border-green-700/30 rounded-lg">
                    <FileText className="w-5 h-5 text-green-400" />
                    <span className="text-sm text-green-300 flex-1">
                      {uploadFile.name}
                    </span>
                    <button
                      type="button"
                      onClick={() => setUploadFile(null)}
                      className="text-red-400 hover:text-red-300"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>
                )}
              </div>

              <div className="flex gap-3 pt-4">
                <button
                  onClick={handleMarkDoneWithFile}
                  disabled={uploadingTaskId === selectedTask.id}
                  className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {uploadingTaskId === selectedTask.id
                    ? 'Uploading...'
                    : uploadFile
                    ? 'Upload & Mark Done'
                    : 'Mark Done'}
                </button>
                <button
                  onClick={() => {
                    setIsMarkDoneModalOpen(false);
                    setSelectedTask(null);
                    setUploadFile(null);
                  }}
                  disabled={uploadingTaskId === selectedTask.id}
                  className="flex-1 btn btn-secondary"
                >
                  Cancel
                </button>
              </div>

              <p className="text-xs text-gray-500 text-center">
                AI will generate a summary of your document to help managers review faster
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
