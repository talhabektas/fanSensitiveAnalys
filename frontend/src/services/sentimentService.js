import { apiClient } from './api';

export const sentimentService = {
  // Analyze text sentiment
  analyzeText: async (text, language = 'tr') => {
    return apiClient.post('/sentiments/analyze', {
      text,
      language,
    });
  },

  // Batch analyze multiple texts
  analyzeBatch: async (texts) => {
    return apiClient.post('/sentiments/analyze/batch', {
      texts,
    });
  },

  // Get sentiment statistics
  getSentimentStats: async () => {
    return apiClient.get('/sentiments/stats');
  },

  // Get team sentiment report
  getTeamReport: async (teamId, startDate = null, endDate = null) => {
    let url = `/sentiments/report/${teamId}`;
    const params = new URLSearchParams();
    
    if (startDate) {
      params.append('start_date', startDate);
    }
    if (endDate) {
      params.append('end_date', endDate);
    }
    
    if (params.toString()) {
      url += `?${params.toString()}`;
    }
    
    return apiClient.get(url);
  },

  // Save sentiment result
  saveSentiment: async (commentId, teamId, result) => {
    return apiClient.post('/sentiments', {
      comment_id: commentId,
      team_id: teamId,
      result,
    });
  },

  // Get sentiment trends over time
  getSentimentTrends: async (days = 30) => {
    const endDate = new Date();
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - days);
    
    return apiClient.get('/dashboard/stats');
  },

  // Get sentiment breakdown by team
  getTeamSentimentBreakdown: async () => {
    return apiClient.get('/dashboard/comparison');
  },

  // Get hourly sentiment distribution
  getHourlySentimentDistribution: async (teamId, days = 7) => {
    const endDate = new Date().toISOString().split('T')[0];
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - days);
    const startDateStr = startDate.toISOString().split('T')[0];
    
    return sentimentService.getTeamReport(teamId, startDateStr, endDate);
  },

  // Get sentiment by source platform
  getSentimentBySource: async (source, params = {}) => {
    const queryParams = { ...params, source };
    const queryString = new URLSearchParams(queryParams).toString();
    return apiClient.get(`/comments?has_sentiment=true&${queryString}`);
  },

  // Get model performance metrics
  getModelPerformance: async () => {
    return apiClient.get('/sentiments/stats');
  },

  // Get confidence distribution
  getConfidenceDistribution: async () => {
    return apiClient.get('/sentiments/stats');
  },

  // Get sentiment summary for dashboard
  getSentimentSummary: async () => {
    return apiClient.get('/dashboard/data');
  },

  // Process unprocessed comments
  processUnprocessedComments: async () => {
    const unprocessedComments = await apiClient.get('/comments/unprocessed');
    
    if (unprocessedComments.comments && unprocessedComments.comments.length > 0) {
      const texts = unprocessedComments.comments.map(comment => comment.text);
      const results = await sentimentService.analyzeBatch(texts);
      
      // Save results back to comments
      const savePromises = unprocessedComments.comments.map((comment, index) => {
        if (results[index]) {
          return sentimentService.saveSentiment(
            comment.id,
            comment.team_id,
            results[index]
          );
        }
        return Promise.resolve();
      });
      
      await Promise.all(savePromises);
      return { processed: unprocessedComments.comments.length };
    }
    
    return { processed: 0 };
  },
};