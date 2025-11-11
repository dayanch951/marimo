import React, { useEffect, useState } from 'react';
import { api } from '../../services/api';

const Accounting = () => {
  const [balance, setBalance] = useState(null);
  const [transactions, setTransactions] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchData();
  }, []);

  const fetchData = async () => {
    try {
      const [balanceRes, txRes] = await Promise.all([
        api.get('/accounting/balance'),
        api.get('/accounting/transactions'),
      ]);
      setBalance(balanceRes.data);
      setTransactions(txRes.data.transactions || []);
    } catch (error) {
      console.error('Failed to fetch accounting data:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) return <div className="loading">Loading...</div>;

  return (
    <div className="module-page">
      <h2>💰 Accounting</h2>

      {balance && (
        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-value" style={{ color: '#43e97b' }}>
              ${balance.income?.toFixed(2) || '0.00'}
            </div>
            <div className="stat-label">Total Income</div>
          </div>
          <div className="stat-card">
            <div className="stat-value" style={{ color: '#f093fb' }}>
              ${balance.expense?.toFixed(2) || '0.00'}
            </div>
            <div className="stat-label">Total Expenses</div>
          </div>
          <div className="stat-card">
            <div className="stat-value" style={{ color: '#667eea' }}>
              ${balance.balance?.toFixed(2) || '0.00'}
            </div>
            <div className="stat-label">Balance</div>
          </div>
        </div>
      )}

      <h3>Recent Transactions</h3>
      <div className="table-container">
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Type</th>
              <th>Amount</th>
              <th>Category</th>
              <th>Description</th>
              <th>Date</th>
            </tr>
          </thead>
          <tbody>
            {transactions.map((tx) => (
              <tr key={tx.id}>
                <td>{tx.id}</td>
                <td>
                  <span className={`badge badge-${tx.type}`}>
                    {tx.type}
                  </span>
                </td>
                <td>${tx.amount.toFixed(2)}</td>
                <td>{tx.category}</td>
                <td>{tx.description}</td>
                <td>{new Date(tx.created_at).toLocaleDateString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default Accounting;
