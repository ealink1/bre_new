import axios from 'axios';

const api = axios.create({
  // baseURL: 'http://localhost:8080/api',
  baseURL: 'https://cj-api.wsky.fun/api',
  timeout: 10000,
});

export const fetchLatestNews = () => api.get('/news/latest');
export const fetchAnalysis = (days) => api.get(`/analysis/latest?days=${days}`);
