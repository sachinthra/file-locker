import axios from 'axios';
import { getToken } from './auth';

const API_BASE_URL = 'http://localhost:9010/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Add token to requests
api.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Auth APIs
export const login = (username, password) => {
  return api.post('/auth/login', { username, password });
};

export const register = (username, password, email) => {
  return api.post('/auth/register', { username, password, email });
};

export const logout = () => {
  return api.post('/auth/logout');
};

// File APIs
export const uploadFile = (file, tags, expireAfter, onProgress) => {
  const formData = new FormData();
  formData.append('file', file);
  if (tags) formData.append('tags', tags);
  if (expireAfter) formData.append('expire_after', expireAfter);

  return api.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress: onProgress,
  });
};

export const listFiles = () => {
  return api.get('/files');
};

export const searchFiles = (query) => {
  return api.get(`/files/search?q=${encodeURIComponent(query)}`);
};

export const deleteFile = (fileId) => {
  return api.delete(`/files?id=${fileId}`);
};

export const getDownloadUrl = (fileId) => {
  const token = getToken();
  return `${API_BASE_URL}/download/${fileId}?token=${token}`;
};

export const getStreamUrl = (fileId) => {
  const token = getToken();
  return `${API_BASE_URL}/stream/${fileId}?token=${token}`;
};

export default api;
