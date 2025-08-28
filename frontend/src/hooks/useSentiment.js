import { useState, useCallback, useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import { sentimentService } from '../services/sentimentService';
import toast from 'react-hot-toast';

export const useSentiment = () => {
  const [analysisHistory, setAnalysisHistory] = useState([]);
  const queryClient = useQueryClient();

  // Fetch sentiment statistics
  const {
    data: stats,
    isLoading: statsLoading,
    error: statsError,
    refetch: refetchStats,
  } = useQuery(
    'sentiment-stats',
    sentimentService.getSentimentStats,
    {
      staleTime: 60000, // 1 minute
      onError: (error) => {
        console.error('Error fetching sentiment stats:', error);
      },
    }
  );

  // Fetch sentiment summary for dashboard
  const {
    data: summary,
    isLoading: summaryLoading,
    error: summaryError,
  } = useQuery(
    'sentiment-summary',
    sentimentService.getSentimentSummary,
    {
      staleTime: 30000, // 30 seconds
      onError: (error) => {
        console.error('Error fetching sentiment summary:', error);
      },
    }
  );

  // Analyze text mutation
  const analyzeTextMutation = useMutation(
    ({ text, language }) => sentimentService.analyzeText(text, language),
    {
      onSuccess: (data, variables) => {
        toast.success('Duygu analizi tamamlandı');
        // Add to analysis history
        setAnalysisHistory(prev => [{
          id: Date.now(),
          text: variables.text,
          result: data.result,
          timestamp: new Date(),
        }, ...prev.slice(0, 9)]); // Keep last 10 analyses
        
        queryClient.invalidateQueries('sentiment-stats');
      },
      onError: (error) => {
        console.error('Error analyzing text:', error);
        toast.error('Duygu analizi sırasında hata oluştu');
      },
    }
  );

  // Batch analyze mutation
  const analyzeBatchMutation = useMutation(
    sentimentService.analyzeBatch,
    {
      onSuccess: (data) => {
        const successCount = data.results.filter(result => result !== null).length;
        toast.success(`${successCount}/${data.total_texts} metin başarıyla analiz edildi`);
        queryClient.invalidateQueries('sentiment-stats');
      },
      onError: (error) => {
        console.error('Error in batch analysis:', error);
        toast.error('Toplu analiz sırasında hata oluştu');
      },
    }
  );

  // Process unprocessed comments mutation
  const processCommentsMutation = useMutation(
    sentimentService.processUnprocessedComments,
    {
      onSuccess: (data) => {
        if (data.processed > 0) {
          toast.success(`${data.processed} yorum işlendi`);
          queryClient.invalidateQueries('sentiment-stats');
          queryClient.invalidateQueries('comment-stats');
        } else {
          toast.info('İşlenecek yeni yorum bulunamadı');
        }
      },
      onError: (error) => {
        console.error('Error processing comments:', error);
        toast.error('Yorumlar işlenirken hata oluştu');
      },
    }
  );

  // Actions
  const analyzeText = useCallback((text, language = 'tr') => {
    if (!text?.trim()) {
      toast.error('Lütfen analiz edilecek metni girin');
      return;
    }
    
    if (text.length < 5) {
      toast.error('Metin en az 5 karakter olmalıdır');
      return;
    }
    
    return analyzeTextMutation.mutate({ text, language });
  }, [analyzeTextMutation]);

  const analyzeBatch = useCallback((texts) => {
    if (!texts || texts.length === 0) {
      toast.error('Analiz edilecek metin bulunamadı');
      return;
    }
    
    if (texts.length > 50) {
      toast.error('Bir seferde en fazla 50 metin analiz edilebilir');
      return;
    }
    
    return analyzeBatchMutation.mutate(texts);
  }, [analyzeBatchMutation]);

  const processUnprocessedComments = useCallback(() => {
    return processCommentsMutation.mutate();
  }, [processCommentsMutation]);

  const clearAnalysisHistory = useCallback(() => {
    setAnalysisHistory([]);
  }, []);

  // Team-specific sentiment hooks
  const useTeamSentiment = (teamId, startDate = null, endDate = null) => {
    return useQuery(
      ['team-sentiment', teamId, startDate, endDate],
      () => sentimentService.getTeamReport(teamId, startDate, endDate),
      {
        enabled: !!teamId,
        staleTime: 300000, // 5 minutes
        onError: (error) => {
          console.error(`Error fetching sentiment for team ${teamId}:`, error);
        },
      }
    );
  };

  // Sentiment trends
  const useSentimentTrends = (days = 30) => {
    return useQuery(
      ['sentiment-trends', days],
      () => sentimentService.getSentimentTrends(days),
      {
        staleTime: 300000, // 5 minutes
        onError: (error) => {
          console.error('Error fetching sentiment trends:', error);
        },
      }
    );
  };

  // Team comparison
  const useTeamComparison = () => {
    return useQuery(
      'team-sentiment-breakdown',
      sentimentService.getTeamSentimentBreakdown,
      {
        staleTime: 300000, // 5 minutes
        onError: (error) => {
          console.error('Error fetching team comparison:', error);
        },
      }
    );
  };

  // Computed values
  const sentimentBreakdown = useMemo(() => {
    if (!stats?.sentiment_breakdown) return null;
    
    const breakdown = stats.sentiment_breakdown;
    const total = Object.values(breakdown).reduce((sum, count) => sum + count, 0);
    
    if (total === 0) return null;
    
    return {
      positive: {
        count: breakdown.POSITIVE || 0,
        percentage: ((breakdown.POSITIVE || 0) / total * 100).toFixed(1),
      },
      negative: {
        count: breakdown.NEGATIVE || 0,
        percentage: ((breakdown.NEGATIVE || 0) / total * 100).toFixed(1),
      },
      neutral: {
        count: breakdown.NEUTRAL || 0,
        percentage: ((breakdown.NEUTRAL || 0) / total * 100).toFixed(1),
      },
      total,
    };
  }, [stats?.sentiment_breakdown]);

  const overallSentimentScore = useMemo(() => {
    if (!sentimentBreakdown) return 50;
    
    const { positive, negative, neutral, total } = sentimentBreakdown;
    
    // Calculate weighted score (positive=100, neutral=50, negative=0)
    const score = (
      (positive.count * 100) + 
      (neutral.count * 50) + 
      (negative.count * 0)
    ) / total;
    
    return Math.round(score);
  }, [sentimentBreakdown]);

  const sentimentTrend = useMemo(() => {
    if (!stats?.recent_trends || stats.recent_trends.length < 2) return 'stable';
    
    const trends = stats.recent_trends;
    const recent = trends[trends.length - 1];
    const previous = trends[trends.length - 2];
    
    if (recent.score > previous.score + 5) return 'improving';
    if (recent.score < previous.score - 5) return 'declining';
    return 'stable';
  }, [stats?.recent_trends]);

  // Loading states
  const isAnalyzing = analyzeTextMutation.isLoading;
  const isBatchAnalyzing = analyzeBatchMutation.isLoading;
  const isProcessing = processCommentsMutation.isLoading;
  const isLoading = statsLoading || summaryLoading;

  // Error states
  const hasError = statsError || summaryError;

  // Utility functions
  const getSentimentColor = useCallback((sentiment) => {
    switch (sentiment?.toUpperCase()) {
      case 'POSITIVE':
        return '#10B981'; // Green
      case 'NEGATIVE':
        return '#EF4444'; // Red
      case 'NEUTRAL':
        return '#6B7280'; // Gray
      default:
        return '#6B7280';
    }
  }, []);

  const getSentimentIcon = useCallback((sentiment) => {
    switch (sentiment?.toUpperCase()) {
      case 'POSITIVE':
        return '😊';
      case 'NEGATIVE':
        return '😞';
      case 'NEUTRAL':
        return '😐';
      default:
        return '❓';
    }
  }, []);

  const formatSentimentLabel = useCallback((sentiment) => {
    switch (sentiment?.toUpperCase()) {
      case 'POSITIVE':
        return 'Pozitif';
      case 'NEGATIVE':
        return 'Negatif';
      case 'NEUTRAL':
        return 'Nötr';
      default:
        return 'Bilinmiyor';
    }
  }, []);

  return {
    // Data
    stats,
    summary,
    sentimentBreakdown,
    overallSentimentScore,
    sentimentTrend,
    analysisHistory,
    
    // Loading states
    isLoading,
    statsLoading,
    summaryLoading,
    isAnalyzing,
    isBatchAnalyzing,
    isProcessing,
    
    // Error states
    statsError,
    summaryError,
    hasError,
    
    // Actions
    analyzeText,
    analyzeBatch,
    processUnprocessedComments,
    clearAnalysisHistory,
    refetchStats,
    
    // Hooks for specific use cases
    useTeamSentiment,
    useSentimentTrends,
    useTeamComparison,
    
    // Utility functions
    getSentimentColor,
    getSentimentIcon,
    formatSentimentLabel,
  };
};