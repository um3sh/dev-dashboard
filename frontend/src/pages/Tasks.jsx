import React, { useState, useEffect } from 'react';
import { GetTasksGroupedByScheduledDate, UpdateTaskStatus } from '../../wailsjs/go/main/App';
import { Copy, CheckCircle, Clock, AlertCircle, Calendar, ExternalLink } from 'lucide-react';

const Tasks = () => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [copiedTicketId, setCopiedTicketId] = useState(null);

  useEffect(() => {
    loadTasks();
  }, []);

  const loadTasks = async () => {
    setLoading(true);
    try {
      const data = await GetTasksGroupedByScheduledDate();
      setTasks(data || []);
    } catch (err) {
      setError('Failed to load tasks: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const handleStatusChange = async (taskId, newStatus) => {
    try {
      await UpdateTaskStatus(taskId, newStatus);
      await loadTasks();
    } catch (err) {
      setError('Failed to update task status: ' + err.message);
    }
  };

  const handleCopyTicketId = async (ticketId) => {
    try {
      await navigator.clipboard.writeText(ticketId);
      setCopiedTicketId(ticketId);
      setTimeout(() => setCopiedTicketId(null), 2000);
    } catch (err) {
      console.error('Failed to copy ticket ID:', err);
      // Fallback for environments where clipboard API isn't available
      const textArea = document.createElement('textarea');
      textArea.value = ticketId;
      document.body.appendChild(textArea);
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
      setCopiedTicketId(ticketId);
      setTimeout(() => setCopiedTicketId(null), 2000);
    }
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800';
      case 'in_progress': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusIcon = (status) => {
    switch (status) {
      case 'completed': return <CheckCircle className="w-4 h-4" />;
      case 'in_progress': return <Clock className="w-4 h-4" />;
      default: return <AlertCircle className="w-4 h-4" />;
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return null;
    return new Date(dateString).toLocaleDateString();
  };

  const isOverdue = (task) => {
    if (!task.deadline) return false;
    const today = new Date();
    const deadline = new Date(task.deadline);
    return deadline < today && deadline.toDateString() !== today.toDateString() && task.status !== 'completed';
  };

  const isToday = (dateString) => {
    if (!dateString) return false;
    const date = new Date(dateString);
    const today = new Date();
    return date.toDateString() === today.toDateString();
  };

  const isPast = (dateString) => {
    if (!dateString) return false;
    const date = new Date(dateString);
    const today = new Date();
    return date < today && date.toDateString() !== today.toDateString();
  };

  const groupTasksByDate = (tasks) => {
    const groups = {};
    
    tasks.forEach(task => {
      let dateKey;
      let dateLabel;
      
      if (!task.scheduled_date) {
        dateKey = 'unscheduled';
        dateLabel = 'Unscheduled';
      } else {
        const date = new Date(task.scheduled_date);
        dateKey = date.toDateString();
        
        if (isToday(task.scheduled_date)) {
          dateLabel = `Today - ${date.toLocaleDateString()}`;
        } else if (isPast(task.scheduled_date)) {
          dateLabel = `${date.toLocaleDateString()} (Past)`;
        } else {
          dateLabel = date.toLocaleDateString();
        }
      }
      
      if (!groups[dateKey]) {
        groups[dateKey] = {
          label: dateLabel,
          tasks: [],
          isToday: task.scheduled_date && isToday(task.scheduled_date),
          isPast: task.scheduled_date && isPast(task.scheduled_date),
          isUnscheduled: !task.scheduled_date
        };
      }
      
      groups[dateKey].tasks.push(task);
    });
    
    return groups;
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center min-h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  const groupedTasks = groupTasksByDate(tasks);
  const sortedGroups = Object.entries(groupedTasks).sort(([keyA, groupA], [keyB, groupB]) => {
    if (groupA.isUnscheduled && !groupB.isUnscheduled) return 1;
    if (!groupA.isUnscheduled && groupB.isUnscheduled) return -1;
    if (keyA === keyB) return 0;
    return new Date(keyA) - new Date(keyB);
  });

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Tasks</h1>
          <p className="text-gray-600">Your task list organized by scheduled dates</p>
        </div>
        <div className="text-sm text-gray-500">
          Total: {tasks.length} task{tasks.length !== 1 ? 's' : ''}
        </div>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="space-y-6">
        {sortedGroups.length === 0 ? (
          <div className="bg-white rounded-lg shadow p-8 text-center">
            <Calendar className="w-12 h-12 mx-auto mb-4 text-gray-300" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">No Tasks Yet</h3>
            <p className="text-gray-600">
              Create your first task from the Projects page to get started.
            </p>
          </div>
        ) : (
          sortedGroups.map(([dateKey, group]) => (
            <div key={dateKey} className="bg-white rounded-lg shadow">
              <div className={`px-6 py-4 border-b border-gray-200 ${
                group.isToday ? 'bg-blue-50' : group.isPast ? 'bg-red-50' : group.isUnscheduled ? 'bg-gray-50' : ''
              }`}>
                <div className="flex items-center justify-between">
                  <h2 className={`text-lg font-semibold ${
                    group.isToday ? 'text-blue-900' : group.isPast ? 'text-red-900' : 'text-gray-900'
                  }`}>
                    {group.label}
                  </h2>
                  <span className="text-sm text-gray-600">
                    {group.tasks.length} task{group.tasks.length !== 1 ? 's' : ''}
                  </span>
                </div>
              </div>

              <div className="divide-y divide-gray-200">
                {group.tasks.map(task => (
                  <div key={task.id} className="p-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-3">
                          <button
                            onClick={() => handleCopyTicketId(task.jira_ticket_id)}
                            className="flex items-center gap-2 text-blue-600 hover:text-blue-800 font-medium"
                            title="Click to copy ticket ID"
                          >
                            <ExternalLink className="w-4 h-4" />
                            {task.jira_ticket_id}
                            {copiedTicketId === task.jira_ticket_id ? (
                              <span className="text-green-600 text-sm">Copied!</span>
                            ) : (
                              <Copy className="w-4 h-4" />
                            )}
                          </button>
                          
                          <span className={`px-3 py-1 text-sm font-medium rounded-full flex items-center gap-1 ${getStatusColor(task.status)}`}>
                            {getStatusIcon(task.status)}
                            {task.status.replace('_', ' ')}
                          </span>
                        </div>

                        <h3 className="font-medium text-gray-900 mb-2">{task.title}</h3>
                        
                        <div className="text-sm text-gray-600 space-y-1">
                          <p>Project: <span className="font-medium">{task.project_name}</span></p>
                          {task.description && (
                            <p className="text-gray-700">{task.description}</p>
                          )}
                        </div>

                        <div className="flex items-center gap-4 mt-3 text-sm">
                          {task.scheduled_date && (
                            <span className="flex items-center gap-1 text-gray-600">
                              <Calendar className="w-4 h-4" />
                              Scheduled: {formatDate(task.scheduled_date)}
                            </span>
                          )}
                          {task.deadline && (
                            <span className={`flex items-center gap-1 ${
                              isOverdue(task) ? 'text-red-600 font-medium' : 'text-gray-600'
                            }`}>
                              <Clock className="w-4 h-4" />
                              Due: {formatDate(task.deadline)}
                              {isOverdue(task) && <span className="ml-1">(Overdue)</span>}
                            </span>
                          )}
                        </div>
                      </div>

                      <div className="flex items-center gap-2 ml-4">
                        {task.status === 'pending' && (
                          <button
                            onClick={() => handleStatusChange(task.id, 'in_progress')}
                            className="text-blue-600 hover:text-blue-800 text-sm px-3 py-1 border border-blue-600 rounded-md"
                          >
                            Start
                          </button>
                        )}
                        {task.status !== 'completed' && (
                          <button
                            onClick={() => handleStatusChange(task.id, 'completed')}
                            className="text-green-600 hover:text-green-800 text-sm px-3 py-1 border border-green-600 rounded-md flex items-center gap-1"
                          >
                            <CheckCircle className="w-4 h-4" />
                            Complete
                          </button>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default Tasks;