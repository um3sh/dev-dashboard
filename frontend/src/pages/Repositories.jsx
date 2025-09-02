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

const Repositories = () => {
  const [repositories, setRepositories] = useState([]);
  const [showAddForm, setShowAddForm] = useState(false);
  const [newRepo, setNewRepo] = useState({
    name: '',
    url: '',
    type: 'monorepo',
    description: '',
    serviceName: '',
    serviceLocation: ''
  });

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

  const handleAddRepository = async (e) => {
    e.preventDefault();
    try {
      await window.go.main.App.CreateRepository({
        name: newRepo.name,
        url: newRepo.url,
        type: newRepo.type,
        description: newRepo.description,
        serviceName: newRepo.serviceName,
        serviceLocation: newRepo.serviceLocation
      });
      
      await loadRepositories(); // Refresh the list
      setNewRepo({ name: '', url: '', type: 'monorepo', description: '', serviceName: '', serviceLocation: '' });
      setShowAddForm(false);
    } catch (error) {
      console.error('Failed to add repository:', error);
      alert('Failed to add repository: ' + error);
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
          onClick={() => setShowAddForm(true)}
          className="btn-primary flex items-center"
        >
          <Plus className="h-5 w-5 mr-2" />
          Add Repository
        </button>
      </div>

      {/* Add Repository Form */}
      {showAddForm && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-md w-full mx-4">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Add Repository</h3>
            <form onSubmit={handleAddRepository} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">Name</label>
                <input
                  type="text"
                  value={newRepo.name}
                  onChange={(e) => setNewRepo({...newRepo, name: e.target.value})}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">URL</label>
                <input
                  type="url"
                  value={newRepo.url}
                  onChange={(e) => setNewRepo({...newRepo, url: e.target.value})}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  placeholder="https://github.com/owner/repo"
                  required
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Type</label>
                <select
                  value={newRepo.type}
                  onChange={(e) => setNewRepo({...newRepo, type: e.target.value})}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                >
                  <option value="monorepo">Monorepo</option>
                  <option value="kubernetes">Kubernetes</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700">Description</label>
                <textarea
                  value={newRepo.description}
                  onChange={(e) => setNewRepo({...newRepo, description: e.target.value})}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                  rows={3}
                />
              </div>
              {newRepo.type === 'monorepo' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Service Name</label>
                    <input
                      type="text"
                      value={newRepo.serviceName}
                      onChange={(e) => setNewRepo({...newRepo, serviceName: e.target.value})}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                      placeholder="e.g., user-service, api-gateway"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Service Location</label>
                    <input
                      type="text"
                      value={newRepo.serviceLocation}
                      onChange={(e) => setNewRepo({...newRepo, serviceLocation: e.target.value})}
                      className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-blue-500 focus:ring-blue-500"
                      placeholder="e.g., services/, apps/backend/"
                    />
                  </div>
                </>
              )}
              <div className="flex justify-end space-x-3 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAddForm(false)}
                  className="btn-secondary"
                >
                  Cancel
                </button>
                <button type="submit" className="btn-primary">
                  Add Repository
                </button>
              </div>
            </form>
          </div>
        </div>
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
                <button className="p-2 text-gray-400 hover:text-blue-600 rounded-md hover:bg-gray-100">
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
              onClick={() => setShowAddForm(true)}
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