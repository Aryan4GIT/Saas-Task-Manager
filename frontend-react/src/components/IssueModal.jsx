import { useState, useEffect, useMemo } from 'react';
import { issueService, userService } from '../services/api.service';
import toast from 'react-hot-toast';
import { X } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export default function IssueModal({ issue, onClose }) {
  const { user: currentUser } = useAuth();

  const [formData, setFormData] = useState({
    title: '',
    description: '',
    status: 'open',
    severity: 'medium',
    assigned_to: '',
  });
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState([]);
  const [usersLoading, setUsersLoading] = useState(false);

  const canAssign = useMemo(() => {
    return currentUser?.role === 'admin' || currentUser?.role === 'manager';
  }, [currentUser?.role]);

  useEffect(() => {
    if (issue) {
      setFormData({
        title: issue.title || '',
        description: issue.description || '',
        status: issue.status || 'open',
        severity: issue.severity || 'medium',
        assigned_to: issue.assigned_to || '',
      });
    }
  }, [issue]);

  useEffect(() => {
    const loadUsers = async () => {
      if (!canAssign) return;
      try {
        setUsersLoading(true);
        const response = await userService.getUsers();
        setUsers(response.data || []);
      } catch (error) {
        setUsers([]);
      } finally {
        setUsersLoading(false);
      }
    };

    loadUsers();
  }, [canAssign]);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);

    try {
      if (issue) {
        await issueService.updateIssue(issue.id, {
          ...formData,
          assigned_to: canAssign ? (formData.assigned_to || '') : undefined,
        });
        toast.success('Issue updated successfully');
      } else {
        const payload = {
          ...formData,
          assigned_to: canAssign ? (formData.assigned_to || '') : undefined,
        };
        if (!canAssign || !formData.assigned_to) delete payload.assigned_to;
        await issueService.createIssue(payload);
        toast.success('Issue created successfully');
      }

      onClose(true);
    } catch (error) {
      const message = error.response?.data?.message || 'Operation failed';
      toast.error(message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center p-4 z-50">
      <div className="rounded-2xl w-full max-w-md max-h-[90vh] overflow-y-auto bg-slate-900 border border-slate-800 shadow-2xl">
        <div className="sticky top-0 px-6 py-4 flex justify-between items-center bg-slate-800 border-b border-slate-700">
          <h2 className="text-2xl font-bold text-white">
            {issue ? 'Edit Issue' : 'Create Issue'}
          </h2>
          <button
            onClick={() => onClose(false)}
            className="p-2 hover:bg-navy-700 rounded-lg transition-colors"
          >
            <X className="w-5 h-5 text-gray-400" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6 space-y-5">
          <div>
            <label htmlFor="title" className="block text-sm font-medium text-gray-700 mb-2">
              Title *
            </label>
            <input
              id="title"
              type="text"
              required
              className="input"
              placeholder="Enter issue title"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            />
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
              Description
            </label>
            <textarea
              id="description"
              rows="4"
              className="input"
              placeholder="Describe the issue in detail"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            />
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="status" className="block text-sm font-medium text-gray-700 mb-2">
                Status *
              </label>
              <select
                id="status"
                required
                className="input"
                value={formData.status}
                onChange={(e) => setFormData({ ...formData, status: e.target.value })}
              >
                <option value="open">Open</option>
                <option value="in_progress">In Progress</option>
                <option value="resolved">Resolved</option>
                <option value="closed">Closed</option>
              </select>
            </div>

            <div>
              <label htmlFor="severity" className="block text-sm font-medium text-gray-700 mb-2">
                Severity *
              </label>
              <select
                id="severity"
                required
                className="input"
                value={formData.severity}
                onChange={(e) => setFormData({ ...formData, severity: e.target.value })}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="critical">Critical</option>
              </select>
            </div>
          </div>

          {canAssign && (
            <div>
              <label htmlFor="assigned_to" className="block text-sm font-medium text-gray-700 mb-2">
                Assign To
              </label>
              <select
                id="assigned_to"
                className="input"
                value={formData.assigned_to}
                onChange={(e) => setFormData({ ...formData, assigned_to: e.target.value })}
                disabled={usersLoading}
              >
                <option value="">Unassigned</option>
                {users.map((u) => (
                  <option key={u.id} value={u.id}>
                    {u.first_name} {u.last_name} ({u.role})
                  </option>
                ))}
              </select>
            </div>
          )}

          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Saving...' : issue ? 'Update Issue' : 'Create Issue'}
            </button>
            <button
              type="button"
              onClick={() => onClose(false)}
              className="flex-1 btn btn-secondary"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
