import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { authAPI } from '../services/api';

const Dashboard = () => {
  const { user, logout, setUser } = useAuth();
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const navigate = useNavigate();

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const data = await authAPI.getProfile();
        setProfile(data);
        setUser(data);
      } catch (err) {
        setError('Failed to load profile');
        console.error(err);
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, [setUser]);

  const handleLogout = () => {
    logout();
    navigate('/login');
  };

  if (loading) {
    return (
      <div className="dashboard">
        <h2>Loading...</h2>
      </div>
    );
  }

  return (
    <div className="dashboard">
      <h2>Dashboard</h2>
      {error && <div className="error-message">{error}</div>}
      {profile && (
        <div className="user-info">
          <h3>Welcome, {profile.name}!</h3>
          <p><strong>Email:</strong> {profile.email}</p>
          <p><strong>User ID:</strong> {profile.id}</p>
          <p><strong>Member since:</strong> {new Date(profile.created_at).toLocaleDateString()}</p>
        </div>
      )}
      <button onClick={handleLogout} className="btn">
        Logout
      </button>
    </div>
  );
};

export default Dashboard;
