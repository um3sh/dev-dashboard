import React, { useState, useEffect } from 'react';
import { GetAllConfig, SetConfig, TestJiraConnection, RefreshAllJiraTitles } from '../../wailsjs/go/main/App';
import { Save, TestTube, RefreshCw, CheckCircle, XCircle, Settings as SettingsIcon } from 'lucide-react';

const Settings = () => {
  const [config, setConfig] = useState({
    jira_url: '',
    jira_token: ''
  });
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [refreshing, setRefreshing] = useState(false);
  const [message, setMessage] = useState('');
  const [messageType, setMessageType] = useState(''); // 'success', 'error', or ''

  useEffect(() => {
    loadConfig();
  }, []);

  const loadConfig = async () => {
    setLoading(true);
    try {
      const configData = await GetAllConfig();
      setConfig({
        jira_url: configData.jira_url || '',
        jira_token: configData.jira_token || ''
      });
    } catch (err) {
      console.error('Failed to load config:', err);
      setMessage('Failed to load configuration');
      setMessageType('error');
    } finally {
      setLoading(false);
    }
  };

  const showMessage = (text, type) => {
    setMessage(text);
    setMessageType(type);
    setTimeout(() => {
      setMessage('');
      setMessageType('');
    }, 5000);
  };

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setConfig(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSave = async () => {
    setSaving(true);
    try {
      await SetConfig('jira_url', config.jira_url);
      await SetConfig('jira_token', config.jira_token);
      showMessage('Configuration saved successfully!', 'success');
    } catch (err) {
      console.error('Failed to save config:', err);
      showMessage('Failed to save configuration: ' + err.message, 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleTestConnection = async () => {
    if (!config.jira_url || !config.jira_token) {
      showMessage('Please enter both JIRA URL and token before testing', 'error');
      return;
    }

    // Save config first if changed
    await handleSave();

    setTesting(true);
    try {
      await TestJiraConnection();
      showMessage('JIRA connection test successful!', 'success');
    } catch (err) {
      console.error('JIRA connection test failed:', err);
      showMessage('JIRA connection test failed: ' + err.message, 'error');
    } finally {
      setTesting(false);
    }
  };

  const handleRefreshTitles = async () => {
    if (!config.jira_url || !config.jira_token) {
      showMessage('Please configure and test JIRA connection first', 'error');
      return;
    }

    setRefreshing(true);
    try {
      await RefreshAllJiraTitles();
      showMessage('Successfully refreshed all JIRA ticket titles!', 'success');
    } catch (err) {
      console.error('Failed to refresh JIRA titles:', err);
      showMessage('Failed to refresh JIRA titles: ' + err.message, 'error');
    } finally {
      setRefreshing(false);
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <SettingsIcon className="w-8 h-8 text-gray-700" />
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Settings</h1>
          <p className="text-gray-600">Configure your JIRA integration and other preferences</p>
        </div>
      </div>

      {message && (
        <div className={`p-4 rounded-lg border ${
          messageType === 'success' 
            ? 'bg-green-50 border-green-200 text-green-800' 
            : 'bg-red-50 border-red-200 text-red-800'
        }`}>
          <div className="flex items-center gap-2">
            {messageType === 'success' ? (
              <CheckCircle className="w-5 h-5" />
            ) : (
              <XCircle className="w-5 h-5" />
            )}
            {message}
          </div>
        </div>
      )}

      <div className="bg-white rounded-lg shadow-sm border border-gray-200">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-900">JIRA Integration</h2>
          <p className="text-sm text-gray-600 mt-1">
            Configure your JIRA connection to automatically fetch ticket titles
          </p>
        </div>

        <div className="p-6 space-y-6">
          <div>
            <label htmlFor="jira_url" className="block text-sm font-medium text-gray-700 mb-2">
              JIRA Base URL
            </label>
            <input
              type="url"
              id="jira_url"
              name="jira_url"
              value={config.jira_url}
              onChange={handleInputChange}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="https://your-domain.atlassian.net"
              disabled={saving}
            />
            <p className="text-xs text-gray-500 mt-1">
              Your JIRA instance URL (e.g., https://company.atlassian.net)
            </p>
          </div>

          <div>
            <label htmlFor="jira_token" className="block text-sm font-medium text-gray-700 mb-2">
              JIRA Personal Access Token
            </label>
            <input
              type="password"
              id="jira_token"
              name="jira_token"
              value={config.jira_token}
              onChange={handleInputChange}
              className="w-full border border-gray-300 rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Your JIRA API token"
              disabled={saving}
            />
            <p className="text-xs text-gray-500 mt-1">
              Create a token at: JIRA → Profile → Personal Access Tokens → Create token
            </p>
          </div>

          <div className="flex gap-3">
            <button
              onClick={handleSave}
              disabled={saving}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {saving ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <Save className="w-4 h-4" />
              )}
              {saving ? 'Saving...' : 'Save Configuration'}
            </button>

            <button
              onClick={handleTestConnection}
              disabled={testing || !config.jira_url || !config.jira_token}
              className="flex items-center gap-2 px-4 py-2 border border-green-600 text-green-600 rounded-lg hover:bg-green-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {testing ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <TestTube className="w-4 h-4" />
              )}
              {testing ? 'Testing...' : 'Test Connection'}
            </button>

            <button
              onClick={handleRefreshTitles}
              disabled={refreshing || !config.jira_url || !config.jira_token}
              className="flex items-center gap-2 px-4 py-2 border border-purple-600 text-purple-600 rounded-lg hover:bg-purple-50 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {refreshing ? (
                <RefreshCw className="w-4 h-4 animate-spin" />
              ) : (
                <RefreshCw className="w-4 h-4" />
              )}
              {refreshing ? 'Refreshing...' : 'Refresh All JIRA Titles'}
            </button>
          </div>

          <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
            <h3 className="font-medium text-blue-900 mb-2">How to set up JIRA integration:</h3>
            <ol className="text-sm text-blue-800 space-y-1 list-decimal list-inside">
              <li>Go to your JIRA instance (e.g., https://company.atlassian.net)</li>
              <li>Click your profile picture → Profile → Personal Access Tokens</li>
              <li>Create a new token with appropriate permissions (read access to issues)</li>
              <li>Copy the token and paste it above</li>
              <li>Enter your JIRA base URL and save the configuration</li>
              <li>Test the connection to verify everything works</li>
            </ol>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Settings;