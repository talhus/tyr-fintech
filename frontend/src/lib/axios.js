import axios from 'axios';

const api = axios.create({
  baseURL: 'http://localhost:8080/api/v1',
  withCredentials: true,
});

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      if (error.response.status === 401) {
        localStorage.removeItem('user');
        if (window.location.pathname !== '/login' && window.location.pathname !== '/register') {
          window.location.href = '/login';
        }
      } else if (error.response.status === 429) {
        const retryAfter = error.response.headers['retry-after'];
        const message = retryAfter 
          ? `Rate limit reached. Please try again in ${retryAfter} seconds.`
          : 'Too many requests. Please wait a moment before trying again.';
        error.response.data = { ...error.response.data, error: message };
      }
    }
    return Promise.reject(error);
  }
);

export default api;
