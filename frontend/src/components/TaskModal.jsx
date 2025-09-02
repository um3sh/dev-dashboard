import React, { useState } from 'react';
import { CreateTask } from '../../wailsjs/go/main/App';
import { X } from 'lucide-react';

const TaskModal = ({ project, onClose, onTaskCreated }) => {
  const [formData, setFormData] = useState({
    jira_ticket_id: '',
    scheduled_date: '',
    deadline: ''
  });
  const [errors, setErrors] = useState({});
  const [isSubmitting, setIsSubmitting] = useState(false);

  const validateForm = () => {
    const newErrors = {};
    
    if (!formData.jira_ticket_id.trim()) {
      newErrors.jira_ticket_id = 'Jira ticket ID is required';
    } else if (formData.jira_ticket_id.trim().length < 2) {
      newErrors.jira_ticket_id = 'Jira ticket ID must be at least 2 characters';
    }

    if (!formData.scheduled_date) {
      newErrors.scheduled_date = 'Scheduled date is required';
    }

    // Validate deadline only if provided
    if (formData.deadline) {
      const selectedDate = new Date(formData.deadline);
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      
      if (selectedDate < today) {
        newErrors.deadline = 'Deadline cannot be in the past';
      }

      // Check if deadline is before scheduled date
      if (formData.scheduled_date) {
        const scheduledDate = new Date(formData.scheduled_date);
        if (selectedDate < scheduledDate) {
          newErrors.deadline = 'Deadline cannot be before scheduled date';
        }
      }
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!validateForm()) {
      return;
    }

    setIsSubmitting(true);
    
    try {
      const taskData = {
        project_id: project.id,
        jira_ticket_id: formData.jira_ticket_id.trim(),
        title: `Task for ${formData.jira_ticket_id.trim()}`,
        description: `Task associated with Jira ticket ${formData.jira_ticket_id.trim()}`,
        scheduled_date: new Date(formData.scheduled_date + 'T00:00:00').toISOString(),
        deadline: formData.deadline ? new Date(formData.deadline + 'T23:59:59').toISOString() : null,
        status: 'pending'
      };

      await CreateTask(taskData);
      onTaskCreated();
    } catch (err) {
      setErrors({ submit: err.message || 'Failed to create task' });
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
  };

  // Get today's date in YYYY-MM-DD format for min attribute
  const today = new Date().toISOString().split('T')[0];

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl w-full max-w-md mx-4">
        <div className="flex items-center justify-between p-6 border-b border-gray-200">
          <h2 className="text-xl font-semibold text-gray-900">Add New Task</h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-600"
          >
            <X className="w-6 h-6" />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-6">
          <div className="mb-4">
            <p className="text-sm text-gray-600 mb-4">
              Adding task to: <span className="font-medium text-gray-900">{project.name}</span>
            </p>
          </div>

          <div className="space-y-4">
            <div>
              <label htmlFor="jira_ticket_id" className="block text-sm font-medium text-gray-700 mb-1">
                Jira Ticket ID *
              </label>
              <input
                type="text"
                id="jira_ticket_id"
                name="jira_ticket_id"
                value={formData.jira_ticket_id}
                onChange={handleInputChange}
                className={`w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.jira_ticket_id ? 'border-red-500' : 'border-gray-300'
                }`}
                placeholder="e.g., ABC-123, PROJ-456"
                disabled={isSubmitting}
              />
              {errors.jira_ticket_id && (
                <p className="text-red-500 text-sm mt-1">{errors.jira_ticket_id}</p>
              )}
              <p className="text-xs text-gray-500 mt-1">
                Enter the Jira ticket identifier for this task
              </p>
            </div>

            <div>
              <label htmlFor="scheduled_date" className="block text-sm font-medium text-gray-700 mb-1">
                Scheduled Date *
              </label>
              <input
                type="date"
                id="scheduled_date"
                name="scheduled_date"
                value={formData.scheduled_date}
                onChange={handleInputChange}
                className={`w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.scheduled_date ? 'border-red-500' : 'border-gray-300'
                }`}
                disabled={isSubmitting}
              />
              {errors.scheduled_date && (
                <p className="text-red-500 text-sm mt-1">{errors.scheduled_date}</p>
              )}
              <p className="text-xs text-gray-500 mt-1">
                Select the date you plan to work on this task
              </p>
            </div>

            <div>
              <label htmlFor="deadline" className="block text-sm font-medium text-gray-700 mb-1">
                Deadline (Optional)
              </label>
              <input
                type="date"
                id="deadline"
                name="deadline"
                value={formData.deadline}
                onChange={handleInputChange}
                min={formData.scheduled_date || today}
                className={`w-full border rounded-lg px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-500 ${
                  errors.deadline ? 'border-red-500' : 'border-gray-300'
                }`}
                disabled={isSubmitting}
              />
              {errors.deadline && (
                <p className="text-red-500 text-sm mt-1">{errors.deadline}</p>
              )}
              <p className="text-xs text-gray-500 mt-1">
                Optional: Select the final due date for this task
              </p>
            </div>
          </div>

          {errors.submit && (
            <div className="mt-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">
              {errors.submit}
            </div>
          )}

          <div className="flex justify-end gap-3 mt-6">
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
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Creating...' : 'Create Task'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default TaskModal;