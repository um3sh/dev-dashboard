import React, { useState, useEffect } from 'react';
import { GetTasksInDateRange, UpdateTaskStatus } from '../../wailsjs/go/main/App';
import { Calendar as CalendarIcon, ChevronLeft, ChevronRight, Clock, CheckCircle } from 'lucide-react';

const Calendar = () => {
  const [currentDate, setCurrentDate] = useState(new Date());
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedDate, setSelectedDate] = useState(null);
  const [selectedDateTasks, setSelectedDateTasks] = useState([]);

  useEffect(() => {
    loadTasks();
  }, [currentDate]);

  const loadTasks = async () => {
    setLoading(true);
    try {
      const startDate = new Date(currentDate.getFullYear(), currentDate.getMonth(), 1);
      const endDate = new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 0);
      
      const data = await GetTasksInDateRange(startDate.toISOString(), endDate.toISOString());
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
      
      // Update selected date tasks if needed
      if (selectedDate) {
        const dateStr = selectedDate.toDateString();
        const dateTasks = tasks.filter(task => 
          new Date(task.deadline).toDateString() === dateStr
        );
        setSelectedDateTasks(dateTasks);
      }
    } catch (err) {
      setError('Failed to update task status: ' + err.message);
    }
  };

  const navigateMonth = (direction) => {
    const newDate = new Date(currentDate);
    newDate.setMonth(currentDate.getMonth() + direction);
    setCurrentDate(newDate);
    setSelectedDate(null);
    setSelectedDateTasks([]);
  };

  const getDaysInMonth = () => {
    const year = currentDate.getFullYear();
    const month = currentDate.getMonth();
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const daysInMonth = lastDay.getDate();
    const startingDayOfWeek = firstDay.getDay();

    const days = [];
    
    // Add empty cells for previous month
    for (let i = 0; i < startingDayOfWeek; i++) {
      days.push(null);
    }
    
    // Add days of current month
    for (let day = 1; day <= daysInMonth; day++) {
      days.push(new Date(year, month, day));
    }
    
    return days;
  };

  const getTasksForDate = (date) => {
    if (!date) return [];
    const dateStr = date.toDateString();
    return tasks.filter(task => 
      task.deadline && new Date(task.deadline).toDateString() === dateStr
    );
  };

  const handleDateClick = (date) => {
    if (!date) return;
    setSelectedDate(date);
    setSelectedDateTasks(getTasksForDate(date));
  };

  const getStatusColor = (status) => {
    switch (status) {
      case 'completed': return 'bg-green-100 text-green-800';
      case 'in_progress': return 'bg-blue-100 text-blue-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  const isToday = (date) => {
    if (!date) return false;
    const today = new Date();
    return date.toDateString() === today.toDateString();
  };

  const isOverdue = (task) => {
    if (!task.deadline) return false;
    const today = new Date();
    const deadline = new Date(task.deadline);
    return deadline < today && deadline.toDateString() !== today.toDateString() && task.status !== 'completed';
  };

  const monthNames = [
    'January', 'February', 'March', 'April', 'May', 'June',
    'July', 'August', 'September', 'October', 'November', 'December'
  ];

  const dayNames = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

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
          <h1 className="text-2xl font-bold text-gray-900">Task Calendar</h1>
          <p className="text-gray-600">View and manage task deadlines in calendar view</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
          {error}
        </div>
      )}

      <div className="grid grid-cols-12 gap-6">
        {/* Calendar */}
        <div className="col-span-8">
          <div className="bg-white rounded-lg shadow">
            {/* Calendar Header */}
            <div className="flex items-center justify-between p-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">
                {monthNames[currentDate.getMonth()]} {currentDate.getFullYear()}
              </h2>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => navigateMonth(-1)}
                  className="p-2 hover:bg-gray-100 rounded-lg"
                >
                  <ChevronLeft className="w-4 h-4" />
                </button>
                <button
                  onClick={() => navigateMonth(1)}
                  className="p-2 hover:bg-gray-100 rounded-lg"
                >
                  <ChevronRight className="w-4 h-4" />
                </button>
              </div>
            </div>

            {/* Calendar Grid */}
            <div className="p-4">
              <div className="grid grid-cols-7 gap-1 mb-2">
                {dayNames.map(day => (
                  <div key={day} className="text-center text-sm font-medium text-gray-700 py-2">
                    {day}
                  </div>
                ))}
              </div>
              <div className="grid grid-cols-7 gap-1">
                {getDaysInMonth().map((date, index) => {
                  const dateTasks = getTasksForDate(date);
                  const overdueTasks = dateTasks.filter(isOverdue);
                  
                  return (
                    <div
                      key={index}
                      className={`min-h-20 p-1 border border-gray-200 rounded cursor-pointer hover:bg-gray-50 ${
                        !date ? 'bg-gray-50' : ''
                      } ${
                        selectedDate?.toDateString() === date?.toDateString() ? 'bg-blue-50 border-blue-300' : ''
                      } ${
                        isToday(date) ? 'bg-yellow-50 border-yellow-300' : ''
                      }`}
                      onClick={() => handleDateClick(date)}
                    >
                      {date && (
                        <>
                          <div className={`text-sm font-medium mb-1 ${
                            isToday(date) ? 'text-yellow-600' : 'text-gray-900'
                          }`}>
                            {date.getDate()}
                          </div>
                          <div className="space-y-1">
                            {dateTasks.slice(0, 2).map(task => (
                              <div
                                key={task.id}
                                className={`text-xs p-1 rounded truncate ${
                                  isOverdue(task) ? 'bg-red-100 text-red-800' : getStatusColor(task.status)
                                }`}
                                title={`${task.project_name}: ${task.title}`}
                              >
                                {task.jira_ticket_id}
                              </div>
                            ))}
                            {dateTasks.length > 2 && (
                              <div className="text-xs text-gray-500">
                                +{dateTasks.length - 2} more
                              </div>
                            )}
                          </div>
                        </>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          </div>
        </div>

        {/* Tasks for Selected Date */}
        <div className="col-span-4">
          <div className="bg-white rounded-lg shadow">
            <div className="p-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900 flex items-center gap-2">
                <CalendarIcon className="w-5 h-5" />
                {selectedDate ? (
                  <>Tasks for {selectedDate.toLocaleDateString()}</>
                ) : (
                  'Select a Date'
                )}
              </h2>
            </div>
            <div className="p-4">
              {!selectedDate ? (
                <p className="text-gray-500 text-center">Click on a date to view tasks</p>
              ) : selectedDateTasks.length === 0 ? (
                <p className="text-gray-500 text-center">No tasks for this date</p>
              ) : (
                <div className="space-y-3">
                  {selectedDateTasks.map(task => (
                    <div key={task.id} className={`p-3 border rounded-lg ${
                      isOverdue(task) ? 'border-red-200 bg-red-50' : 'border-gray-200'
                    }`}>
                      <div className="flex items-start justify-between mb-2">
                        <div className="flex-1">
                          <h3 className="font-medium text-gray-900">{task.jira_ticket_id}</h3>
                          <p className="text-sm text-gray-600">{task.project_name}</p>
                        </div>
                        {isOverdue(task) && (
                          <span className="text-red-600 text-xs flex items-center gap-1">
                            <Clock className="w-3 h-3" />
                            Overdue
                          </span>
                        )}
                      </div>
                      
                      <div className="flex items-center justify-between">
                        <span className={`px-2 py-1 text-xs font-medium rounded-full ${getStatusColor(task.status)}`}>
                          {task.status.replace('_', ' ')}
                        </span>
                        
                        {task.status !== 'completed' && (
                          <div className="flex gap-1">
                            {task.status === 'pending' && (
                              <button
                                onClick={() => handleStatusChange(task.id, 'in_progress')}
                                className="text-blue-600 hover:text-blue-800 text-xs px-2 py-1 border border-blue-600 rounded"
                              >
                                Start
                              </button>
                            )}
                            <button
                              onClick={() => handleStatusChange(task.id, 'completed')}
                              className="text-green-600 hover:text-green-800 text-xs flex items-center gap-1"
                            >
                              <CheckCircle className="w-3 h-3" />
                              Complete
                            </button>
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Calendar;