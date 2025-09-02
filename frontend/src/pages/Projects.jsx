import React, { useState, useEffect } from 'react';
import { GetProjects, GetTasksByProject, CreateProject, UpdateProject, DeleteProject } from '../../wailsjs/go/main/App';
import { Plus, Edit, Trash2, Calendar, Clock } from 'lucide-react';
import ProjectModal from '../components/ProjectModal';
import TaskModal from '../components/TaskModal';

const Projects = () => {
  const [projects, setProjects] = useState([]);
  const [selectedProject, setSelectedProject] = useState(null);
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [showProjectModal, setShowProjectModal] = useState(false);
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [editingProject, setEditingProject] = useState(null);

  useEffect(() => {
    loadProjects();
  }, []);

  useEffect(() => {
    if (selectedProject) {
      loadTasks(selectedProject.id);
    }
  }, [selectedProject]);

  const loadProjects = async () => {
    try {
      const data = await GetProjects();
      setProjects(data || []);
      if (data && data.length > 0) {
        setSelectedProject(data[0]);
      }
    } catch (err) {
      setError('Failed to load projects: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const loadTasks = async (projectId) => {
    try {
      const data = await GetTasksByProject(projectId);
      setTasks(data || []);
    } catch (err) {
      console.error('Failed to load tasks:', err);
      setTasks([]);
    }
  };

  const handleCreateProject = async (projectData) => {
    try {
      await CreateProject(projectData);
      await loadProjects();
      setShowProjectModal(false);
    } catch (err) {
      setError('Failed to create project: ' + err.message);
    }
  };

  const handleUpdateProject = async (projectData) => {
    try {
      await UpdateProject(projectData);
      await loadProjects();
      setShowProjectModal(false);
      setEditingProject(null);
    } catch (err) {
      setError('Failed to update project: ' + err.message);
    }
  };

  const handleDeleteProject = async (projectId) => {
    if (!window.confirm('Are you sure you want to delete this project? This will also delete all associated tasks.')) {
      return;
    }
    
    try {
      await DeleteProject(projectId);
      await loadProjects();
      if (selectedProject && selectedProject.id === projectId) {
        setSelectedProject(null);
        setTasks([]);
      }
    } catch (err) {
      setError('Failed to delete project: ' + err.message);
    }
  };

  const handleEditProject = (project) => {
    setEditingProject(project);
    setShowProjectModal(true);
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800';
      case 'in_progress': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return null;
    return new Date(dateString).toLocaleDateString();
  };

  const isOverdue = (task) => {
    if (!task.deadline) return false;
    const deadline = new Date(task.deadline);
    const today = new Date();
    return deadline < today && deadline.toDateString() !== today.toDateString() && task.status !== 'completed';
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
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Project Planner</h1>
          <p className="text-gray-600">Manage your projects and track tasks with deadlines</p>
        </div>
        <button
          onClick={() => {
            setEditingProject(null);
            setShowProjectModal(true);
          }}
          className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center gap-2"
        >
          <Plus className="w-4 h-4" />
          New Project
        </button>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="grid grid-cols-12 gap-6">
        {/* Projects List */}
        <div className="col-span-4">
          <div className="bg-white rounded-lg shadow">
            <div className="p-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">Projects</h2>
            </div>
            <div className="divide-y divide-gray-200">
              {projects.length === 0 ? (
                <div className="p-6 text-center text-gray-500">
                  <p>No projects yet.</p>
                  <p className="text-sm mt-1">Create your first project to get started.</p>
                </div>
              ) : (
                projects.map(project => (
                  <div
                    key={project.id}
                    className={`p-4 cursor-pointer hover:bg-gray-50 ${
                      selectedProject?.id === project.id ? 'bg-blue-50 border-r-4 border-blue-600' : ''
                    }`}
                    onClick={() => setSelectedProject(project)}
                  >
                    <div className="flex justify-between items-start">
                      <div className="flex-1">
                        <h3 className="font-medium text-gray-900">{project.name}</h3>
                        {project.description && (
                          <p className="text-sm text-gray-600 mt-1">{project.description}</p>
                        )}
                        <p className="text-xs text-gray-500 mt-2">
                          Created {formatDate(project.created_at)}
                        </p>
                      </div>
                      <div className="flex items-center gap-1 ml-2">
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleEditProject(project);
                          }}
                          className="text-gray-400 hover:text-blue-600"
                        >
                          <Edit className="w-4 h-4" />
                        </button>
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleDeleteProject(project.id);
                          }}
                          className="text-gray-400 hover:text-red-600"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                ))
              )}
            </div>
          </div>
        </div>

        {/* Tasks List */}
        <div className="col-span-8">
          {selectedProject ? (
            <div className="bg-white rounded-lg shadow">
              <div className="p-4 border-b border-gray-200 flex justify-between items-center">
                <div>
                  <h2 className="text-lg font-semibold text-gray-900">{selectedProject.name} - Tasks</h2>
                  <p className="text-sm text-gray-600">{tasks.length} task(s)</p>
                </div>
                <button
                  onClick={() => setShowTaskModal(true)}
                  className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center gap-2"
                >
                  <Plus className="w-4 h-4" />
                  Add Task
                </button>
              </div>
              <div className="divide-y divide-gray-200">
                {tasks.length === 0 ? (
                  <div className="p-6 text-center text-gray-500">
                    <Calendar className="w-12 h-12 mx-auto mb-4 text-gray-300" />
                    <p>No tasks in this project.</p>
                    <p className="text-sm mt-1">Add your first task to start tracking work.</p>
                  </div>
                ) : (
                  tasks.map(task => (
                    <div key={task.id} className="p-4">
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center gap-3 mb-2">
                            <h3 className="font-medium text-gray-900">{task.title}</h3>
                            <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(task.status)}`}>
                              {task.status.replace('_', ' ')}
                            </span>
                          </div>
                          <div className="flex items-center gap-4 text-sm text-gray-600">
                            <span>Jira: {task.jira_ticket_id}</span>
                            {task.scheduled_date && (
                              <span className="flex items-center gap-1">
                                <Calendar className="w-4 h-4" />
                                Scheduled: {formatDate(task.scheduled_date)}
                              </span>
                            )}
                            {task.deadline && (
                              <span className={`flex items-center gap-1 ${
                                isOverdue(task) ? 'text-red-600 font-medium' : ''
                              }`}>
                                <Clock className="w-4 h-4" />
                                Due: {formatDate(task.deadline)}
                                {isOverdue(task) && <span className="text-red-600 ml-1">(Overdue)</span>}
                              </span>
                            )}
                          </div>
                          {task.description && (
                            <p className="text-sm text-gray-700 mt-2">{task.description}</p>
                          )}
                        </div>
                      </div>
                    </div>
                  ))
                )}
              </div>
            </div>
          ) : (
            <div className="bg-white rounded-lg shadow p-6 text-center text-gray-500">
              <h2 className="text-lg font-semibold mb-2">Select a Project</h2>
              <p>Choose a project from the left to view and manage its tasks.</p>
            </div>
          )}
        </div>
      </div>

      {/* Modals */}
      {showProjectModal && (
        <ProjectModal
          project={editingProject}
          onClose={() => {
            setShowProjectModal(false);
            setEditingProject(null);
          }}
          onSubmit={editingProject ? handleUpdateProject : handleCreateProject}
        />
      )}

      {showTaskModal && selectedProject && (
        <TaskModal
          project={selectedProject}
          onClose={() => setShowTaskModal(false)}
          onTaskCreated={() => {
            loadTasks(selectedProject.id);
            setShowTaskModal(false);
          }}
        />
      )}
    </div>
  );
};

export default Projects;