import { useState, useEffect, useMemo } from 'react';
import { taskService, userService } from '../services/api.service';
import toast from 'react-hot-toast';
import { X } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export default function TaskModal({ task, onClose }) {
  const { user: currentUser } = useAuth();

  const [formData, setFormData] = useState({
    title: '',
    description: '',
    status: 'todo',
    priority: 'medium',
    assigned_to: '',
    due_date: '',
  });
  const [loading, setLoading] = useState(false);
  const [users, setUsers] = useState([]);
  const [usersLoading, setUsersLoading] = useState(false);

  const canAssign = useMemo(() => {
    return currentUser?.role === 'admin' || currentUser?.role === 'manager';
  }, [currentUser?.role]);

  const isAdmin = useMemo(() => {
    return currentUser?.role === 'admin';
  }, [currentUser?.role]);

  // Filter users based on role - Admin can assign to managers/members, Manager only to members
  const assignableUsers = useMemo(() => {
    if (!users || users.length === 0) return [];
    
    // Admin can assign to anyone (managers and members)
    if (isAdmin) {
      return users.filter(u => u.role === 'manager' || u.role === 'member');
    }
    
    // Manager can only assign to members
    return users.filter(u => u.role === 'member');
  }, [users, isAdmin]);

  useEffect(() => {
    if (task) {
      setFormData({
        title: task.title || '',
        description: task.description || '',
        status: task.status || 'todo',
        priority: task.priority || 'medium',
        assigned_to: task.assigned_to || '',
        due_date: task.due_date ? new Date(task.due_date).toISOString().slice(0, 16) : '',
      });
    }
  }, [task]);

  useEffect(() => {
    const loadUsers = async () => {
      if (!canAssign) return;
      try {
        setUsersLoading(true);
        const response = await userService.getUsers();
        setUsers(response.data || []);
      } catch (error) {
        // Avoid noisy toasts; assignment is optional.
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
      const data = {
        title: formData.title,
        description: formData.description,
        status: formData.status,
        priority: formData.priority,
        assigned_to: canAssign ? (formData.assigned_to || '') : undefined,
        due_date: (() => {
          if (formData.due_date) return new Date(formData.due_date).toISOString();
          // On update, send empty string to clear the due date.
          return task ? '' : null;
        })(),
      };

      // For create, omit empty assigned_to/due_date to avoid sending "".
      if (!task) {
        if (!canAssign) delete data.assigned_to;
        else if (!formData.assigned_to) delete data.assigned_to;
        if (data.due_date === null) delete data.due_date;
      }

      if (task) {
        await taskService.updateTask(task.id, data);
        toast.success('Task updated successfully');
      } else {
        await taskService.createTask(data);
        toast.success('Task created successfully');
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
            {task ? 'Edit Task' : 'Create Task'}
          </h2>
          <button
            onClick={() => onClose(false)}
            className="p-2 hover:bg-slate-700 rounded-lg transition-colors"
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
              placeholder="Enter task title"
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
              placeholder="Enter task description"
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
                <option value="todo">To Do</option>
                <option value="in_progress">In Progress</option>
                <option value="done">Done</option>
                <option value="blocked">Blocked</option>
              </select>
            </div>

            <div>
              <label htmlFor="priority" className="block text-sm font-medium text-gray-700 mb-2">
                Priority *
              </label>
              <select
                id="priority"
                required
                className="input"
                value={formData.priority}
                onChange={(e) => setFormData({ ...formData, priority: e.target.value })}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
                <option value="urgent">Urgent</option>
              </select>
            </div>
          </div>

          {canAssign && (
            <div>
              <label htmlFor="assigned_to" className="block text-sm font-medium text-gray-700 mb-2">
                Assign To
                {currentUser?.role === 'manager' && (
                  <span className="ml-2 text-xs text-indigo-600 font-normal">
                    (You can assign to yourself or team members)
                  </span>
                )}
              </label>
              <select
                id="assigned_to"
                className="input"
                value={formData.assigned_to}
                onChange={(e) => setFormData({ ...formData, assigned_to: e.target.value })}
                disabled={usersLoading}
              >
                <option value="">{isAdmin ? 'Unassigned (Admin will handle)' : 'Unassigned'}</option>
                {isAdmin && assignableUsers.filter(u => u.role === 'manager').length > 0 && <option disabled>── Managers ──</option>}
                {assignableUsers.filter(u => u.role === 'manager').map((u) => (
                  <option key={u.id} value={u.id}>
                    👔 {u.first_name} {u.last_name} (Manager)
                  </option>
                ))}
                {assignableUsers.filter(u => u.role === 'member').length > 0 && <option disabled>── Members ──</option>}
                {assignableUsers.filter(u => u.role === 'member').map((u) => (
                  <option key={u.id} value={u.id}>
                    👤 {u.first_name} {u.last_name} (Member)
                  </option>
                ))}
              </select>
            </div>
          )}

          <div>
            <label htmlFor="due_date" className="block text-sm font-medium text-gray-700 mb-2">
              Due Date
            </label>
            <input
              id="due_date"
              type="datetime-local"
              className="input"
              value={formData.due_date}
              onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
            />
          </div>

          <div className="flex gap-3 pt-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 btn btn-primary disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Saving...' : task ? 'Update Task' : 'Create Task'}
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
