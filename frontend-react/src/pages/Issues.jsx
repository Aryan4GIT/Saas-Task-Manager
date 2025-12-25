import { useEffect, useState } from 'react';
import { issueService } from '../services/api.service';
import toast from 'react-hot-toast';
import { Plus, Edit2, Trash2, Filter, Sparkles } from 'lucide-react';
import IssueModal from '../components/IssueModal';
import { useAuth } from '../context/AuthContext';

export default function Issues() {
  const { user } = useAuth();
  const userRole = user?.role || 'member';
  const [issues, setIssues] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedIssue, setSelectedIssue] = useState(null);
  const [filters, setFilters] = useState({
    status: '',
    severity: '',
  });

  useEffect(() => {
    loadIssues();
  }, [filters]);

  const loadIssues = async () => {
    try {
      setLoading(true);
      const params = {};
      if (filters.status) params.status = filters.status;
      if (filters.severity) params.severity = filters.severity;

      const response = await issueService.getIssues(params);
      // Handle both array and object with issues property
      const issuesData = Array.isArray(response.data) ? response.data : (response.data?.issues || response.data?.data || []);
      setIssues(issuesData);
      console.log('[Issues] Loaded', issuesData.length, 'issues');
    } catch (error) {
      console.error('[Issues] Failed to load:', error);
      toast.error('Failed to load issues - ' + (error.response?.data?.message || 'Server error'));
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setSelectedIssue(null);
    setIsModalOpen(true);
  };

  const handleEdit = (issue) => {
    setSelectedIssue(issue);
    setIsModalOpen(true);
  };

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this issue?')) return;

    try {
      await issueService.deleteIssue(id);
      toast.success('Issue deleted successfully');
      loadIssues();
    } catch (error) {
      toast.error('Failed to delete issue');
    }
  };

  const handleModalClose = (shouldRefresh) => {
    setIsModalOpen(false);
    setSelectedIssue(null);
    if (shouldRefresh) {
      loadIssues();
    }
  };

  const getStatusColor = (status) => {
    const colors = {
      open: 'bg-blue-100 text-blue-800',
      in_progress: 'bg-yellow-100 text-yellow-800',
      resolved: 'bg-green-100 text-green-800',
      closed: 'bg-gray-100 text-gray-300',
    };
    return colors[status] || 'bg-gray-100 text-gray-300';
  };

  const getSeverityColor = (severity) => {
    const colors = {
      low: 'bg-gray-100 text-gray-800',
      medium: 'bg-orange-100 text-orange-800',
      high: 'bg-red-100 text-red-800',
      critical: 'bg-red-600 text-white',
    };
    return colors[severity] || 'bg-gray-100 text-gray-300';
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
        <h2 className="text-3xl font-bold text-white">Issues</h2>
        <button onClick={handleCreate} className="btn btn-primary flex items-center gap-2">
          <Plus className="w-5 h-5" />
          New Issue
        </button>
      </div>

      {/* Filters */}
      <div className="card mb-6">
        <div className="flex items-center gap-4 flex-wrap">
          <div className="flex items-center gap-2">
            <Filter className="w-5 h-5 text-gray-300" />
            <span className="text-sm font-medium text-black-300">Filters:</span>
          </div>

          <select
            value={filters.status}
            onChange={(e) => setFilters({ ...filters, status: e.target.value })}
            className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-black focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="">All Status</option>
            <option value="open">Open</option>
            <option value="in_progress">In Progress</option>
            <option value="resolved">Resolved</option>
            <option value="closed">Closed</option>
          </select>

          <select
            value={filters.severity}
            onChange={(e) => setFilters({ ...filters, severity: e.target.value })}
            className="px-4 py-2 border border-gray-300 rounded-lg text-sm text-black focus:outline-none focus:ring-2 focus:ring-primary-500"
          >
            <option value="">All Severity</option>
            <option value="low">Low</option>
            <option value="medium">Medium</option>
            <option value="high">High</option>
            <option value="critical">Critical</option>
          </select>

          {(filters.status || filters.severity) && (
            <button
              onClick={() => setFilters({ status: '', severity: '' })}
              className="text-sm text-primary-600 hover:text-primary-700 font-medium"
            >
              Clear Filters
            </button>
          )}
        </div>
      </div>

      {/* Issues List */}
      {issues.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-500 mb-4">No issues found</p>
          <button onClick={handleCreate} className="btn btn-primary">
            Create Your First Issue
          </button>
        </div>
      ) : (
        <div className="grid gap-4">
          {issues.map((issue) => (
            <div key={issue.id} className="card hover:shadow-xl transition-shadow">
              <div className="flex justify-between items-start mb-3">
                <div className="flex-1">
                  <h3 className="text-xl font-semibold text-gray-300 mb-2">{issue.title}</h3>
                  <p className="text-gray-300 text-sm mb-3">{issue.description || 'No description'}</p>
                  
                  {issue.ai_summary && (
                    <div className="bg-gradient-to-r from-purple-50 to-blue-50 border border-purple-200 rounded-lg p-3 mb-3">
                      <div className="flex items-center gap-2 mb-1">
                        <Sparkles className="w-4 h-4 text-purple-600" />
                        <span className="text-xs font-semibold text-purple-600">AI Summary</span>
                      </div>
                      <p className="text-sm text-gray-700">{issue.ai_summary}</p>
                    </div>
                  )}
                </div>
                
                <div className="flex gap-2 ml-4">
                  <button
                    onClick={() => handleEdit(issue)}
                    className="p-2 hover:bg-green-50 rounded-lg transition-colors"
                    title="Edit"
                  >
                    <Edit2 className="w-5 h-5 text-green-600" />
                  </button>
                  {(userRole === 'admin' || userRole === 'manager') && (
                    <button
                      onClick={() => handleDelete(issue.id)}
                      className="p-2 hover:bg-red-50 rounded-lg transition-colors"
                      title="Delete"
                    >
                      <Trash2 className="w-5 h-5 text-red-600" />
                    </button>
                  )}
                </div>
              </div>

              <div className="flex flex-wrap gap-2 items-center">
                <span className={`badge ${getStatusColor(issue.status)}`}>
                  {issue.status.replace('_', ' ')}
                </span>
                <span className={`badge ${getSeverityColor(issue.severity)}`}>
                  {issue.severity}
                </span>
                {issue.reported_by_name && (
                  <span className="badge bg-purple-100 text-purple-700">
                    👤 {issue.reported_by_name}
                  </span>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {isModalOpen && (
        <IssueModal issue={selectedIssue} onClose={handleModalClose} />
      )}
    </div>
  );
}
