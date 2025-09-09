import React, { useState, useEffect } from 'react';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { 
  Home, 
  Database, 
  Package, 
  Server,
  Activity,
  Github,
  FolderOpen,
  Calendar,
  CheckSquare,
  Settings,
  ChevronDown,
  GitPullRequest,
  GitCommit,
  Cloud,
  Clock
} from 'lucide-react';

const Layout = ({ children }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const [services, setServices] = useState([]);
  const [selectedService, setSelectedService] = useState('');
  const [selectedServiceId, setSelectedServiceId] = useState('');
  const [isDropdownOpen, setIsDropdownOpen] = useState(false);

  // Extract service ID from current URL if we're on a service page
  useEffect(() => {
    const serviceRouteMatch = location.pathname.match(/^\/service\/(\d+)/);
    if (serviceRouteMatch) {
      const serviceId = serviceRouteMatch[1];
      setSelectedServiceId(serviceId);
      
      // Find the service name from services list
      const service = services.find(s => s.id === parseInt(serviceId));
      if (service) {
        setSelectedService(service.name);
      }
    } else {
      // Clear selection if not on a service page
      setSelectedServiceId('');
      setSelectedService('');
    }
  }, [location.pathname, services]);

  // Load services for dropdown
  useEffect(() => {
    loadServices();
  }, []);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (isDropdownOpen && !event.target.closest('.service-dropdown')) {
        setIsDropdownOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isDropdownOpen]);

  const loadServices = async () => {
    try {
      const allServices = await window.go.main.App.GetMicroservices(0);
      setServices(allServices || []);
    } catch (error) {
      console.error('Failed to load services for dropdown:', error);
      setServices([]);
    }
  };

  const handleServiceSelect = (serviceId, serviceName) => {
    setSelectedService(serviceName);
    setSelectedServiceId(serviceId);
    setIsDropdownOpen(false);
    navigate(`/service/${serviceId}`);
  };

  const baseNavigation = [
    { name: 'Dashboard', href: '/', icon: Home },
    { name: 'Repositories', href: '/repositories', icon: Database },
    { name: 'Microservices', href: '/microservices', icon: Package },
    { name: 'Projects', href: '/projects', icon: FolderOpen },
    { name: 'Tasks', href: '/tasks', icon: CheckSquare },
    { name: 'Calendar', href: '/calendar', icon: Calendar },
    { name: 'Settings', href: '/settings', icon: Settings },
  ];

  const getNavigation = () => {
    if (!selectedServiceId || !selectedService) {
      return baseNavigation;
    }

    // Add service-specific navigation items after Microservices
    const serviceNavItems = [
      { 
        name: `${selectedService}`, 
        href: `/service/${selectedServiceId}`, 
        icon: Package,
        isServiceHeader: true 
      },
      { 
        name: `Pull Requests`, 
        href: `/service/${selectedServiceId}/pull-requests`, 
        icon: GitPullRequest,
        isServiceItem: true 
      },
      { 
        name: `Commits`, 
        href: `/service/${selectedServiceId}/commits`, 
        icon: GitCommit,
        isServiceItem: true 
      },
      { 
        name: `Deployments Overview`, 
        href: `/service/${selectedServiceId}/deployments`, 
        icon: Cloud,
        isServiceItem: true 
      },
      { 
        name: `Deployment History`, 
        href: `/service/${selectedServiceId}/deployment-history`, 
        icon: Clock,
        isServiceItem: true 
      },
    ];

    // Insert service items after Microservices (index 2)
    return [
      ...baseNavigation.slice(0, 3),
      { isDivider: true },
      ...serviceNavItems,
      { isDivider: true },
      ...baseNavigation.slice(3)
    ];
  };

  const navigation = getNavigation();

  const isActive = (href) => {
    if (href === '/') {
      return location.pathname === '/';
    }
    return location.pathname.startsWith(href);
  };

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar */}
      <div className="fixed inset-y-0 left-0 z-50 w-64 bg-white shadow-lg">
        <div className="flex h-16 items-center px-6 border-b border-gray-200">
          <Github className="h-8 w-8 text-blue-600" />
          <span className="ml-2 text-xl font-semibold text-gray-900">
            Dev Dashboard
          </span>
        </div>
        
        <nav className="mt-8 px-4 space-y-1">
          {navigation.map((item, index) => {
            // Render divider
            if (item.isDivider) {
              return (
                <div key={`divider-${index}`} className="my-4">
                  <div className="border-t border-gray-200"></div>
                </div>
              );
            }

            const Icon = item.icon;
            const isServiceHeader = item.isServiceHeader;
            const isServiceItem = item.isServiceItem;
            
            return (
              <Link
                key={item.name}
                to={item.href}
                className={`
                  group flex items-center px-2 py-2 text-sm rounded-md transition-colors
                  ${isServiceHeader 
                    ? 'font-semibold bg-blue-50 text-blue-800 hover:bg-blue-100'
                    : isServiceItem
                    ? 'pl-8 font-medium text-gray-600 hover:bg-gray-100 hover:text-gray-900'
                    : 'font-medium'
                  }
                  ${!isServiceHeader && !isServiceItem && isActive(item.href)
                    ? 'bg-blue-100 text-blue-700'
                    : !isServiceHeader && !isServiceItem
                    ? 'text-gray-700 hover:bg-gray-100 hover:text-gray-900'
                    : ''
                  }
                  ${(isServiceHeader || isServiceItem) && isActive(item.href)
                    ? 'bg-blue-200 text-blue-900'
                    : ''
                  }
                `}
              >
                <Icon 
                  className={`
                    ${isServiceItem ? 'mr-3 h-4 w-4' : 'mr-3 h-5 w-5'} flex-shrink-0
                    ${isServiceHeader 
                      ? 'text-blue-600'
                      : isServiceItem
                      ? 'text-gray-400 group-hover:text-gray-500'
                      : isActive(item.href) 
                      ? 'text-blue-500' 
                      : 'text-gray-400 group-hover:text-gray-500'
                    }
                  `} 
                />
                {item.name}
              </Link>
            );
          })}
        </nav>

        <div className="absolute bottom-4 left-4 right-4">
          <div className="flex items-center p-3 bg-gray-100 rounded-lg">
            <Activity className="h-5 w-5 text-green-500 mr-2" />
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-gray-900 truncate">
                Sync Status
              </p>
              <p className="text-sm text-gray-500 truncate">
                Last updated: Just now
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Main content */}
      <div className="flex-1 ml-64">
        {/* Top bar with service selector */}
        <div className="bg-white border-b border-gray-200 px-8 py-4">
          <div className="flex justify-end">
            {/* Service Selector Dropdown */}
            <div className="relative service-dropdown">
              <button
                onClick={() => setIsDropdownOpen(!isDropdownOpen)}
                className="flex items-center px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 rounded-md hover:bg-gray-200 transition-colors min-w-[200px] justify-between"
              >
                <div className="flex items-center">
                  <Package className="h-4 w-4 mr-2" />
                  <span className="truncate">{selectedService || 'Select Service'}</span>
                </div>
                <ChevronDown className="h-4 w-4 ml-2 flex-shrink-0" />
              </button>
              
              {isDropdownOpen && (
                <div className="absolute right-0 mt-2 w-80 bg-white rounded-md shadow-lg z-50 border border-gray-200">
                  <div className="py-1 max-h-64 overflow-y-auto">
                    {services.length > 0 ? (
                      services.map((service) => (
                        <button
                          key={service.id}
                          onClick={() => handleServiceSelect(service.id, service.name)}
                          className="w-full text-left px-4 py-3 text-sm text-gray-700 hover:bg-gray-100 transition-colors"
                        >
                          <div className="flex items-start">
                            <Package className="h-4 w-4 mr-3 mt-0.5 text-blue-500 flex-shrink-0" />
                            <div className="min-w-0 flex-1">
                              <div className="font-medium text-gray-900 truncate">{service.name}</div>
                              <div className="text-xs text-gray-500 truncate mt-0.5">{service.path}</div>
                              {service.description && (
                                <div className="text-xs text-gray-400 truncate mt-1">{service.description}</div>
                              )}
                            </div>
                          </div>
                        </button>
                      ))
                    ) : (
                      <div className="px-4 py-3 text-sm text-gray-500">
                        No services available
                      </div>
                    )}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
        
        <main className="p-8">
          {children}
        </main>
      </div>
    </div>
  );
};

export default Layout;