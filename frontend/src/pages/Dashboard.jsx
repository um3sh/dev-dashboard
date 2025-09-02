import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { 
  Database, 
  Package, 
  Server, 
  Activity,
  GitBranch,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle
} from 'lucide-react';

const Dashboard = () => {
  const [stats, setStats] = useState({
    repositories: 0,
    microservices: 0,
    kubernetesResources: 0,
    recentActions: []
  });

  // Mock data for demonstration
  useEffect(() => {
    setStats({
      repositories: 3,
      microservices: 45,
      kubernetesResources: 12,
      recentActions: [
        {
          id: 1,
          type: 'build',
          service: 'user-service',
          status: 'success',
          commit: 'a1b2c3d',
          branch: 'main',
          timestamp: '2 minutes ago'
        },
        {
          id: 2,
          type: 'deployment',
          service: 'payment-service',
          status: 'running',
          commit: 'e4f5g6h',
          branch: 'main',
          timestamp: '5 minutes ago'
        },
        {
          id: 3,
          type: 'build',
          service: 'notification-service',
          status: 'failure',
          commit: 'i7j8k9l',
          branch: 'feature/alerts',
          timestamp: '10 minutes ago'
        }
      ]
    });
  }, []);

  const getStatusIcon = (status) => {
    switch (status) {
      case 'success':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'failure':
        return <XCircle className="h-5 w-5 text-red-500" />;
      case 'running':
        return <Activity className="h-5 w-5 text-blue-500 animate-pulse" />;
      default:
        return <AlertCircle className="h-5 w-5 text-yellow-500" />;
    }
  };

  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-2 text-gray-600">
          Overview of your monorepo services and Kubernetes resources
        </p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 gap-6 sm:grid-cols-2 lg:grid-cols-4 mb-8">
        <div className="card">
          <div className="flex items-center">
            <div className="p-3 rounded-lg bg-blue-100">
              <Database className="h-6 w-6 text-blue-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">Repositories</p>
              <p className="text-2xl font-bold text-gray-900">{stats.repositories}</p>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center">
            <div className="p-3 rounded-lg bg-green-100">
              <Package className="h-6 w-6 text-green-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">Microservices</p>
              <p className="text-2xl font-bold text-gray-900">{stats.microservices}</p>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center">
            <div className="p-3 rounded-lg bg-purple-100">
              <Server className="h-6 w-6 text-purple-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">K8s Resources</p>
              <p className="text-2xl font-bold text-gray-900">{stats.kubernetesResources}</p>
            </div>
          </div>
        </div>

        <div className="card">
          <div className="flex items-center">
            <div className="p-3 rounded-lg bg-orange-100">
              <Activity className="h-6 w-6 text-orange-600" />
            </div>
            <div className="ml-4">
              <p className="text-sm font-medium text-gray-500">Recent Actions</p>
              <p className="text-2xl font-bold text-gray-900">{stats.recentActions.length}</p>
            </div>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Recent Actions</h2>
          <div className="space-y-4">
            {stats.recentActions.map((action) => (
              <div key={action.id} className="flex items-center space-x-4 p-3 bg-gray-50 rounded-lg">
                {getStatusIcon(action.status)}
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900">
                    {action.type} • {action.service}
                  </p>
                  <div className="flex items-center space-x-2 text-sm text-gray-500">
                    <GitBranch className="h-4 w-4" />
                    <span>{action.branch}</span>
                    <span>•</span>
                    <span>{action.commit}</span>
                  </div>
                </div>
                <div className="flex items-center text-sm text-gray-500">
                  <Clock className="h-4 w-4 mr-1" />
                  {action.timestamp}
                </div>
              </div>
            ))}
          </div>
          <div className="mt-4">
            <Link 
              to="/repositories" 
              className="text-blue-600 hover:text-blue-500 text-sm font-medium"
            >
              View all actions →
            </Link>
          </div>
        </div>

        {/* Quick Actions */}
        <div className="card">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
          <div className="space-y-3">
            <Link
              to="/repositories"
              className="block w-full text-left p-4 bg-gray-50 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <div className="flex items-center">
                <Database className="h-5 w-5 text-gray-400 mr-3" />
                <div>
                  <p className="text-sm font-medium text-gray-900">Manage Repositories</p>
                  <p className="text-sm text-gray-500">Add, configure, and sync repositories</p>
                </div>
              </div>
            </Link>

            <Link
              to="/microservices"
              className="block w-full text-left p-4 bg-gray-50 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <div className="flex items-center">
                <Package className="h-5 w-5 text-gray-400 mr-3" />
                <div>
                  <p className="text-sm font-medium text-gray-900">View Microservices</p>
                  <p className="text-sm text-gray-500">Monitor builds and deployments</p>
                </div>
              </div>
            </Link>

            <Link
              to="/kubernetes"
              className="block w-full text-left p-4 bg-gray-50 hover:bg-gray-100 rounded-lg transition-colors"
            >
              <div className="flex items-center">
                <Server className="h-5 w-5 text-gray-400 mr-3" />
                <div>
                  <p className="text-sm font-medium text-gray-900">Kubernetes Resources</p>
                  <p className="text-sm text-gray-500">Manage K8s overlays and deployments</p>
                </div>
              </div>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;