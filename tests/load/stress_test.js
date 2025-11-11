// k6 Stress Testing Script
// Tests system behavior under extreme load
// Usage: k6 run stress_test.js

import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },  // Ramp up to 100 users
    { duration: '5m', target: 100 },  // Stay at 100
    { duration: '2m', target: 200 },  // Ramp up to 200 users
    { duration: '5m', target: 200 },  // Stay at 200
    { duration: '2m', target: 300 },  // Ramp up to 300 users
    { duration: '5m', target: 300 },  // Stay at 300
    { duration: '5m', target: 0 },    // Ramp down to 0
  ],
  thresholds: {
    'http_req_duration': ['p(99)<1000'], // 99% under 1s
    'http_req_failed': ['rate<0.2'],     // Error rate under 20%
  },
};

const API_URL = __ENV.API_URL || 'http://localhost:8080';

export default function() {
  const responses = http.batch([
    ['GET', `${API_URL}/health`],
    ['GET', `${API_URL}/health`],
    ['GET', `${API_URL}/health`],
  ]);

  check(responses[0], {
    'health check status is 200': (r) => r.status === 200,
  });

  sleep(0.1);
}
