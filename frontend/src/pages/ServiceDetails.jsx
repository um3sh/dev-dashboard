import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  Package, 
  GitPullRequest,
  GitCommit,
  Clock,
  GitBranch,
  CheckCircle,
  XCircle,
  AlertCircle,
  Activity,
  ExternalLink,
  User,
  Calendar,
  Hash
} from 'lucide-react';

const ServiceDetails = () => {
  const { serviceId } = useParams();
  const [service, setService] = useState(null);
  const [pullRequests, setPullRequests] = useState([]);
  const [commits, setCommits] = useState([]);
  const [loading, setLoading] = useState(true);
  const [githubIntegrationAvailable, setGithubIntegrationAvailable] = useState(true);

  useEffect(() => {
    if (serviceId) {
      loadServiceDetails();
    }
  }, [serviceId]);

  const loadServiceDetails = async () => {
    setLoading(true);
    try {
      // Load service info
      const allServices = await window.go.main.App.GetMicroservices(0);
      const selectedService = allServices?.find(s => s.id === parseInt(serviceId));
      setService(selectedService || null);

      if (selectedService) {
        // Load service-specific PRs and commits
        // Note: These methods need to be implemented in the backend
        try {
          const prs = await window.go.main.App.GetServicePullRequests(parseInt(serviceId));
          setPullRequests(prs || []);
        } catch (error) {
          console.error('Failed to load pull requests:', error);
          setPullRequests([]);
          // Only set GitHub integration as unavailable if it's a real error, not just empty results
          if (error.message && error.message.includes('no GitHub token')) {
            setGithubIntegrationAvailable(false);
          }
        }

        try {
          const commitHistory = await window.go.main.App.GetServiceCommits(parseInt(serviceId));
          setCommits(commitHistory || []);
        } catch (error) {
          console.error('Failed to load commits:', error);
          setCommits([]);
          // Only set GitHub integration as unavailable if it's a real error, not just empty results
          if (error.message && error.message.includes('no GitHub token')) {
            setGithubIntegrationAvailable(false);
          }
        }
      }
    } catch (error) {
      console.error('Failed to load service details:', error);
      setService(null);
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'success':
      case 'merged':
        return <CheckCircle className="h-5 w-5 text-green-500" />;
      case 'failure':
      case 'closed':
        return <XCircle className="h-5 w-5 text-red-500" />;
      case 'running':
      case 'open':
        return <Activity className="h-5 w-5 text-blue-500 animate-pulse" />;
      default:
        return <AlertCircle className="h-5 w-5 text-yellow-500" />;
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

  const formatCommitHash = (hash) => {
    return hash?.substring(0, 7) || '';
  };

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="flex items-center justify-center py-12">
          <Activity className="h-8 w-8 text-blue-500 animate-spin" />
          <span className="ml-2 text-lg text-gray-600">Loading service details...</span>
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
            The requested service could not be found. Service ID: {serviceId}
          </p>
          <p className="mt-1 text-xs text-gray-400">
            Available services should have IDs like 3, 4, etc.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto">
      {/* Service Header */}
      <div className="mb-8">
        <div className="flex items-center mb-4">
          <div className="p-3 bg-blue-100 rounded-lg mr-4">
            <Package className="h-8 w-8 text-blue-600" />
          </div>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">{service.name}</h1>
            <p className="mt-1 text-gray-600">{service.description || 'No description available'}</p>
            <div className="flex items-center mt-2 text-sm text-gray-500">
              <ExternalLink className="h-4 w-4 mr-1" />
              <span>{service.path}</span>
            </div>
          </div>
        </div>
      </div>

      {/* GitHub Integration Notice */}
      {!githubIntegrationAvailable && (
        <div className="mb-6 bg-amber-50 border border-amber-200 rounded-lg p-4">
          <div className="flex items-center">
            <AlertCircle className="h-5 w-5 text-amber-600 mr-2" />
            <div className="flex-1">
              <h4 className="text-sm font-medium text-amber-800">GitHub Integration Not Available</h4>
              <p className="text-sm text-amber-700 mt-1">
                Configure your GitHub Personal Access Token in <strong>Settings</strong> to view pull requests and commit history from GitHub.
              </p>
            </div>
          </div>
        </div>
      )}

      {/* Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Pull Requests Section */}
        <div className="card">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-gray-900 flex items-center">
              <GitPullRequest className="h-6 w-6 mr-2 text-blue-600" />
              Pull Requests
            </h2>
            <span className="text-sm text-gray-500">
              {pullRequests.length} total
            </span>
          </div>
          
          <div className="space-y-4 max-h-96 overflow-y-auto">
            {pullRequests.length > 0 ? (
              pullRequests.map((pr) => (
                <div key={pr.id} className="border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors">
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex items-center space-x-2">
                      {getStatusIcon(pr.status)}
                      <h3 className="font-medium text-gray-900">#{pr.number}</h3>
                      <span className="text-sm text-gray-500">•</span>
                      <span className="text-sm text-gray-500">{pr.title}</span>
                    </div>
                  </div>
                  
                  <div className="flex items-center space-x-4 text-sm text-gray-600">
                    <div className="flex items-center">
                      <User className="h-4 w-4 mr-1" />
                      <span>{pr.author}</span>
                    </div>
                    <div className="flex items-center">
                      <GitBranch className="h-4 w-4 mr-1" />
                      <span>{pr.branch}</span>
                    </div>
                    <div className="flex items-center">
                      <Clock className="h-4 w-4 mr-1" />
                      <span>{formatDate(pr.createdAt)}</span>
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8">
                <GitPullRequest className="mx-auto h-8 w-8 text-gray-400" />
                <p className="mt-2 text-sm text-gray-500">No pull requests found for this service</p>
                <p className="mt-1 text-xs text-gray-400">Configure GitHub token in Settings to see GitHub data</p>
              </div>
            )}
          </div>
        </div>

        {/* Commit History Section */}
        <div className="card">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-semibold text-gray-900 flex items-center">
              <GitCommit className="h-6 w-6 mr-2 text-green-600" />
              Recent Commits
            </h2>
            <span className="text-sm text-gray-500">
              {commits.length} total
            </span>
          </div>
          
          <div className="space-y-4 max-h-96 overflow-y-auto">
            {commits.length > 0 ? (
              commits.map((commit) => (
                <div key={commit.hash} className="border border-gray-200 rounded-lg p-4 hover:border-gray-300 transition-colors">
                  <div className="flex items-start justify-between mb-2">
                    <div className="flex-1">
                      <p className="font-medium text-gray-900 text-sm">{commit.message}</p>
                      <div className="flex items-center space-x-4 mt-2 text-sm text-gray-600">
                        <div className="flex items-center">
                          <Hash className="h-4 w-4 mr-1" />
                          <span className="font-mono">{formatCommitHash(commit.hash)}</span>
                        </div>
                        <div className="flex items-center">
                          <User className="h-4 w-4 mr-1" />
                          <span>{commit.author}</span>
                        </div>
                        <div className="flex items-center">
                          <Calendar className="h-4 w-4 mr-1" />
                          <span>{formatDate(commit.date)}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ))
            ) : (
              <div className="text-center py-8">
                <GitCommit className="mx-auto h-8 w-8 text-gray-400" />
                <p className="mt-2 text-sm text-gray-500">No commit history found for this service</p>
                <p className="mt-1 text-xs text-gray-400">Configure GitHub token in Settings to see GitHub data</p>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ServiceDetails;