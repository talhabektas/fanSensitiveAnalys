import { apiClient } from './api';

export const teamService = {
  // Get all teams
  getTeams: async () => {
    return apiClient.get('/teams');
  },

  // Create a new team
  createTeam: async (teamData) => {
    return apiClient.post('/teams', teamData);
  },

  // Get a single team by ID
  getTeam: async (id) => {
    return apiClient.get(`/teams/${id}`);
  },

  // Update a team
  updateTeam: async (id, updateData) => {
    return apiClient.put(`/teams/${id}`, updateData);
  },

  // Delete a team (soft delete)
  deleteTeam: async (id) => {
    return apiClient.delete(`/teams/${id}`);
  },

  // Get team sentiment analysis
  getTeamSentiment: async (id, startDate = null, endDate = null) => {
    let url = `/teams/${id}/sentiment`;
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

  // Get team statistics
  getTeamStats: async (id) => {
    return apiClient.get(`/teams/${id}/stats`);
  },

  // Seed Turkish teams
  seedTeams: async () => {
    return apiClient.post('/teams/seed');
  },

  // Get team comparison data
  getTeamComparison: async (days = 30) => {
    const endDate = new Date().toISOString().split('T')[0];
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - days);
    const startDateStr = startDate.toISOString().split('T')[0];
    
    return apiClient.get('/dashboard/comparison');
  },

  // Get team sentiment trends
  getTeamSentimentTrends: async (teamId, days = 30) => {
    const endDate = new Date().toISOString().split('T')[0];
    const startDate = new Date();
    startDate.setDate(startDate.getDate() - days);
    const startDateStr = startDate.toISOString().split('T')[0];
    
    return teamService.getTeamSentiment(teamId, startDateStr, endDate);
  },

  // Get team keywords and their sentiment
  getTeamKeywords: async (teamId, days = 30) => {
    const report = await teamService.getTeamSentimentTrends(teamId, days);
    return report.top_keywords || [];
  },

  // Get team performance metrics
  getTeamPerformanceMetrics: async (teamId) => {
    const stats = await teamService.getTeamStats(teamId);
    const sentiment = await teamService.getTeamSentiment(teamId);
    
    return {
      ...stats,
      sentiment_details: sentiment,
    };
  },

  // Get active teams only
  getActiveTeams: async () => {
    const response = await teamService.getTeams();
    return {
      ...response,
      teams: response.teams.filter(team => team.is_active),
    };
  },

  // Update team activity status
  updateTeamStatus: async (id, isActive) => {
    return teamService.updateTeam(id, { is_active: isActive });
  },

  // Get teams with their latest sentiment data
  getTeamsWithSentiment: async () => {
    const teamsResponse = await teamService.getTeams();
    const teams = teamsResponse.teams || [];
    
    const teamsWithSentiment = await Promise.all(
      teams.map(async (team) => {
        try {
          const stats = await teamService.getTeamStats(team.id);
          return {
            ...team,
            sentiment_stats: stats,
          };
        } catch (error) {
          console.error(`Error getting sentiment for team ${team.id}:`, error);
          return {
            ...team,
            sentiment_stats: null,
          };
        }
      })
    );
    
    return {
      teams: teamsWithSentiment,
      count: teamsWithSentiment.length,
    };
  },

  // Search teams by name or keyword
  searchTeams: async (query) => {
    const response = await teamService.getTeams();
    const teams = response.teams || [];
    
    const filteredTeams = teams.filter(team => 
      team.name.toLowerCase().includes(query.toLowerCase()) ||
      team.keywords.some(keyword => 
        keyword.toLowerCase().includes(query.toLowerCase())
      )
    );
    
    return {
      teams: filteredTeams,
      count: filteredTeams.length,
    };
  },

  // Get team colors for chart visualization
  getTeamColors: (teamSlug) => {
    const teamColorMap = {
      galatasaray: { primary: '#FFA500', secondary: '#8B0000' },
      fenerbahce: { primary: '#FFFF00', secondary: '#000080' },
      besiktas: { primary: '#000000', secondary: '#FFFFFF' },
      trabzonspor: { primary: '#800080', secondary: '#000080' },
    };
    
    return teamColorMap[teamSlug] || { primary: '#6B7280', secondary: '#9CA3AF' };
  },

  // Get team logo URL (if implementing logo storage)
  getTeamLogoUrl: (teamSlug) => {
    return `/assets/team-logos/${teamSlug}.png`;
  },

  // Validate team data before submission
  validateTeamData: (teamData) => {
    const errors = [];
    
    if (!teamData.name || teamData.name.trim().length < 2) {
      errors.push('Takım adı en az 2 karakter olmalıdır');
    }
    
    if (!teamData.slug || teamData.slug.trim().length < 2) {
      errors.push('Takım kısa adı en az 2 karakter olmalıdır');
    }
    
    if (!teamData.keywords || teamData.keywords.length === 0) {
      errors.push('En az bir anahtar kelime belirtmelisiniz');
    }
    
    return errors;
  },
};