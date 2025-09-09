import React from 'react';
import { BrowserRouter as Router, Routes, Route } from 'react-router-dom';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Repositories from './pages/Repositories';
import Microservices from './pages/Microservices';
import ServiceDetails from './pages/ServiceDetails';
import ServicePullRequests from './pages/ServicePullRequests';
import ServiceCommits from './pages/ServiceCommits';
import ServiceDeployments from './pages/ServiceDeployments';
import ServiceDeploymentHistory from './pages/ServiceDeploymentHistory';
import KubernetesResources from './pages/KubernetesResources';
import Projects from './pages/Projects';
import Tasks from './pages/Tasks';
import Calendar from './pages/Calendar';
import Settings from './pages/Settings';

function App() {
  return (
    <Router>
      <Layout>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/repositories" element={<Repositories />} />
          <Route path="/microservices/:repoId?" element={<Microservices />} />
          <Route path="/service/:serviceId" element={<ServiceDetails />} />
          <Route path="/service/:serviceId/pull-requests" element={<ServicePullRequests />} />
          <Route path="/service/:serviceId/commits" element={<ServiceCommits />} />
          <Route path="/service/:serviceId/deployments" element={<ServiceDeployments />} />
          <Route path="/service/:serviceId/deployment-history" element={<ServiceDeploymentHistory />} />
          <Route path="/kubernetes/:repoId?" element={<KubernetesResources />} />
          <Route path="/projects" element={<Projects />} />
          <Route path="/tasks" element={<Tasks />} />
          <Route path="/calendar" element={<Calendar />} />
          <Route path="/settings" element={<Settings />} />
        </Routes>
      </Layout>
    </Router>
  );
}

export default App;