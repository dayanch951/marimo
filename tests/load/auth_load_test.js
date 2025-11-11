// k6 Load Testing Script for Authentication API
// Usage: k6 run auth_load_test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');

// Test configuration
export const options = {
  stages: [
    { duration: '30s', target: 50 },  // Ramp up to 50 users
    { duration: '1m', target: 50 },   // Stay at 50 users
    { duration: '30s', target: 100 }, // Ramp up to 100 users
    { duration: '1m', target: 100 },  // Stay at 100 users
    { duration: '30s', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    'http_req_duration': ['p(95)<500'], // 95% of requests must complete below 500ms
    'http_req_failed': ['rate<0.1'],     // Error rate must be below 10%
    'errors': ['rate<0.1'],              // Custom error rate must be below 10%
  },
};

const API_URL = __ENV.API_URL || 'http://localhost:8080';

// Generate unique user credentials
function generateUser() {
  const timestamp = Date.now();
  const random = Math.floor(Math.random() * 1000000);
  return {
    email: `loadtest-${timestamp}-${random}@example.com`,
    password: 'LoadTest123!',
    name: `Load Test User ${random}`,
  };
}

export default function() {
  const user = generateUser();

  // 1. Register new user
  const registerResponse = http.post(
    `${API_URL}/api/users/register`,
    JSON.stringify(user),
    { headers: { 'Content-Type': 'application/json' } }
  );

  const registerSuccess = check(registerResponse, {
    'register status is 201': (r) => r.status === 201,
    'register response has success': (r) => JSON.parse(r.body).success === true,
  });

  if (!registerSuccess) {
    errorRate.add(1);
    return;
  }

  sleep(1);

  // 2. Login
  const loginResponse = http.post(
    `${API_URL}/api/users/login`,
    JSON.stringify({
      email: user.email,
      password: user.password,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  const loginSuccess = check(loginResponse, {
    'login status is 200': (r) => r.status === 200,
    'login has access_token': (r) => JSON.parse(r.body).access_token !== undefined,
    'login has refresh_token': (r) => JSON.parse(r.body).refresh_token !== undefined,
  });

  if (!loginSuccess) {
    errorRate.add(1);
    return;
  }

  const loginData = JSON.parse(loginResponse.body);
  const accessToken = loginData.access_token;
  const refreshToken = loginData.refresh_token;

  sleep(1);

  // 3. Access protected resource
  const profileResponse = http.get(
    `${API_URL}/api/users/profile`,
    {
      headers: {
        'Authorization': `Bearer ${accessToken}`,
      },
    }
  );

  const profileSuccess = check(profileResponse, {
    'profile status is 200': (r) => r.status === 200,
    'profile has user email': (r) => JSON.parse(r.body).email === user.email,
  });

  if (!profileSuccess) {
    errorRate.add(1);
  }

  sleep(1);

  // 4. Refresh token
  const refreshResponse = http.post(
    `${API_URL}/api/users/refresh`,
    JSON.stringify({
      refresh_token: refreshToken,
    }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  const refreshSuccess = check(refreshResponse, {
    'refresh status is 200': (r) => r.status === 200,
    'refresh has new access_token': (r) => JSON.parse(r.body).access_token !== undefined,
  });

  if (!refreshSuccess) {
    errorRate.add(1);
  }

  sleep(1);
}

// Summary report after test
export function handleSummary(data) {
  return {
    'stdout': textSummary(data, { indent: ' ', enableColors: true }),
    'summary.json': JSON.stringify(data),
  };
}

function textSummary(data, options) {
  const { indent = '', enableColors = false } = options;
  const metrics = data.metrics;

  let output = '\n';
  output += `${indent}Test Results:\n`;
  output += `${indent}==============\n\n`;

  if (metrics.http_reqs) {
    output += `${indent}Total Requests: ${metrics.http_reqs.values.count}\n`;
  }

  if (metrics.http_req_duration) {
    output += `${indent}Request Duration:\n`;
    output += `${indent}  avg: ${metrics.http_req_duration.values.avg.toFixed(2)}ms\n`;
    output += `${indent}  min: ${metrics.http_req_duration.values.min.toFixed(2)}ms\n`;
    output += `${indent}  max: ${metrics.http_req_duration.values.max.toFixed(2)}ms\n`;
    output += `${indent}  p(95): ${metrics.http_req_duration.values['p(95)'].toFixed(2)}ms\n`;
  }

  if (metrics.http_req_failed) {
    const failRate = (metrics.http_req_failed.values.rate * 100).toFixed(2);
    output += `${indent}Error Rate: ${failRate}%\n`;
  }

  output += '\n';
  return output;
}
