import React, { useEffect, useState } from 'react';
import { api } from '../../services/api';

const Config = () => {
  const [configs, setConfigs] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchConfigs();
  }, []);

  const fetchConfigs = async () => {
    try {
      const response = await api.get('/config');
      setConfigs(response.data.configs || []);
    } catch (error) {
      console.error('Failed to fetch configs:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="module-page">
      <h2>⚙️ System Configuration</h2>

      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>Key</th>
              <th>Value</th>
              <th>Type</th>
            </tr>
          </thead>
          <tbody>
            {configs.map((config) => (
              <tr key={config.key}>
                <td><strong>{config.key}</strong></td>
                <td>{config.value}</td>
                <td><span className="badge">{config.type}</span></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default Config;
