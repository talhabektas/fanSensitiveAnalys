import axios from 'axios';
import toast from 'react-hot-toast';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8060/api/v1';

const api = axios.create({
  baseURL: API_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json; charset=UTF-8',
    'Accept': 'application/json; charset=UTF-8',
    'X-API-Key': 'fan-sentiment-2024-secret',
  },
});

// Request interceptor
api.interceptors.request.use(
  (config) => {
    console.log(`Making ${config.method?.toUpperCase()} request to ${config.url}`);
    return config;
  },
  (error) => {
    console.error('Request error:', error);
    return Promise.reject(error);
  }
);

// Response interceptor
api.interceptors.response.use(
  (response) => {
    return response;
  },
  (error) => {
    console.error('Response error:', error);
    
    if (error.response) {
      const { status, data } = error.response;
      
      switch (status) {
        case 400:
          toast.error(data.message || 'Geçersiz istek');
          break;
        case 401:
          toast.error('Yetkilendirme hatası');
          break;
        case 403:
          toast.error('Erişim reddedildi');
          break;
        case 404:
          toast.error('Kaynak bulunamadı');
          break;
        case 429:
          toast.error('Çok fazla istek. Lütfen bekleyin.');
          break;
        case 500:
          toast.error('Sunucu hatası. Lütfen daha sonra tekrar deneyin.');
          break;
        default:
          toast.error(data.message || 'Bilinmeyen bir hata oluştu');
      }
    } else if (error.request) {
      toast.error('Sunucuya bağlanılamadı. İnternet bağlantınızı kontrol edin.');
    } else {
      toast.error('Beklenmeyen bir hata oluştu');
    }
    
    return Promise.reject(error);
  }
);

// Utility function to handle API responses
const handleApiResponse = (response) => {
  if (response.data) {
    return response.data;
  }
  return response;
};

// Utility function to handle API errors
const handleApiError = (error) => {
  if (error.response?.data?.message) {
    throw new Error(error.response.data.message);
  }
  throw error;
};

// Generic API methods
export const apiClient = {
  get: async (url, config = {}) => {
    try {
      const response = await api.get(url, config);
      return handleApiResponse(response);
    } catch (error) {
      handleApiError(error);
    }
  },

  post: async (url, data, config = {}) => {
    try {
      const response = await api.post(url, data, config);
      return handleApiResponse(response);
    } catch (error) {
      handleApiError(error);
    }
  },

  put: async (url, data, config = {}) => {
    try {
      const response = await api.put(url, data, config);
      return handleApiResponse(response);
    } catch (error) {
      handleApiError(error);
    }
  },

  patch: async (url, data, config = {}) => {
    try {
      const response = await api.patch(url, data, config);
      return handleApiResponse(response);
    } catch (error) {
      handleApiError(error);
    }
  },

  delete: async (url, config = {}) => {
    try {
      const response = await api.delete(url, config);
      return handleApiResponse(response);
    } catch (error) {
      handleApiError(error);
    }
  },
};

// Reddit Live Stream service
export const redditLiveService = {
  restartStream: async () => {
    try {
      const response = await api.post('/live/reddit/start', {
        subreddits: [
          'Turkey', 'superlig', 'soccer',
          'galatasaray', 'fenerbahce', 'besiktas', 'trabzonspor',
          'turkishfootball', 'SuperLigTurkey'
        ]
      });
      return handleApiResponse(response);
    } catch (error) {
      handleApiError(error);
    }
  },

  getStatus: async () => {
    try {
      const response = await api.get('/live/reddit/status');
      return handleApiResponse(response);
    } catch (error) {
      handleApiError(error);
    }
  }
};

export default api;