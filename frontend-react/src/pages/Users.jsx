import { useEffect, useState } from 'react';
import { userService } from '../services/api.service';
import toast from 'react-hot-toast';
import { Plus, Trash2, Mail, Shield } from 'lucide-react';
import UserModal from '../components/UserModal';
import { useAuth } from '../context/AuthContext';

export default function Users() {
  const { user: currentUser } = useAuth();
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [isModalOpen, setIsModalOpen] = useState(false);

  useEffect(() => {
    loadUsers();
  }, []);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const response = await userService.getUsers();
      // Handle both array and object with users property
      const usersData = Array.isArray(response.data) ? response.data : (response.data?.users || response.data?.data || []);
      setUsers(usersData);
      console.log('[Users] Loaded', usersData.length, 'users');
    } catch (error) {
      console.error('[Users] Failed to load:', error);
      if (error.response?.status === 403) {
        toast.error('You do not have permission to view users');
      } else {
        toast.error('Failed to load users - ' + (error.response?.data?.message || 'Server error'));
      }
    } finally {
      setLoading(false);
    }
  };

  const handleCreate = () => {
    setIsModalOpen(true);
  };

  const handleDelete = async (id) => {
    if (!confirm('Are you sure you want to delete this user?')) return;

    try {
      await userService.deleteUser(id);
      toast.success('User deleted successfully');
      loadUsers();
    } catch (error) {
      toast.error('Failed to delete user');
    }
  };

  const handleModalClose = (shouldRefresh) => {
    setIsModalOpen(false);
    if (shouldRefresh) {
      loadUsers();
    }
  };

  const getRoleColor = (role) => {
    const colors = {
      admin: 'from-red-500 to-pink-500',
      manager: 'from-teal-500 to-green-500',
      member: 'from-blue-500 to-cyan-500',
    };
    return colors[role] || 'from-gray-500 to-gray-600';
  };

  const canDelete = (user) => {
    return currentUser?.role === 'admin' && user.id !== currentUser.id;
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
        <h2 className="text-3xl font-bold text-white">Users</h2>
        {(currentUser?.role === 'admin' || currentUser?.role === 'manager') && (
          <button onClick={handleCreate} className="btn btn-primary flex items-center gap-2">
            <Plus className="w-5 h-5" />
            New User
          </button>
        )}
      </div>

      {/* Users List */}
      {users.length === 0 ? (
        <div className="card text-center py-12">
          <p className="text-gray-500 mb-4">No users found</p>
          {(currentUser?.role === 'admin' || currentUser?.role === 'manager') && (
            <button onClick={handleCreate} className="btn btn-primary">
              Add Your First User
            </button>
          )}
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {users.map((user) => (
            <div key={user.id} className="card hover:shadow-xl transition-all">
              <div className="flex justify-between items-start mb-4">
                <div className="flex items-center gap-3">
                  <div className={`w-12 h-12 bg-gradient-to-br ${getRoleColor(user.role)} rounded-xl flex items-center justify-center shadow-lg`}>
                    <span className="text-lg font-bold text-white">
                      {user.first_name?.[0]}{user.last_name?.[0]}
                    </span>
                  </div>
                  <div>
                    <h3 className="font-semibold text-white">
                      {user.first_name} {user.last_name}
                    </h3>
                    <div className="flex items-center gap-1 mt-1">
                      <Shield className="w-3 h-3 text-gray-400" />
                      <span className="text-xs text-gray-400 capitalize">{user.role}</span>
                    </div>
                  </div>
                </div>

                {canDelete(user) && (
                  <button
                    onClick={() => handleDelete(user.id)}
                    className="p-2 hover:bg-red-50 rounded-lg transition-colors"
                    title="Delete"
                  >
                    <Trash2 className="w-5 h-5 text-red-600" />
                  </button>
                )}
              </div>

              <div className="flex items-center gap-2 text-sm text-gray-400">
                <Mail className="w-4 h-4" />
                <span>{user.email}</span>
              </div>

              {user.id === currentUser?.id && (
                <div className="mt-3 px-3 py-1.5 bg-primary-500/20 border border-primary-500/30 rounded-lg text-xs text-primary-400 font-medium">
                  This is you
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {isModalOpen && <UserModal onClose={handleModalClose} />}
    </div>
  );
}
