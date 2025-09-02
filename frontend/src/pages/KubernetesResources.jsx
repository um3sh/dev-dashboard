import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  Server, 
  GitBranch,
  Clock,
  CheckCircle,
  XCircle,
  AlertCircle,
  Activity,
  ExternalLink,
  Filter,
  FileText,
  Settings,
  Plus
} from 'lucide-react';

const KubernetesResources = () => {
  const { repoId } = useParams();
  const [resources, setResources] = useState([]);
  const [repository, setRepository] = useState(null);
  const [filter, setFilter] = useState('all');
  const [selectedResource, setSelectedResource] = useState(null);

  // Mock data for demonstration
  useEffect(() => {
    setRepository({
      id: 2,
      name: 'k8s-infrastructure',
      type: 'kubernetes'
    });

    setResources([
      {
        id: 1,
        name: 'user-service-deployment',
        path: 'overlays/prod/user-service/deployment.yaml',
        resourceType: 'Deployment',
        namespace: 'production',
        description: 'Production deployment for user service',
        lastDeployment: {
          status: 'success',
          commit: 'k8s-commit-1',
          branch: 'main',
          timestamp: '2024-01-20T14:30:00Z',
          buildHash: 'build-123'
        },
        recentActions: [
          { type: 'deployment', status: 'success', timestamp: '2024-01-20T14:30:00Z', prNumber: 456 },
          { type: 'deployment', status: 'success', timestamp: '2024-01-19T10:15:00Z', prNumber: 455 }
        ]
      },
      {
        id: 2,
        name: 'payment-service-configmap',
        path: 'overlays/prod/payment-service/configmap.yaml',
        resourceType: 'ConfigMap',
        namespace: 'production',
        description: 'Configuration for payment service',
        lastDeployment: {
          status: 'running',
          commit: 'k8s-commit-2',
          branch: 'feature/payment-config',
          timestamp: '2024-01-20T15:00:00Z',
          buildHash: 'build-124'
        },
        recentActions: [
          { type: 'deployment', status: 'running', timestamp: '2024-01-20T15:00:00Z', prNumber: 457 }
        ]
      },
      {
        id: 3,
        name: 'notification-service-ingress',
        path: 'overlays/staging/notification-service/ingress.yaml',
        resourceType: 'Ingress',
        namespace: 'staging',
        description: 'Ingress configuration for notification service',
        lastDeployment: {
          status: 'failure',
          commit: 'k8s-commit-3',
          branch: 'feature/new-ingress',
          timestamp: '2024-01-20T13:45:00Z',
          buildHash: null
        },
        recentActions: [
          { type: 'deployment', status: 'failure', timestamp: '2024-01-20T13:45:00Z', prNumber: 458 },
          { type: 'deployment', status: 'success', timestamp: '2024-01-19T16:20:00Z', prNumber: 454 }
        ]
      },
      {
        id: 4,
        name: 'database-secret',
        path: 'base/secrets/database-secret.yaml',
        resourceType: 'Secret',
        namespace: 'production',
        description: 'Database credentials and configuration',
        lastDeployment: {
          status: 'success',
          commit: 'k8s-commit-4',
          branch: 'main',
          timestamp: '2024-01-18T09:30:00Z',
          buildHash: 'build-120'
        },
        recentActions: [
          { type: 'deployment', status: 'success', timestamp: '2024-01-18T09:30:00Z', prNumber: 453 }
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

  const getResourceTypeColor = (resourceType) => {
    const colors = {
      'Deployment': 'bg-blue-100 text-blue-800',
      'ConfigMap': 'bg-green-100 text-green-800',
      'Secret': 'bg-red-100 text-red-800',
      'Service': 'bg-purple-100 text-purple-800',
      'Ingress': 'bg-orange-100 text-orange-800',
      'Pod': 'bg-gray-100 text-gray-800',
    };
    return colors[resourceType] || 'bg-gray-100 text-gray-800';
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const filteredResources = resources.filter(resource => {
    if (filter === 'all') return true;
    if (filter === 'success') return resource.lastDeployment.status === 'success';
    if (filter === 'failure') return resource.lastDeployment.status === 'failure';
    if (filter === 'running') return resource.lastDeployment.status === 'running';
    return true;
  });

  const groupedResources = filteredResources.reduce((groups, resource) => {
    const namespace = resource.namespace || 'default';
    if (!groups[namespace]) {
      groups[namespace] = [];
    }
    groups[namespace].push(resource);
    return groups;
  }, {});

  return (
    <div className="max-w-7xl mx-auto">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Kubernetes Resources</h1>
        <p className="mt-2 text-gray-600">
          {repository?.name ? `Resources in ${repository.name}` : 'Manage Kubernetes resources and overlays'}
        </p>
      </div>

      {/* Filters */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center space-x-4">
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
        
        <button className="btn-primary flex items-center">
          <Plus className="h-4 w-4 mr-2" />
          Create PR
        </button>
      </div>

      {/* Resources by Namespace */}
      <div className="space-y-8">
        {Object.entries(groupedResources).map(([namespace, namespaceResources]) => (
          <div key={namespace}>
            <div className="flex items-center space-x-3 mb-4">
              <h2 className="text-xl font-semibold text-gray-900">
                {namespace}
              </h2>
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                {namespaceResources.length} resources
              </span>
            </div>
            
            <div className="grid gap-4">
              {namespaceResources.map((resource) => (
                <div key={resource.id} className="card">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex items-start space-x-4">
                      <div className="p-3 bg-purple-100 rounded-lg">
                        <Server className="h-6 w-6 text-purple-600" />
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center space-x-2 mb-2">
                          <h3 className="text-lg font-semibold text-gray-900">{resource.name}</h3>
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getResourceTypeColor(resource.resourceType)}`}>
                            {resource.resourceType}
                          </span>
                        </div>
                        <p className="text-gray-600 mb-2">{resource.description}</p>
                        <div className="flex items-center text-sm text-gray-500">
                          <FileText className="h-4 w-4 mr-1" />
                          <span>{resource.path}</span>
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={() => setSelectedResource(selectedResource === resource.id ? null : resource.id)}
                      className="btn-secondary"
                    >
                      View Details
                    </button>
                  </div>

                  {/* Last Deployment Status */}
                  <div className="bg-gray-50 p-4 rounded-lg">
                    <div className="flex items-center justify-between mb-2">
                      <h4 className="text-sm font-medium text-gray-700">Last Deployment</h4>
                      <span className={getStatusClass(resource.lastDeployment.status)}>
                        {resource.lastDeployment.status}
                      </span>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-3 gap-4 text-sm text-gray-600">
                      <div className="flex items-center">
                        <GitBranch className="h-4 w-4 mr-2" />
                        <span>{resource.lastDeployment.branch} • {resource.lastDeployment.commit}</span>
                      </div>
                      <div className="flex items-center">
                        <Clock className="h-4 w-4 mr-2" />
                        <span>{formatDate(resource.lastDeployment.timestamp)}</span>
                      </div>
                      {resource.lastDeployment.buildHash && (
                        <div className="text-xs text-gray-500">
                          Build: {resource.lastDeployment.buildHash}
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Detailed Actions (expandable) */}
                  {selectedResource === resource.id && (
                    <div className="mt-4 p-4 bg-gray-50 rounded-lg">
                      <h4 className="text-sm font-medium text-gray-700 mb-3">Recent Actions</h4>
                      <div className="space-y-3">
                        {resource.recentActions.map((action, index) => (
                          <div key={index} className="flex items-center justify-between p-3 bg-white rounded border">
                            <div className="flex items-center space-x-3 text-sm">
                              {getStatusIcon(action.status)}
                              <span className="capitalize">{action.type}</span>
                              <span className="text-gray-500">•</span>
                              <span className="text-gray-500">{formatDate(action.timestamp)}</span>
                              {action.prNumber && (
                                <>
                                  <span className="text-gray-500">•</span>
                                  <a 
                                    href={`#pr-${action.prNumber}`}
                                    className="text-blue-600 hover:text-blue-500"
                                  >
                                    PR #{action.prNumber}
                                  </a>
                                </>
                              )}
                            </div>
                            <div className="flex space-x-2">
                              <button className="text-gray-400 hover:text-blue-600">
                                <ExternalLink className="h-4 w-4" />
                              </button>
                            </div>
                          </div>
                        ))}
                      </div>
                      <div className="mt-3 flex space-x-2">
                        <button className="btn-primary text-xs">
                          <GitBranch className="h-4 w-4 mr-1" />
                          Create PR
                        </button>
                        <button className="btn-secondary text-xs">
                          <Settings className="h-4 w-4 mr-1" />
                          Edit Overlay
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        ))}
      </div>

      {Object.keys(groupedResources).length === 0 && (
        <div className="text-center py-12">
          <Server className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No Kubernetes resources found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {filter !== 'all' 
              ? `No resources with ${filter} status found.`
              : 'No Kubernetes resources have been discovered yet.'
            }
          </p>
        </div>
      )}
    </div>
  );
};

export default KubernetesResources;