import { useEffect, useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { taskService, issueService, reportService } from '../services/api.service';
import { CheckSquare, AlertCircle, Clock, TrendingUp } from 'lucide-react';
import toast from 'react-hot-toast';

export default function Dashboard() {
  const { user } = useAuth();
  const userRole = user?.role || 'member';
  const [stats, setStats] = useState({
    totalTasks: 0,
    myTasks: 0,
    openIssues: 0,
    completedTasks: 0,
  });
  const [loading, setLoading] = useState(true);

  const [weeklySummary, setWeeklySummary] = useState('');
  const [weeklySummaryLoading, setWeeklySummaryLoading] = useState(false);

  useEffect(() => {
    loadDashboardData();
  }, []);

  const loadDashboardData = async () => {
    try {
      const [tasksRes, myTasksRes, issuesRes, completedRes] = await Promise.all([
        taskService.getTasks(),
        taskService.getMyTasks(),
        issueService.getIssues({ status: 'open' }),
        taskService.getTasks({ status: 'completed' }),
      ]);

      // Helper to extract array from response
      const extractArray = (res) => {
        if (Array.isArray(res.data)) return res.data;
        return res.data?.tasks || res.data?.issues || res.data?.data || [];
      };

      setStats({
        totalTasks: extractArray(tasksRes).length,
        myTasks: extractArray(myTasksRes).length,
        openIssues: extractArray(issuesRes).length,
        completedTasks: extractArray(completedRes).length,
      });

      console.log('[Dashboard] Stats loaded');
    } catch (error) {
      console.error('[Dashboard] Failed to load data:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleWeeklySummary = async () => {
    try {
      setWeeklySummaryLoading(true);
      const res = await reportService.getWeeklySummary();
      const text = res.data?.summary || '';
      setWeeklySummary(text);
      if (!text) toast.error('No summary generated');
    } catch (error) {
      const message = error.response?.data?.message || error.response?.data?.error || 'Failed to generate weekly summary';
      toast.error(message);
    } finally {
      setWeeklySummaryLoading(false);
    }
  };

  const statCards = [
    {
      name: 'Total Tasks',
      value: stats.totalTasks,
      icon: CheckSquare,
      gradient: 'from-blue-500 to-cyan-500',
      iconBg: 'bg-blue-500/20',
      iconColor: 'text-blue-400',
    },
    {
      name: 'My Tasks',
      value: stats.myTasks,
      icon: Clock,
      gradient: 'from-purple-500 to-pink-500',
      iconBg: 'bg-purple-500/20',
      iconColor: 'text-purple-400',
    },
    {
      name: 'Open Issues',
      value: stats.openIssues,
      icon: AlertCircle,
      gradient: 'from-orange-500 to-red-500',
      iconBg: 'bg-orange-500/20',
      iconColor: 'text-orange-400',
    },
    {
      name: 'Completed Tasks',
      value: stats.completedTasks,
      icon: TrendingUp,
      gradient: 'from-teal-500 to-green-500',
      iconBg: 'bg-teal-500/20',
      iconColor: 'text-teal-400',
    },
  ];

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="page">
      <div className="mb-8">
        <h2 className="text-3xl font-bold text-white mb-2">
          Welcome back, {user?.first_name}! 👋
        </h2>
        <p className="text-white/80">Here's what's happening with your projects today.</p>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {statCards.map((stat) => {
          const Icon = stat.icon;
          return (
            <div key={stat.name} className="relative overflow-hidden rounded-2xl bg-slate-900 border border-slate-800 p-6 hover:shadow-xl hover:border-slate-700 transition-all">
              <div className={`absolute top-0 right-0 w-32 h-32 bg-gradient-to-br ${stat.gradient} opacity-10 rounded-full -mr-16 -mt-16`}></div>
              <div className="relative">
                <div className="flex items-center justify-between mb-4">
                  <div className={`${stat.iconBg} p-3 rounded-xl`}>
                    <Icon className={`w-6 h-6 ${stat.iconColor}`} />
                  </div>
                </div>
                <p className="text-sm font-medium text-gray-400 mb-1">{stat.name}</p>
                <p className="text-3xl font-bold text-white">{stat.value}</p>
              </div>
            </div>
          );
        })}
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card">
          <h3 className="text-xl font-bold text-gray-900 mb-4">Quick Stats</h3>
          <div className="space-y-3">
            <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
              <span className="text-sm font-medium text-gray-700">Task Completion Rate</span>
              <span className="text-sm font-bold text-primary-600">
                {stats.totalTasks > 0
                  ? Math.round((stats.completedTasks / stats.totalTasks) * 100)
                  : 0}
                %
              </span>
            </div>
            <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
              <span className="text-sm font-medium text-gray-700">Pending Tasks</span>
              <span className="text-sm font-bold text-gray-900">
                {stats.totalTasks - stats.completedTasks}
              </span>
            </div>
            <div className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
              <span className="text-sm font-medium text-gray-700">Your Role</span>
              <span className="text-sm font-bold text-gray-900 capitalize">{user?.role}</span>
            </div>
          </div>
        </div>

        {(userRole === 'admin' || userRole === 'manager') ? (
          <div className="card">
            <div className="flex items-center justify-between gap-4 mb-4">
              <div>
                <h3 className="text-xl font-bold text-gray-900">Weekly Summary</h3>
                <p className="text-xs text-gray-500">Delayed tasks, high risk issues, bottlenecks</p>
              </div>
              <button
                onClick={handleWeeklySummary}
                disabled={weeklySummaryLoading}
                className="btn btn-primary"
              >
                {weeklySummaryLoading ? 'Generating…' : 'Weekly Summary'}
              </button>
            </div>

            {weeklySummary ? (
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-4 whitespace-pre-wrap text-sm text-gray-700">
                {weeklySummary}
              </div>
            ) : (
              <p className="text-sm text-gray-500">Generate a summary to see insights.</p>
            )}
          </div>
        ) : (
          <div className="card">
            <h3 className="text-xl font-bold text-gray-900 mb-4">Getting Started</h3>
            <ul className="space-y-3">
              <li className="flex items-start gap-3">
                <div className="w-6 h-6 bg-primary-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                  <span className="text-xs font-bold text-primary-600">1</span>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-900">Start working on your tasks</p>
                  <p className="text-xs text-gray-500">Use Tasks to update your status</p>
                </div>
              </li>
              <li className="flex items-start gap-3">
                <div className="w-6 h-6 bg-primary-100 rounded-full flex items-center justify-center flex-shrink-0 mt-0.5">
                  <span className="text-xs font-bold text-primary-600">2</span>
                </div>
                <div>
                  <p className="text-sm font-medium text-gray-900">Report an issue</p>
                  <p className="text-xs text-gray-500">Track bugs and blockers in Issues</p>
                </div>
              </li>
            </ul>
          </div>
        )}
      </div>
    </div>
  );
}
