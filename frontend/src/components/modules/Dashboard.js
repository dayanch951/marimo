import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { api } from '../../services/api';

const Dashboard = () => {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    fetchDashboard();
  }, []);

  const fetchDashboard = async () => {
    try {
      const response = await api.get('/main/dashboard');
      setStats(response.data);
    } catch (error) {
      console.error('Failed to fetch dashboard:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="loading">Loading...</div>;

  const modules = [
    { name: 'Users', path: '/users', icon: '👥', color: '#667eea' },
    { name: 'Config', path: '/config', icon: '⚙️', color: '#764ba2' },
    { name: 'Accounting', path: '/accounting', icon: '💰', color: '#f093fb' },
    { name: 'Factory', path: '/factory', icon: '🏭', color: '#4facfe' },
    { name: 'Shop', path: '/shop', icon: '🛍️', color: '#43e97b' },
  ];

  return (
    <div className="dashboard-page">
      <div className="dashboard-header">
        <h2>{stats?.welcome || 'Welcome to Marimo ERP'}</h2>
        <p>User: {stats?.user?.email} ({stats?.user?.role})</p>
      </div>

      <div className="modules-grid">
        {modules.map((module) => (
          <div
            key={module.path}
            className="module-card"
            style={{ borderColor: module.color }}
            onClick={() => navigate(module.path)}
          >
            <div className="module-icon" style={{ background: module.color }}>
              {module.icon}
            </div>
            <h3>{module.name}</h3>
          </div>
        ))}
      </div>
    </div>
  );
};

export default Dashboard;
