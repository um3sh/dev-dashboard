import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
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
  const navigate = useNavigate();
  const [services, setServices] = useState([]);
  const [repository, setRepository] = useState(null);
  const [filter, setFilter] = useState('all');
  const [selectedService, setSelectedService] = useState(null);

  // Load real microservices data
  useEffect(() => {
    loadMicroservices();
    if (repoId) {
      loadRepository();
    }
  }, [repoId]);

  const loadMicroservices = async () => {
    try {
      // If no repoId, get all microservices (pass 0), otherwise get for specific repo
      const repositoryId = repoId ? parseInt(repoId) : 0;
      const microservices = await window.go.main.App.GetMicroservices(repositoryId);
      
      // Transform the data to include action information
      const servicesWithActions = await Promise.all(
        (microservices || []).map(async (service) => {
          try {
            const actions = await window.go.main.App.GetMicroserviceActions(service.id, 10);
            const buildActions = actions?.filter(a => a.type === 'build') || [];
            const deployActions = actions?.filter(a => a.type === 'deployment') || [];
            
            return {
              ...service,
              lastBuild: buildActions.length > 0 ? buildActions[0] : null,
              lastDeployment: deployActions.length > 0 ? deployActions[0] : null
            };
          } catch (err) {
            console.error(`Failed to load actions for service ${service.name}:`, err);
            return {
              ...service,
              lastBuild: null,
              lastDeployment: null
            };
          }
        })
      );
      
      setServices(servicesWithActions);
    } catch (error) {
      console.error('Failed to load microservices:', error);
      setServices([]);
    }
  };

  const loadRepository = async () => {
    if (!repoId) return;
    
    try {
      const repositories = await window.go.main.App.GetRepositories();
      const repo = repositories?.find(r => r.id === parseInt(repoId));
      setRepository(repo || null);
    } catch (error) {
      console.error('Failed to load repository:', error);
      setRepository(null);
    }
  };

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

  const handleViewDetails = (serviceId) => {
    // Navigate to service details page, which will trigger the Layout component
    // to update the dropdown and show service-specific navigation
    navigate(`/service/${serviceId}`);
  };

  const filteredServices = services.filter(service => {
    if (filter === 'all') return true;
    if (filter === 'success') return service.lastBuild?.status === 'success';
    if (filter === 'failure') return service.lastBuild?.status === 'failure';
    if (filter === 'running') return service.lastBuild?.status === 'running';
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
              <div className="flex space-x-2">
                <button
                  onClick={() => handleViewDetails(service.id)}
                  className="btn-primary"
                >
                  More Details
                </button>
                <button
                  onClick={() => setSelectedService(selectedService === service.id ? null : service.id)}
                  className="btn-secondary"
                >
                  {selectedService === service.id ? 'Hide' : 'Quick View'}
                </button>
              </div>
            </div>

            {/* Build and Deployment Status */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Last Build */}
              <div className="bg-gray-50 p-4 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="text-sm font-medium text-gray-700">Last Build</h4>
                  <span className={getStatusClass(service.lastBuild?.status || 'pending')}>
                    {service.lastBuild?.status || 'No builds'}
                  </span>
                </div>
                <div className="space-y-1 text-sm text-gray-600">
                  {service.lastBuild ? (
                    <>
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
                    </>
                  ) : (
                    <div className="text-gray-500">No build history available</div>
                  )}
                </div>
              </div>

              {/* Last Deployment */}
              <div className="bg-gray-50 p-4 rounded-lg">
                <div className="flex items-center justify-between mb-2">
                  <h4 className="text-sm font-medium text-gray-700">Last Deployment</h4>
                  <span className={getStatusClass(service.lastDeployment?.status || 'pending')}>
                    {service.lastDeployment?.status || 'No deployments'}
                  </span>
                </div>
                <div className="space-y-1 text-sm text-gray-600">
                  {service.lastDeployment ? (
                    <>
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
                    </>
                  ) : (
                    <div className="text-gray-500">No deployment history available</div>
                  )}
                </div>
              </div>
            </div>

            {/* Detailed Actions (expandable) */}
            {selectedService === service.id && (
              <div className="mt-4 p-4 bg-gray-50 rounded-lg">
                <h4 className="text-sm font-medium text-gray-700 mb-3">Recent Actions</h4>
                <div className="space-y-2">
                  {service.lastBuild && (
                    <div className="flex items-center space-x-3 text-sm">
                      {getStatusIcon(service.lastBuild.status)}
                      <span className="capitalize">build</span>
                      <span className="text-gray-500">•</span>
                      <span className="text-gray-500">{formatDate(service.lastBuild.timestamp)}</span>
                    </div>
                  )}
                  {service.lastDeployment && (
                    <div className="flex items-center space-x-3 text-sm">
                      {getStatusIcon(service.lastDeployment.status)}
                      <span className="capitalize">deployment</span>
                      <span className="text-gray-500">•</span>
                      <span className="text-gray-500">{formatDate(service.lastDeployment.timestamp)}</span>
                    </div>
                  )}
                  {!service.lastBuild && !service.lastDeployment && (
                    <div className="text-gray-500 text-sm">No recent actions available</div>
                  )}
                </div>
                <div className="mt-3 flex space-x-2">
                  <button 
                    onClick={() => handleViewDetails(service.id)}
                    className="btn-primary text-xs"
                  >
                    Full Details
                  </button>
                  <button className="btn-secondary text-xs">
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