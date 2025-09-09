import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  Package,
  Cloud,
  MapPin,
  Tag,
  Hash,
  Calendar,
  Activity,
  ExternalLink,
  RefreshCw,
  AlertCircle,
  GitCommit
} from 'lucide-react';

const ServiceDeployments = () => {
  const { serviceId } = useParams();
  const [service, setService] = useState(null);
  const [commitDeployments, setCommitDeployments] = useState([]);
  const [uniqueDeploymentEnvs, setUniqueDeploymentEnvs] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (serviceId) {
      loadServiceDeployments();
    }
  }, [serviceId]);

  const loadServiceDeployments = async () => {
    setLoading(true);
    try {
      // Load service info
      const allServices = await window.go.main.App.GetMicroservices(0);
      const selectedService = allServices?.find(s => s.id === parseInt(serviceId));
      setService(selectedService || null);

      if (selectedService) {
        // Load service commit deployments
        try {
          const serviceCommitDeployments = await window.go.main.App.GetServiceCommitDeployments(parseInt(serviceId));
          setCommitDeployments(serviceCommitDeployments || []);
          
          // Extract unique deployment environments/regions/namespaces for table headers
          const envSet = new Set();
          (serviceCommitDeployments || []).forEach(commitDeployment => {
            commitDeployment.deployments.forEach(deployment => {
              const envKey = `${deployment.environment}-${deployment.region}-${deployment.namespace}`;
              envSet.add(JSON.stringify({
                environment: deployment.environment,
                region: deployment.region,
                namespace: deployment.namespace
              }));
            });
          });
          
          const uniqueEnvs = Array.from(envSet).map(envStr => JSON.parse(envStr));
          setUniqueDeploymentEnvs(uniqueEnvs);
        } catch (error) {
          console.error('Failed to load commit deployments:', error);
          setCommitDeployments([]);
          setUniqueDeploymentEnvs([]);
        }
      }
    } catch (error) {
      console.error('Failed to load service deployments:', error);
      setService(null);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const formatCommitHash = (hash) => {
    return hash?.substring(0, 7) || 'N/A';
  };

  const getEnvironmentColor = (environment) => {
    switch (environment.toLowerCase()) {
      case 'prd':
      case 'prod':
      case 'production':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'stg':
      case 'staging':
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
      case 'dev':
      case 'development':
        return 'bg-blue-100 text-blue-800 border-blue-200';
      default:
        return 'bg-gray-100 text-gray-800 border-gray-200';
    }
  };

  const getRegionFlag = (region) => {
    // Simple mapping of regions to flags/indicators
    if (region.includes('us-west')) return '🇺🇸 West';
    if (region.includes('us-east')) return '🇺🇸 East';
    if (region.includes('eu-')) return '🇪🇺 EU';
    if (region.includes('ap-')) return '🌏 APAC';
    return region;
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="flex items-center justify-center py-12">
          <Activity className="h-8 w-8 text-blue-500 animate-spin" />
          <span className="ml-2 text-lg text-gray-600">Loading deployments...</span>
        </div>
      </div>
    );
  }

  if (!service) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="text-center py-12">
          <Package className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">Service not found</h3>
          <p className="mt-1 text-sm text-gray-500">
            The requested service could not be found.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center">
            <div className="p-3 bg-green-100 rounded-lg mr-4">
              <Cloud className="h-8 w-8 text-green-600" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-gray-900">Deployments</h1>
              <p className="mt-1 text-gray-600">
                Current deployment status for <strong>{service.name}</strong> service
              </p>
              <div className="flex items-center mt-2 text-sm text-gray-500">
                <ExternalLink className="h-4 w-4 mr-1" />
                <span>{service.path}</span>
              </div>
            </div>
          </div>
          <button
            onClick={loadServiceDeployments}
            className="btn-secondary flex items-center"
          >
            <RefreshCw className="h-4 w-4 mr-2" />
            Refresh
          </button>
        </div>
      </div>

      {/* Deployments Overview */}
      <div className="mb-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Deployment Overview</h2>
        
        {commitDeployments.length === 0 ? (
          <div className="card text-center py-12">
            <Cloud className="mx-auto h-12 w-12 text-gray-400 mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No commits or deployments found</h3>
            <p className="text-sm text-gray-500 mb-4">
              This service has no recent commits or deployment tracking is not configured.
            </p>
            <div className="text-xs text-gray-400 space-y-1">
              <p>To track deployments:</p>
              <p>1. Add a Kubernetes resources repository</p>
              <p>2. Ensure kustomization.yaml files contain image tags for this service</p>
              <p>3. Run a sync to discover deployments</p>
            </div>
          </div>
        ) : (
          <div className="card overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider sticky left-0 bg-gray-50 z-10">
                      <div className="flex items-center">
                        <Hash className="h-4 w-4 mr-1" />
                        Commit
                      </div>
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      <div className="flex items-center">
                        <GitCommit className="h-4 w-4 mr-1" />
                        Message
                      </div>
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      <div className="flex items-center">
                        <Calendar className="h-4 w-4 mr-1" />
                        Date
                      </div>
                    </th>
                    {/* Dynamic environment/region/namespace columns */}
                    {uniqueDeploymentEnvs.map((env, index) => (
                      <th key={index} className="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">
                        <div className="text-center">
                          <div className={`inline-flex px-2 py-1 text-xs font-medium rounded-full border ${getEnvironmentColor(env.environment)}`}>
                            {env.environment}
                          </div>
                          <div className="text-xs text-gray-400 mt-1">
                            {getRegionFlag(env.region)}
                          </div>
                          {env.namespace && (
                            <div className="text-xs text-gray-500 mt-1 font-mono">
                              ns: {env.namespace}
                            </div>
                          )}
                        </div>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {commitDeployments.map((commitDeployment, index) => (
                    <tr key={index} className="hover:bg-gray-50">
                      <td className="px-6 py-4 whitespace-nowrap sticky left-0 bg-white z-10">
                        <div className="flex items-center">
                          <span className="font-mono text-sm text-gray-900">
                            {formatCommitHash(commitDeployment.commit.hash)}
                          </span>
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm text-gray-900 max-w-xs truncate" title={commitDeployment.commit.message}>
                          {commitDeployment.commit.message}
                        </div>
                        <div className="text-xs text-gray-500">
                          by {commitDeployment.commit.author}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className="text-sm text-gray-500">
                          {formatDate(commitDeployment.commit.date)}
                        </span>
                      </td>
                      {/* Deployment status cells */}
                      {uniqueDeploymentEnvs.map((env, deployIndex) => {
                        // Find matching deployment for this environment/region/namespace
                        const matchingDeployment = commitDeployment.deployments.find(d => 
                          d.environment === env.environment && 
                          d.region === env.region && 
                          d.namespace === env.namespace
                        );
                        
                        return (
                          <td key={deployIndex} className="px-6 py-4 whitespace-nowrap text-center">
                            {matchingDeployment?.is_deployed ? (
                              <div>
                                <span className="font-mono text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                                  {matchingDeployment.tag}
                                </span>
                                <div className="text-xs text-gray-400 mt-1">
                                  {formatDate(matchingDeployment.deployed_at)}
                                </div>
                              </div>
                            ) : (
                              <div className="text-gray-400 text-xs">
                                Not deployed
                              </div>
                            )}
                          </td>
                        );
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>

      {/* Summary Cards */}
      {commitDeployments.length > 0 && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="card">
            <div className="flex items-center">
              <div className="p-3 bg-blue-100 rounded-lg mr-4">
                <GitCommit className="h-6 w-6 text-blue-600" />
              </div>
              <div>
                <p className="text-2xl font-semibold text-gray-900">
                  {commitDeployments.length}
                </p>
                <p className="text-sm text-gray-500">Recent Commits</p>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center">
              <div className="p-3 bg-green-100 rounded-lg mr-4">
                <Cloud className="h-6 w-6 text-green-600" />
              </div>
              <div>
                <p className="text-2xl font-semibold text-gray-900">
                  {uniqueDeploymentEnvs.length}
                </p>
                <p className="text-sm text-gray-500">Environments</p>
              </div>
            </div>
          </div>

          <div className="card">
            <div className="flex items-center">
              <div className="p-3 bg-purple-100 rounded-lg mr-4">
                <Activity className="h-6 w-6 text-purple-600" />
              </div>
              <div>
                <p className="text-2xl font-semibold text-gray-900">
                  {commitDeployments.reduce((count, commit) => 
                    count + commit.deployments.filter(d => d.is_deployed).length, 0
                  )}
                </p>
                <p className="text-sm text-gray-500">Active Deployments</p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ServiceDeployments;