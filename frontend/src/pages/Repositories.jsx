import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { 
  Plus, 
  Database, 
  Server, 
  ExternalLink,
  Clock,
  Settings,
  Trash2,
  RefreshCw
} from 'lucide-react';
import RepositoryModal from '../components/RepositoryModal';

const Repositories = () => {
  const [repositories, setRepositories] = useState([]);
  const [showAddModal, setShowAddModal] = useState(false);

  // Load repositories from backend
  useEffect(() => {
    loadRepositories();
  }, []);

  const loadRepositories = async () => {
    try {
      const repos = await window.go.main.App.GetRepositories();
      setRepositories(repos || []);
    } catch (error) {
      console.error('Failed to load repositories:', error);
    }
  };

  const handleRepositoryCreated = async () => {
    await loadRepositories(); // Refresh the list
    setShowAddModal(false);
  };

  const handleRediscoverServices = async (repo) => {
    if (repo.type !== 'monorepo') {
      alert('Service discovery is only available for monorepo repositories.');
      return;
    }

    try {
      // Check if GitHub token is configured
      const config = await window.go.main.App.GetAllConfig();
      if (!config.github_token || !config.github_token.trim()) {
        alert('GitHub token not configured. Please go to Settings and configure your GitHub Personal Access Token first.');
        return;
      }

      // Use globally configured token (empty credentials object for PAT)
      await window.go.main.App.RediscoverRepositoryServices(repo.id, 'pat', {});
      
      alert(`Service discovery completed for ${repo.name}.`);
      await loadRepositories(); // Refresh the list
    } catch (error) {
      console.error('Failed to rediscover services:', error);
      if (error.message.includes('GitHub') || error.message.includes('token')) {
        alert(`Failed to rediscover services: ${error.message}. Please check your GitHub token configuration in Settings.`);
      } else {
        alert(`Failed to rediscover services: ${error.message}`);
      }
    }
  };

  const handleDeleteRepository = async (id) => {
    if (window.confirm('Are you sure you want to delete this repository?')) {
      try {
        await window.go.main.App.DeleteRepository(id);
        await loadRepositories(); // Refresh the list
      } catch (error) {
        console.error('Failed to delete repository:', error);
        alert('Failed to delete repository: ' + error);
      }
    }
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const getTypeIcon = (type) => {
    return type === 'monorepo' ? <Database className="h-5 w-5" /> : <Server className="h-5 w-5" />;
  };

  const getTypeColor = (type) => {
    return type === 'monorepo' ? 'bg-blue-100 text-blue-800' : 'bg-purple-100 text-purple-800';
  };

  return (
    <div className="max-w-7xl mx-auto">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold text-gray-900">Repositories</h1>
          <p className="mt-2 text-gray-600">
            Manage your monorepo and Kubernetes resource repositories
          </p>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="btn-primary flex items-center"
        >
          <Plus className="h-5 w-5 mr-2" />
          Add Repository
        </button>
      </div>

      {/* Add Repository Modal */}
      {showAddModal && (
        <RepositoryModal
          onClose={() => setShowAddModal(false)}
          onRepositoryCreated={handleRepositoryCreated}
        />
      )}

      {/* Repositories List */}
      <div className="grid gap-6">
        {repositories.map((repo) => (
          <div key={repo.id} className="card">
            <div className="flex items-start justify-between">
              <div className="flex-1">
                <div className="flex items-center space-x-3 mb-2">
                  <div className={`p-2 rounded-lg ${getTypeColor(repo.type)}`}>
                    {getTypeIcon(repo.type)}
                  </div>
                  <div>
                    <h3 className="text-lg font-semibold text-gray-900">{repo.name}</h3>
                    <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${getTypeColor(repo.type)}`}>
                      {repo.type === 'monorepo' ? 'Monorepo' : 'Kubernetes'}
                    </span>
                  </div>
                </div>
                
                <p className="text-gray-600 mb-4">{repo.description}</p>
                
                <div className="flex items-center space-x-6 text-sm text-gray-500">
                  <div className="flex items-center">
                    <ExternalLink className="h-4 w-4 mr-1" />
                    <a href={repo.url} target="_blank" rel="noopener noreferrer" className="hover:text-blue-600">
                      {repo.url}
                    </a>
                  </div>
                  <div className="flex items-center">
                    <Clock className="h-4 w-4 mr-1" />
                    Last sync: {repo.lastSyncAt ? formatDate(repo.lastSyncAt) : 'Never'}
                  </div>
                  {repo.servicesCount && (
                    <div>
                      {repo.servicesCount} services
                    </div>
                  )}
                  {repo.resourcesCount && (
                    <div>
                      {repo.resourcesCount} resources
                    </div>
                  )}
                </div>
              </div>
              
              <div className="flex items-center space-x-2">
                <button 
                  onClick={() => handleRediscoverServices(repo)}
                  className="p-2 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100"
                  title="Rediscover Services"
                >
                  <RefreshCw className="h-5 w-5" />
                </button>
                <Link
                  to={repo.type === 'monorepo' ? `/microservices/${repo.id}` : `/kubernetes/${repo.id}`}
                  className="btn-primary"
                >
                  View {repo.type === 'monorepo' ? 'Services' : 'Resources'}
                </Link>
                <button 
                  onClick={() => handleDeleteRepository(repo.id)}
                  className="p-2 text-gray-400 hover:text-red-600 rounded-md hover:bg-gray-100"
                >
                  <Trash2 className="h-5 w-5" />
                </button>
              </div>
            </div>
          </div>
        ))}
        
        {repositories.length === 0 && (
          <div className="text-center py-12">
            <Database className="mx-auto h-12 w-12 text-gray-400" />
            <h3 className="mt-2 text-sm font-medium text-gray-900">No repositories</h3>
            <p className="mt-1 text-sm text-gray-500">Get started by adding a new repository.</p>
            <button
              onClick={() => setShowAddModal(true)}
              className="mt-6 btn-primary"
            >
              <Plus className="h-5 w-5 mr-2" />
              Add Repository
            </button>
          </div>
        )}
      </div>
    </div>
  );
};

export default Repositories;