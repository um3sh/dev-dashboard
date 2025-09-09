import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  Package,
  GitCommit,
  User,
  Calendar,
  Hash,
  Activity,
  ExternalLink,
  Clock,
  Search,
  Filter
} from 'lucide-react';

const ServiceDeploymentHistory = () => {
  const { serviceId } = useParams();
  const [service, setService] = useState(null);
  const [commits, setCommits] = useState([]);
  const [deployments, setDeployments] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [authorFilter, setAuthorFilter] = useState('all');

  useEffect(() => {
    if (serviceId) {
      loadDeploymentHistory();
    }
  }, [serviceId]);

  const loadDeploymentHistory = async () => {
    setLoading(true);
    try {
      // Load service info
      const allServices = await window.go.main.App.GetMicroservices(0);
      const selectedService = allServices?.find(s => s.id === parseInt(serviceId));
      setService(selectedService || null);

      if (selectedService) {
        // Load both deployment history and current deployments
        try {
          const [historyCommits, serviceDeployments] = await Promise.all([
            window.go.main.App.GetServiceDeploymentHistory(parseInt(serviceId)),
            window.go.main.App.GetServiceDeployments(parseInt(serviceId))
          ]);
          
          setCommits(historyCommits || []);
          setDeployments(serviceDeployments || []);
        } catch (error) {
          console.error('Failed to load deployment history:', error);
          setCommits([]);
          setDeployments([]);
        }
      }
    } catch (error) {
      console.error('Failed to load deployment history:', error);
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
    return hash?.substring(0, 7) || '';
  };

  const getRelativeTime = (dateString) => {
    const now = new Date();
    const date = new Date(dateString);
    const diffInSeconds = Math.floor((now - date) / 1000);
    
    if (diffInSeconds < 60) return `${diffInSeconds}s ago`;
    if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}m ago`;
    if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}h ago`;
    if (diffInSeconds < 2592000) return `${Math.floor(diffInSeconds / 86400)}d ago`;
    return formatDate(dateString);
  };

  const getCommitDeploymentStatus = (commitHash) => {
    if (!commitHash || !deployments.length) return null;
    
    // Use full hash comparison for better accuracy
    const deployedIn = deployments.filter(d => 
      d.commit_sha && d.commit_sha === commitHash
    );
    
    if (deployedIn.length === 0) return null;
    
    return deployedIn;
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

  // Get unique authors for filter
  const authors = [...new Set(commits.map(commit => commit.author))].filter(Boolean);

  // Filter commits based on search and author
  const filteredCommits = commits.filter(commit => {
    const matchesSearch = !searchTerm || 
      commit.message.toLowerCase().includes(searchTerm.toLowerCase()) ||
      commit.hash.toLowerCase().includes(searchTerm.toLowerCase()) ||
      commit.author.toLowerCase().includes(searchTerm.toLowerCase());
    
    const matchesAuthor = authorFilter === 'all' || commit.author === authorFilter;
    
    return matchesSearch && matchesAuthor;
  });

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="flex items-center justify-center py-12">
          <Activity className="h-8 w-8 text-blue-500 animate-spin" />
          <span className="ml-2 text-lg text-gray-600">Loading deployment history...</span>
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
        <div className="flex items-center mb-4">
          <div className="p-3 bg-purple-100 rounded-lg mr-4">
            <Clock className="h-8 w-8 text-purple-600" />
          </div>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Deployment History</h1>
            <p className="mt-1 text-gray-600">
              Commit history and deployment status for <strong>{service.name}</strong> service
            </p>
            <div className="flex items-center mt-2 text-sm text-gray-500">
              <ExternalLink className="h-4 w-4 mr-1" />
              <span>{service.path}</span>
            </div>
          </div>
        </div>
      </div>

      {/* Filters and Search */}
      <div className="flex flex-col sm:flex-row gap-4 mb-6">
        {/* Search */}
        <div className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-gray-400" />
          <input
            type="text"
            placeholder="Search commits by message, hash, or author..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            className="pl-10 pr-4 py-2 w-full border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
        </div>

        {/* Author Filter */}
        <div className="flex items-center space-x-2">
          <Filter className="h-4 w-4 text-gray-400" />
          <select
            value={authorFilter}
            onChange={(e) => setAuthorFilter(e.target.value)}
            className="border border-gray-300 rounded-lg px-3 py-2 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          >
            <option value="all">All authors</option>
            {authors.map(author => (
              <option key={author} value={author}>{author}</option>
            ))}
          </select>
        </div>

        <div className="text-sm text-gray-500 flex items-center">
          {filteredCommits.length} of {commits.length} commits
        </div>
      </div>

      {/* Commits List with Deployment Status */}
      <div className="space-y-4">
        {filteredCommits.map((commit) => {
          const deploymentStatus = getCommitDeploymentStatus(commit.hash);
          
          return (
            <div key={commit.hash} className="card hover:shadow-md transition-shadow">
              <div className="flex items-start space-x-4">
                {/* Commit Icon */}
                <div className="flex-shrink-0 mt-1">
                  <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
                    <GitCommit className="h-4 w-4 text-gray-600" />
                  </div>
                </div>

                {/* Commit Details */}
                <div className="flex-1 min-w-0">
                  <div className="flex items-start justify-between">
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium text-gray-900 mb-1">
                        {commit.message}
                      </p>
                      <div className="flex items-center space-x-4 text-xs text-gray-500 mb-2">
                        <div className="flex items-center">
                          <Hash className="h-3 w-3 mr-1" />
                          <span className="font-mono">{formatCommitHash(commit.hash)}</span>
                        </div>
                        <div className="flex items-center">
                          <User className="h-3 w-3 mr-1" />
                          <span>{commit.author}</span>
                        </div>
                        <div className="flex items-center">
                          <Calendar className="h-3 w-3 mr-1" />
                          <span title={formatDate(commit.date)}>
                            {getRelativeTime(commit.date)}
                          </span>
                        </div>
                      </div>

                      {/* Deployment Status */}
                      {deploymentStatus && deploymentStatus.length > 0 ? (
                        <div className="mt-3">
                          <div className="text-xs font-medium text-gray-700 mb-1">
                            Deployed in:
                          </div>
                          <div className="flex flex-wrap gap-1">
                            {deploymentStatus.map((deployment, idx) => (
                              <span
                                key={idx}
                                className={`inline-flex items-center px-2 py-1 text-xs font-medium rounded-full border ${getEnvironmentColor(deployment.environment)}`}
                              >
                                {deployment.environment} • {deployment.region}
                              </span>
                            ))}
                          </div>
                        </div>
                      ) : (
                        <div className="mt-3">
                          <span className="inline-flex items-center px-2 py-1 text-xs font-medium rounded-full bg-gray-100 text-gray-600 border border-gray-200">
                            Not deployed
                          </span>
                        </div>
                      )}
                    </div>
                    
                    {/* Commit Actions */}
                    <div className="flex-shrink-0 ml-4">
                      <button className="btn-secondary text-xs p-2">
                        <ExternalLink className="h-3 w-3" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {filteredCommits.length === 0 && (
        <div className="text-center py-12">
          <GitCommit className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No commits found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {searchTerm || authorFilter !== 'all'
              ? 'No commits match your current filters.'
              : 'No commit history found for this service.'
            }
          </p>
          {commits.length === 0 && (
            <p className="mt-2 text-xs text-gray-400">
              Configure GitHub token in Settings to see GitHub data
            </p>
          )}
        </div>
      )}
    </div>
  );
};

export default ServiceDeploymentHistory;