import React, { useState, useEffect } from 'react';
import { X, RefreshCw, Check, AlertCircle, Github, Settings } from 'lucide-react';

const RepositoryModal = ({ onClose, onRepositoryCreated }) => {
  const [formData, setFormData] = useState({
    name: '',
    url: '',
    type: 'monorepo',
    description: '',
    serviceLocation: 'services/'
  });
  const [authMethod, setAuthMethod] = useState('pat'); // Only 'pat' supported
  const [githubTokenConfigured, setGithubTokenConfigured] = useState(false);
  const [isValidating, setIsValidating] = useState(false);
  const [validationStatus, setValidationStatus] = useState(''); // 'success', 'error', or ''
  const [validationMessage, setValidationMessage] = useState('');
  const [discoveredServices, setDiscoveredServices] = useState([]);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [errors, setErrors] = useState({});

  // Check if GitHub token is configured on component mount
  useEffect(() => {
    checkGithubTokenConfiguration();
  }, []);

  // Auto-validate repository when URL and GitHub token are configured
  useEffect(() => {
    const timer = setTimeout(() => {
      if (formData.url && githubTokenConfigured) {
        validateRepository();
      }
    }, 1000); // Debounce for 1 second

    return () => clearTimeout(timer);
  }, [formData.url, githubTokenConfigured]);

  const checkGithubTokenConfiguration = async () => {
    try {
      const config = await window.go.main.App.GetAllConfig();
      setGithubTokenConfigured(!!(config.github_token && config.github_token.trim()));
    } catch (err) {
      console.error('Failed to check GitHub token configuration:', err);
      setGithubTokenConfigured(false);
    }
  };

  const validateRepository = async () => {
    if (!formData.url) return;

    setIsValidating(true);
    setValidationStatus('');
    setValidationMessage('');
    setDiscoveredServices([]);
    
    try {
      // Use empty credentials object since GitHub token is configured globally
      const validationCredentials = {};
      const result = await window.go.main.App.ValidateRepositoryAccess(formData.url, authMethod, validationCredentials);
      
      if (result.success) {
        setValidationStatus('success');
        setValidationMessage('Repository access validated successfully');
        
        // If it's a monorepo, discover services
        if (formData.type === 'monorepo' && formData.serviceLocation) {
          await discoverServices();
        }
      } else {
        setValidationStatus('error');
        setValidationMessage(result.error || 'Failed to access repository');
      }
    } catch (err) {
      console.error('Repository validation failed:', err);
      setValidationStatus('error');
      setValidationMessage(err.message || 'Failed to validate repository access');
    } finally {
      setIsValidating(false);
    }
  };

  const discoverServices = async () => {
    if (!formData.url || formData.type !== 'monorepo') return;

    try {
      const discoveryCredentials = {}; // Use globally configured GitHub token
      const services = await window.go.main.App.DiscoverRepositoryServices(
        formData.url, 
        formData.serviceLocation,
        authMethod, 
        discoveryCredentials
      );
      
      setDiscoveredServices(services || []);
      
      if (services && services.length > 0) {
        setValidationMessage(`Repository validated. Discovered ${services.length} service(s)`);
      }
    } catch (err) {
      console.error('Service discovery failed:', err);
      setValidationMessage('Repository validated but service discovery failed: ' + err.message);
    }
  };

  const validateForm = () => {
    const newErrors = {};
    
    if (!formData.name.trim()) {
      newErrors.name = 'Repository name is required';
    }

    if (!formData.url.trim()) {
      newErrors.url = 'Repository URL is required';
    } else if (!formData.url.match(/^https:\/\/.+/)) {
      newErrors.url = 'Please enter a valid HTTPS URL';
    } else if (!formData.url.match(/^https:\/\/[^\/]+\/[^\/]+\/[^\/]+/)) {
      newErrors.url = 'Please enter a valid GitHub repository URL (e.g., https://github.com/owner/repo or https://github.company.com/owner/repo)';
    }

    if (!githubTokenConfigured) {
      newErrors.githubToken = 'GitHub token must be configured in Settings first';
    }

    if (formData.type === 'monorepo' && !formData.serviceLocation.trim()) {
      newErrors.serviceLocation = 'Service location is required for monorepos';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    if (validationStatus !== 'success') {
      setErrors({ submit: 'Please validate repository access before submitting' });
      return;
    }

    setIsSubmitting(true);
    
    try {
      const repoData = {
        name: formData.name.trim(),
        url: formData.url.trim(),
        type: formData.type,
        description: formData.description.trim(),
        service_location: formData.type === 'monorepo' ? formData.serviceLocation.trim() : '',
        auth_method: authMethod,
        credentials: {}
      };

      await window.go.main.App.CreateRepositoryWithAuth(repoData);
      onRepositoryCreated();
    } catch (err) {
      setErrors({ submit: err.message || 'Failed to create repository' });
      setIsSubmitting(false);
    }
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
    
    // Clear error when user starts typing
    if (errors[name]) {
      setErrors(prev => ({
        ...prev,
        [name]: ''
      }));
    }

    // Reset validation when URL or type changes
    if (name === 'url' || name === 'type') {
      setValidationStatus('');
      setValidationMessage('');
      setDiscoveredServices([]);
    }
  };


  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl mx-4 max-h-screen overflow-y-auto">
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Add Repository</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6">
          <div className="space-y-6">
            {/* Basic Repository Information */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900">Repository Information</h3>
              
              <div>
                <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                  Repository Name *
                </label>
                <input
                  type="text"
                  id="name"
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  className={`w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                    errors.name ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="e.g., my-monorepo, k8s-configs"
                  disabled={isSubmitting}
                />
                {errors.name && (
                  <p className="text-red-500 text-sm mt-1">{errors.name}</p>
                )}
              </div>

              <div>
                <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-1">
                  Repository URL *
                </label>
                <input
                  type="url"
                  id="url"
                  name="url"
                  value={formData.url}
                  onChange={handleInputChange}
                  className={`w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                    errors.url ? 'border-red-500' : 'border-gray-300'
                  }`}
                  placeholder="https://github.com/owner/repository or https://github.company.com/owner/repository"
                  disabled={isSubmitting}
                />
                {errors.url && (
                  <p className="text-red-500 text-sm mt-1">{errors.url}</p>
                )}
              </div>

              <div>
                <label htmlFor="type" className="block text-sm font-medium text-gray-700 mb-1">
                  Repository Type
                </label>
                <select
                  id="type"
                  name="type"
                  value={formData.type}
                  onChange={handleInputChange}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  disabled={isSubmitting}
                >
                  <option value="monorepo">Monorepo (contains multiple services)</option>
                  <option value="kubernetes">Kubernetes Resources</option>
                </select>
              </div>

              <div>
                <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-1">
                  Description (Optional)
                </label>
                <textarea
                  id="description"
                  name="description"
                  value={formData.description}
                  onChange={handleInputChange}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={3}
                  placeholder="Brief description of this repository..."
                  disabled={isSubmitting}
                />
              </div>

              {formData.type === 'monorepo' && (
                <div>
                  <label htmlFor="serviceLocation" className="block text-sm font-medium text-gray-700 mb-1">
                    Service Location *
                  </label>
                  <input
                    type="text"
                    id="serviceLocation"
                    name="serviceLocation"
                    value={formData.serviceLocation}
                    onChange={handleInputChange}
                    className={`w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                      errors.serviceLocation ? 'border-red-500' : 'border-gray-300'
                    }`}
                    placeholder="e.g., services/, apps/, microservices/"
                    disabled={isSubmitting}
                  />
                  {errors.serviceLocation && (
                    <p className="text-red-500 text-sm mt-1">{errors.serviceLocation}</p>
                  )}
                  <p className="text-xs text-gray-500 mt-1">
                    Directory path where services are located (relative to repository root)
                  </p>
                </div>
              )}
            </div>

            {/* Authentication Method */}
            <div className="space-y-4">
              <h3 className="text-lg font-medium text-gray-900">Authentication</h3>
              
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                <div className="flex items-center gap-2 mb-3">
                  <Github className="w-5 h-5 text-gray-700" />
                  <span className="text-sm font-medium text-gray-900">GitHub Personal Access Token</span>
                </div>
                
                {githubTokenConfigured ? (
                  <div className="bg-green-50 border border-green-200 rounded-lg p-3 flex items-center gap-2">
                    <Check className="w-4 h-4 text-green-600" />
                    <span className="text-sm text-green-800">
                      GitHub token is configured globally in Settings
                    </span>
                  </div>
                ) : (
                  <div className="bg-amber-50 border border-amber-200 rounded-lg p-3">
                    <div className="flex items-center gap-2 mb-2">
                      <AlertCircle className="w-4 h-4 text-amber-600" />
                      <span className="text-sm font-medium text-amber-800">
                        GitHub token not configured
                      </span>
                    </div>
                    <p className="text-sm text-amber-700 mb-3">
                      Please configure your GitHub Personal Access Token in Settings first.
                    </p>
                    <button
                      type="button"
                      onClick={() => {
                        // Open settings in new tab/window or navigate
                        window.location.hash = '#/settings';
                      }}
                      className="flex items-center gap-1 text-xs bg-amber-100 text-amber-800 px-2 py-1 rounded hover:bg-amber-200"
                    >
                      <Settings className="w-3 h-3" />
                      Go to Settings
                    </button>
                  </div>
                )}
                
                {errors.githubToken && (
                  <p className="text-red-500 text-sm mt-1">{errors.githubToken}</p>
                )}
              </div>
            </div>

            {/* Repository Validation Status */}
            {(isValidating || validationStatus || validationMessage) && (
              <div className="space-y-2">
                <div className={`p-3 rounded-md flex items-center gap-2 ${
                  validationStatus === 'success' ? 'bg-green-50 border border-green-200' :
                  validationStatus === 'error' ? 'bg-red-50 border border-red-200' :
                  'bg-blue-50 border border-blue-200'
                }`}>
                  {isValidating ? (
                    <RefreshCw className="w-4 h-4 animate-spin text-blue-600" />
                  ) : validationStatus === 'success' ? (
                    <Check className="w-4 h-4 text-green-600" />
                  ) : validationStatus === 'error' ? (
                    <AlertCircle className="w-4 h-4 text-red-600" />
                  ) : null}
                  
                  <span className={`text-sm ${
                    validationStatus === 'success' ? 'text-green-800' :
                    validationStatus === 'error' ? 'text-red-800' :
                    'text-blue-800'
                  }`}>
                    {isValidating ? 'Validating repository access...' : validationMessage}
                  </span>
                </div>

                {/* Discovered Services */}
                {discoveredServices.length > 0 && (
                  <div className="bg-gray-50 border border-gray-200 rounded-md p-3">
                    <h4 className="text-sm font-medium text-gray-900 mb-2">
                      Discovered Services ({discoveredServices.length})
                    </h4>
                    <div className="space-y-1">
                      {discoveredServices.map((service, index) => (
                        <div key={index} className="text-xs text-gray-600 flex items-center gap-2">
                          <span className="w-2 h-2 bg-green-500 rounded-full"></span>
                          <span className="font-medium">{service.name}</span>
                          <span className="text-gray-400">({service.path})</span>
                        </div>
                      ))}
                    </div>
                  </div>
                )}
              </div>
            )}

            {errors.submit && (
              <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                {errors.submit}
              </div>
            )}

            <div className="flex justify-end gap-3 pt-4 border-t border-gray-200">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
                disabled={isSubmitting}
              >
                Cancel
              </button>
              <button
                type="submit"
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                disabled={isSubmitting || validationStatus !== 'success'}
              >
                {isSubmitting ? 'Creating...' : 'Create Repository'}
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default RepositoryModal;