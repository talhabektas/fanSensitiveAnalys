import { apiClient } from './api';

export const commentService = {
  // Get comments with filtering and pagination
  getComments: async (params = {}) => {
    const queryString = new URLSearchParams(params).toString();
    return apiClient.get(`/comments?${queryString}`);
  },

  // Create a new comment
  createComment: async (commentData) => {
    return apiClient.post('/comments', commentData);
  },

  // Get unprocessed comments
  getUnprocessedComments: async (limit = 50) => {
    return apiClient.get(`/comments/unprocessed?limit=${limit}`);
  },

  // Get comment statistics
  getCommentStats: async () => {
    return apiClient.get('/comments/stats');
  },

  // Get a single comment by ID
  getComment: async (id) => {
    return apiClient.get(`/comments/${id}`);
  },

  // Update a comment
  updateComment: async (id, updateData) => {
    return apiClient.put(`/comments/${id}`, updateData);
  },

  // Bulk update comments as processed
  bulkUpdateProcessed: async (commentIds) => {
    return apiClient.post('/comments/bulk/processed', {
      comment_ids: commentIds,
    });
  },

  // Filter comments by team
  getCommentsByTeam: async (teamId, params = {}) => {
    const queryParams = { ...params, team_id: teamId };
    const queryString = new URLSearchParams(queryParams).toString();
    return apiClient.get(`/comments?${queryString}`);
  },

  // Filter comments by source
  getCommentsBySource: async (source, params = {}) => {
    const queryParams = { ...params, source };
    const queryString = new URLSearchParams(queryParams).toString();
    return apiClient.get(`/comments?${queryString}`);
  },

  // Filter comments by sentiment
  getCommentsBySentiment: async (sentiment, params = {}) => {
    const queryParams = { ...params, sentiment };
    const queryString = new URLSearchParams(queryParams).toString();
    return apiClient.get(`/comments?${queryString}`);
  },

  // Filter comments by date range
  getCommentsByDateRange: async (startDate, endDate, params = {}) => {
    const queryParams = {
      ...params,
      start_date: startDate,
      end_date: endDate,
    };
    const queryString = new URLSearchParams(queryParams).toString();
    return apiClient.get(`/comments?${queryString}`);
  },

  // Search comments
  searchComments: async (query, params = {}) => {
    const queryParams = { ...params, search: query };
    const queryString = new URLSearchParams(queryParams).toString();
    return apiClient.get(`/comments?${queryString}`);
  },

  // Get recent comments
  getRecentComments: async (limit = 10) => {
    return apiClient.get(`/comments?limit=${limit}&sort_by=created_at&sort_order=desc`);
  },

  // Export comments data
  exportComments: async (params = {}, format = 'json') => {
    const queryParams = { ...params, export: format };
    const queryString = new URLSearchParams(queryParams).toString();
    return apiClient.get(`/comments/export?${queryString}`, {
      responseType: format === 'csv' ? 'blob' : 'json',
    });
  },
};