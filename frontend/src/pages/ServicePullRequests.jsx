import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  GitPullRequest,
  User,
  GitBranch,
  Clock,
  CheckCircle,
  XCircle,
  Activity,
  AlertCircle,
  ExternalLink,
  Filter,
  Package
} from 'lucide-react';

const ServicePullRequests = () => {
  const { serviceId } = useParams();
  const [service, setService] = useState(null);
  const [pullRequests, setPullRequests] = useState([]);
  const [loading, setLoading] = useState(true);
  const [filter, setFilter] = useState('all');

  useEffect(() => {
    if (serviceId) {
      loadServicePullRequests();
    }
  }, [serviceId]);

  const loadServicePullRequests = async () => {
    setLoading(true);
    try {
      // Load service info
      const allServices = await window.go.main.App.GetMicroservices(0);
      const selectedService = allServices?.find(s => s.id === parseInt(serviceId));
      setService(selectedService || null);

      if (selectedService) {
        // Load service-specific PRs
        try {
          const prs = await window.go.main.App.GetServicePullRequests(parseInt(serviceId));
          setPullRequests(prs || []);
        } catch (error) {
          console.error('Failed to load pull requests:', error);
          setPullRequests([]);
        }
      }
    } catch (error) {
      console.error('Failed to load service pull requests:', error);
      setService(null);
    } finally {
      setLoading(false);
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'merged':
        return <CheckCircle className="h-5 w-5 text-purple-500" />;
      case 'closed':
        return <XCircle className="h-5 w-5 text-red-500" />;
      case 'open':
        return <Activity className="h-5 w-5 text-green-500" />;
      default:
        return <AlertCircle className="h-5 w-5 text-yellow-500" />;
    }
  };

  const getStatusClass = (status) => {
    switch (status) {
      case 'merged':
        return 'bg-purple-100 text-purple-800 border-purple-200';
      case 'closed':
        return 'bg-red-100 text-red-800 border-red-200';
      case 'open':
        return 'bg-green-100 text-green-800 border-green-200';
      default:
        return 'bg-yellow-100 text-yellow-800 border-yellow-200';
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

  const filteredPRs = pullRequests.filter(pr => {
    if (filter === 'all') return true;
    return pr.status === filter;
  });

  if (loading) {
    return (
      <div className="max-w-7xl mx-auto">
        <div className="flex items-center justify-center py-12">
          <Activity className="h-8 w-8 text-blue-500 animate-spin" />
          <span className="ml-2 text-lg text-gray-600">Loading pull requests...</span>
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
          <div className="p-3 bg-blue-100 rounded-lg mr-4">
            <GitPullRequest className="h-8 w-8 text-blue-600" />
          </div>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Pull Requests</h1>
            <p className="mt-1 text-gray-600">
              Pull requests for <strong>{service.name}</strong> service
            </p>
            <div className="flex items-center mt-2 text-sm text-gray-500">
              <ExternalLink className="h-4 w-4 mr-1" />
              <span>{service.path}</span>
            </div>
          </div>
        </div>
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
            { key: 'open', label: 'Open' },
            { key: 'merged', label: 'Merged' },
            { key: 'closed', label: 'Closed' }
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
        <div className="ml-auto text-sm text-gray-500">
          {filteredPRs.length} of {pullRequests.length} pull requests
        </div>
      </div>

      {/* Pull Requests List */}
      <div className="space-y-4">
        {filteredPRs.map((pr) => (
          <div key={pr.id} className="card hover:shadow-md transition-shadow">
            <div className="flex items-start justify-between mb-4">
              <div className="flex-1">
                <div className="flex items-center space-x-3 mb-2">
                  {getStatusIcon(pr.status)}
                  <h3 className="text-lg font-semibold text-gray-900">
                    #{pr.number} - {pr.title}
                  </h3>
                  <span className={`px-2 py-1 text-xs font-medium rounded-full border ${getStatusClass(pr.status)}`}>
                    {pr.status}
                  </span>
                </div>
                
                <div className="flex items-center space-x-6 text-sm text-gray-600">
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
            </div>

            {/* PR Actions */}
            <div className="flex space-x-2">
              <button className="btn-secondary text-xs p-2">
                <ExternalLink className="h-3 w-3" />
              </button>
            </div>
          </div>
        ))}
      </div>

      {filteredPRs.length === 0 && (
        <div className="text-center py-12">
          <GitPullRequest className="mx-auto h-12 w-12 text-gray-400" />
          <h3 className="mt-2 text-sm font-medium text-gray-900">No pull requests found</h3>
          <p className="mt-1 text-sm text-gray-500">
            {filter !== 'all' 
              ? `No ${filter} pull requests found for this service.`
              : 'No pull requests found for this service.'
            }
          </p>
          {pullRequests.length === 0 && (
            <p className="mt-2 text-xs text-gray-400">
              Configure GitHub token in Settings to see GitHub data
            </p>
          )}
        </div>
      )}
    </div>
  );
};

export default ServicePullRequests;