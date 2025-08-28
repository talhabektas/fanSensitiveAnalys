import { apiClient } from './api';

export const dashboardService = {
  // Get all dashboard data in one request
  getDashboardData: async () => {
    return apiClient.get('/dashboard/data');
  },

  // Get overall statistics
  getOverallStats: async () => {
    return apiClient.get('/dashboard/stats');
  },

  // Get team comparison data
  getTeamComparison: async () => {
    return apiClient.get('/dashboard/comparison');
  },

  // Collect data from Reddit
  collectRedditData: async () => {
    return apiClient.post('/reddit/collect');
  },

  // Collect data from specific subreddit
  collectFromSubreddit: async (subredditName) => {
    return apiClient.post(`/reddit/subreddit/${subredditName}`);
  },

  // Get system health status
  getHealthStatus: async () => {
    return apiClient.get('/health');
  },

  // Get API status
  getApiStatus: async () => {
    return apiClient.get('/');
  },

  // Process unprocessed comments
  processComments: async () => {
    return apiClient.get('/webhook/unprocessed');
  },

  // Get recent activity summary
  getRecentActivity: async (hours = 24) => {
    const endDate = new Date();
    const startDate = new Date();
    startDate.setHours(startDate.getHours() - hours);
    
    const startDateStr = startDate.toISOString().split('T')[0];
    const endDateStr = endDate.toISOString().split('T')[0];
    
    // This would need to be implemented in backend
    return {
      comments_collected: 0,
      comments_analyzed: 0,
      sentiment_processed: 0,
      period: `${hours} hours`,
    };
  },

  // Get performance metrics
  getPerformanceMetrics: async () => {
    const stats = await dashboardService.getOverallStats();
    const health = await dashboardService.getHealthStatus();
    
    return {
      total_comments: stats.total_analyzed || 0,
      processing_rate: 0, // Would need backend implementation
      api_response_time: 0, // Would need backend implementation
      system_health: health.status,
      services_status: health.services,
    };
  },

  // Get data collection summary
  getDataCollectionSummary: async () => {
    try {
      const commentStats = await apiClient.get('/comments/stats');
      const sentimentStats = await apiClient.get('/sentiments/stats');
      
      return {
        total_comments: commentStats.total_comments || 0,
        processed_comments: commentStats.processed_comments || 0,
        unprocessed_comments: commentStats.unprocessed_comments || 0,
        total_analyzed: sentimentStats.total_analyzed || 0,
        source_breakdown: commentStats.source_breakdown || {},
        sentiment_breakdown: sentimentStats.sentiment_breakdown || {},
      };
    } catch (error) {
      console.error('Error getting data collection summary:', error);
      return {
        total_comments: 0,
        processed_comments: 0,
        unprocessed_comments: 0,
        total_analyzed: 0,
        source_breakdown: {},
        sentiment_breakdown: {},
      };
    }
  },

  // Get sentiment trends for chart
  getSentimentTrends: async (days = 7) => {
    try {
      const stats = await dashboardService.getOverallStats();
      return stats.recent_trends || [];
    } catch (error) {
      console.error('Error getting sentiment trends:', error);
      return [];
    }
  },

  // Get top performing teams
  getTopTeams: async (limit = 5) => {
    try {
      const comparison = await dashboardService.getTeamComparison();
      const teams = comparison.teams || [];
      
      return teams
        .sort((a, b) => b.avg_sentiment - a.avg_sentiment)
        .slice(0, limit);
    } catch (error) {
      console.error('Error getting top teams:', error);
      return [];
    }
  },

  // Get real-time updates (polling-based)
  getRealTimeUpdates: async () => {
    try {
      const [commentStats, sentimentStats] = await Promise.all([
        apiClient.get('/comments/stats'),
        apiClient.get('/sentiments/stats'),
      ]);
      
      return {
        timestamp: new Date().toISOString(),
        comments: {
          total: commentStats.total_comments || 0,
          unprocessed: commentStats.unprocessed_comments || 0,
        },
        sentiments: {
          total: sentimentStats.total_analyzed || 0,
          breakdown: sentimentStats.sentiment_breakdown || {},
        },
      };
    } catch (error) {
      console.error('Error getting real-time updates:', error);
      return null;
    }
  },

  // Calculate sentiment score (normalize to 0-100)
  calculateSentimentScore: (sentimentBreakdown) => {
    if (!sentimentBreakdown) return 50;
    
    const positive = sentimentBreakdown.POSITIVE || 0;
    const negative = sentimentBreakdown.NEGATIVE || 0;
    const neutral = sentimentBreakdown.NEUTRAL || 0;
    const total = positive + negative + neutral;
    
    if (total === 0) return 50;
    
    // Calculate weighted score
    const score = ((positive * 100) + (neutral * 50) + (negative * 0)) / total;
    return Math.round(score);
  },

  // Format numbers for display
  formatNumber: (num) => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  },

  // Get sentiment color based on score
  getSentimentColor: (score) => {
    if (score >= 70) return '#10B981'; // Green
    if (score >= 40) return '#6B7280'; // Gray
    return '#EF4444'; // Red
  },

  // Export dashboard data
  exportDashboardData: async (format = 'json') => {
    const data = await dashboardService.getDashboardData();
    
    if (format === 'csv') {
      // Convert to CSV format (simplified)
      const csv = this.convertToCSV(data);
      return new Blob([csv], { type: 'text/csv' });
    }
    
    return new Blob([JSON.stringify(data, null, 2)], { 
      type: 'application/json' 
    });
  },

  // Helper function to convert data to CSV
  convertToCSV: (data) => {
    // This is a simplified CSV conversion
    // In a real application, you'd want a more robust CSV library
    const rows = [];
    rows.push(['Metric', 'Value']);
    
    if (data.overall_stats) {
      rows.push(['Total Analyzed', data.overall_stats.total_analyzed]);
      rows.push(['Overall Sentiment', data.overall_stats.overall_sentiment]);
    }
    
    return rows.map(row => row.join(',')).join('\n');
  },
};