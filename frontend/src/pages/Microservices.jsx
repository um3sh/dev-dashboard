import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  Package, 
  Play,
  Clock,
  GitBranch,
  CheckCircle,
  XCircle,
  AlertCircle,
  Activity,
  ExternalLink,
  Filter
} from 'lucide-react';

const Microservices = () => {
  const { repoId } = useParams();
  const [services, setServices] = useState([]);
  const [repository, setRepository] = useState(null);
  const [filter, setFilter] = useState('all');
  const [selectedService, setSelectedService] = useState(null);

  // Mock data for demonstration
  useEffect(() => {
    setRepository({
      id: 1,
      name: 'main-monorepo',
      type: 'monorepo'
    });

    setServices([
      {
        id: 1,
        name: 'user-service',
        path: 'services/user-service',
        description: 'User authentication and management service',
        lastBuild: {
          status: 'success',
          commit: 'a1b2c3d',
          branch: 'main',
          timestamp: '2024-01-20T14:30:00Z',
          buildHash: 'build-123'
        },
        lastDeployment: {
          status: 'success',
          commit: 'a1b2c3d',
          branch: 'main',
          timestamp: '2024-01-20T14:35:00Z',
          buildHash: 'build-123'
        },
        recentActions: [
          { type: 'build', status: 'success', timestamp: '2024-01-20T14:30:00Z' },
          { type: 'deployment', status: 'success', timestamp: '2024-01-20T14:35:00Z' }
        ]
      },
      {
        id: 2,
        name: 'payment-service',
        path: 'services/payment-service',
        description: 'Payment processing and billing service',
        lastBuild: {
          status: 'running',
          commit: 'e4f5g6h',
          branch: 'main',
          timestamp: '2024-01-20T15:00:00Z',
          buildHash: 'build-124'
        },
        lastDeployment: {
          status: 'success',
          commit: 'x9y8z7w',
          branch: 'main',
          timestamp: '2024-01-20T10:15:00Z',
          buildHash: 'build-122'
        },
        recentActions: [
          { type: 'build', status: 'running', timestamp: '2024-01-20T15:00:00Z' },
          { type: 'deployment', status: 'success', timestamp: '2024-01-20T10:15:00Z' }
        ]
      },
      {
        id: 3,
        name: 'notification-service',
        path: 'services/notification-service',
        description: 'Email and SMS notification service',
        lastBuild: {
          status: 'failure',
          commit: 'i7j8k9l',
          branch: 'feature/alerts',
          timestamp: '2024-01-20T13:45:00Z',
          buildHash: null
        },
        lastDeployment: {
          status: 'success',
          commit: 'm3n4o5p',
          branch: 'main',
          timestamp: '2024-01-19T16:20:00Z',
          buildHash: 'build-121'
        },
        recentActions: [
          { type: 'build', status: 'failure', timestamp: '2024-01-20T13:45:00Z' },
          { type: 'deployment', status: 'success', timestamp: '2024-01-19T16:20:00Z' }
        ]
      }
    ]);
  }, [repoId]);

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

  const getStatusClass = (status) => {
    switch (status) {
      case 'success':
        return 'status-success';
      case 'failure':
        return 'status-failure';
      case 'running':
        return 'status-running';
      default:
        return 'status-pending';
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const filteredServices = services.filter(service => {
    if (filter === 'all') return true;
    if (filter === 'success') return service.lastBuild.status === 'success';
    if (filter === 'failure') return service.lastBuild.status === 'failure';
    if (filter === 'running') return service.lastBuild.status === 'running';
    return true;
  });

  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Microservices</h1>
        <p className="mt-2 text-gray-600">
          {repository?.name ? `Services in ${repository.name}` : 'Monitor builds and deployments for all microservices'}
        </p>
      </div>

      {/* Filters */}
      <div className="flex items-center space-x-4 mb-6">
        <div className="flex items-center">
          <Filter className="h-5 w-5 text-gray-400 mr-2" />
          <span className="text-sm font-medium text-gray-700">Filter by status:</span>
        </div>
        <div className="flex space-x-2">
          {[
            { key: 'all', label: 'All' },
            { key: 'success', label: 'Success' },
            { key: 'failure', label: 'Failed' },
            { key: 'running', label: 'Running' }
          ].map(({ key, label }) => (
            <button
              key={key}
              onClick={() => setFilter(key)}
              className={`px-3 py-1 rounded-full text-sm font-medium transition-colors ${
                filter === key
                  ? 'bg-blue-100 text-blue-800'
                  : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
              }`}
            >
              {label}
            </button>
          ))}
        </div>
      </div>

      {/* Services Grid */}
      <div className="grid gap-6">
        {filteredServices.map((service) => (
          <div key={service.id} className="card">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-start space-x-4">
                <div className="p-3 bg-blue-100 rounded-lg">
                  <Package className="h-6 w-6 text-blue-600" />
                </div>
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">{service.name}</h3>
                  <p className="text-gray-600">{service.description}</p>
                  <div className="flex items-center mt-1 text-sm text-gray-500">
                    <ExternalLink className="h-4 w-4 mr-1" />
                    <span>{service.path}</span>
                  </div>
                </div>
              </div>
              <button
                onClick={() => setSelectedService(selectedService === service.id ? null : service.id)}
                className="btn-secondary"
              >
                View Details
              </button>
            </div>

            {/* Build and Deployment Status */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Last Build */}
              <div className="bg-gray-50 p-4 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="text-sm font-medium text-gray-700">Last Build</h4>
                  <span className={getStatusClass(service.lastBuild.status)}>
                    {service.lastBuild.status}
                  </span>
                </div>
                <div className="space-y-1 text-sm text-gray-600">
                  <div className="flex items-center">
                    <GitBranch className="h-4 w-4 mr-2" />
                    <span>{service.lastBuild.branch} • {service.lastBuild.commit}</span>
                  </div>
                  <div className="flex items-center">
                    <Clock className="h-4 w-4 mr-2" />
                    <span>{formatDate(service.lastBuild.timestamp)}</span>
                  </div>
                  {service.lastBuild.buildHash && (
                    <div className="text-xs text-gray-500">
                      Build: {service.lastBuild.buildHash}
                    </div>
                  )}
                </div>
              </div>

              {/* Last Deployment */}
              <div className="bg-gray-50 p-4 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="text-sm font-medium text-gray-700">Last Deployment</h4>
                  <span className={getStatusClass(service.lastDeployment.status)}>
                    {service.lastDeployment.status}
                  </span>
                </div>
                <div className="space-y-1 text-sm text-gray-600">
                  <div className="flex items-center">
                    <GitBranch className="h-4 w-4 mr-2" />
                    <span>{service.lastDeployment.branch} • {service.lastDeployment.commit}</span>
                  </div>
                  <div className="flex items-center">
                    <Clock className="h-4 w-4 mr-2" />
                    <span>{formatDate(service.lastDeployment.timestamp)}</span>
                  </div>
                  {service.lastDeployment.buildHash && (
                    <div className="text-xs text-gray-500">
                      Build: {service.lastDeployment.buildHash}
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Detailed Actions (expandable) */}
            {selectedService === service.id && (
              <div className="mt-4 p-4 bg-gray-50 rounded-lg">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Recent Actions</h4>
                <div className="space-y-2">
                  {service.recentActions.map((action, index) => (
                    <div key={index} className="flex items-center space-x-3 text-sm">
                      {getStatusIcon(action.status)}
                      <span className="capitalize">{action.type}</span>
                      <span className="text-gray-500">•</span>
                      <span className="text-gray-500">{formatDate(action.timestamp)}</span>
                    </div>
                  ))}
                </div>
                <div className="mt-3 flex space-x-2">
                  <button className="btn-primary text-xs">
                    <Play className="h-4 w-4 mr-1" />
                    Trigger Build
                  </button>
                  <button className="btn-secondary text-xs">
                    View Logs
                  </button>
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {filteredServices.length === 0 && (
        <div className="text-center py-12">
          <Package className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No microservices found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {filter !== 'all' 
              ? `No services with ${filter} status found.`
              : 'No microservices have been discovered yet.'
            }
          </p>
        </div>
      )}
    </div>
  );
};

export default Microservices;