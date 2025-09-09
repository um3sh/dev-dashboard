import React, { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { 
  GitCommit,
  User,
  Calendar,
  Hash,
  Activity,
  ExternalLink,
  Package,
  Search,
  Filter
} from 'lucide-react';

const ServiceCommits = () => {
  const { serviceId } = useParams();
  const [service, setService] = useState(null);
  const [commits, setCommits] = useState([]);
  const [loading, setLoading] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [authorFilter, setAuthorFilter] = useState('all');

  useEffect(() => {
    if (serviceId) {
      loadServiceCommits();
    }
  }, [serviceId]);

  const loadServiceCommits = async () => {
    setLoading(true);
    try {
      // Load service info
      const allServices = await window.go.main.App.GetMicroservices(0);
      const selectedService = allServices?.find(s => s.id === parseInt(serviceId));
      setService(selectedService || null);

      if (selectedService) {
        // Load service-specific commits
        try {
          const commitHistory = await window.go.main.App.GetServiceCommits(parseInt(serviceId));
          setCommits(commitHistory || []);
        } catch (error) {
          console.error('Failed to load commits:', error);
          setCommits([]);
        }
      }
    } catch (error) {
      console.error('Failed to load service commits:', error);
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
          <span className="ml-2 text-lg text-gray-600">Loading commit history...</span>
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
          <div className="p-3 bg-green-100 rounded-lg mr-4">
            <GitCommit className="h-8 w-8 text-green-600" />
          </div>
          <div>
            <h1 className="text-3xl font-bold text-gray-900">Commit History</h1>
            <p className="mt-1 text-gray-600">
              Recent commits for <strong>{service.name}</strong> service
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

      {/* Commits List */}
      <div className="space-y-4">
        {filteredCommits.map((commit) => (
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
                    <div className="flex items-center space-x-4 text-xs text-gray-500">
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
        ))}
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

export default ServiceCommits;