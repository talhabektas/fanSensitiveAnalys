import React, { useState, useEffect } from 'react';
import { useQuery } from 'react-query';
import { 
  TrendingUp, 
  TrendingDown, 
  MessageSquare, 
  Users, 
  Activity,
  RefreshCw,
  AlertCircle
} from 'lucide-react';
import { SentimentChart } from './SentimentChart';
import { useSentiment } from '../hooks/useSentiment';
import { dashboardService } from '../services/dashboardService';
import { commentService } from '../services/commentService';
import { redditLiveService } from '../services/api';
import { TeamLogo } from '../utils/teamLogos.jsx';
import toast from 'react-hot-toast';

export const Dashboard = () => {
  const [isRefreshing, setIsRefreshing] = useState(false);
  const { 
    stats, 
    sentimentBreakdown, 
    overallSentimentScore, 
    sentimentTrend,
    isLoading: sentimentLoading 
  } = useSentiment();

  // Dashboard data query
  const { 
    data: dashboardData, 
    isLoading: dashboardLoading, 
    refetch: refetchDashboard 
  } = useQuery(
    'dashboard-data',
    dashboardService.getDashboardData,
    {
      staleTime: 30000, // 30 seconds
      refetchInterval: 60000, // Refresh every minute
      onError: (error) => {
        console.error('Error fetching dashboard data:', error);
      },
    }
  );

  // Comment stats query
  const { 
    data: commentStats, 
    isLoading: commentStatsLoading 
  } = useQuery(
    'comment-stats',
    commentService.getCommentStats,
    {
      staleTime: 30000,
      onError: (error) => {
        console.error('Error fetching comment stats:', error);
      },
    }
  );

  // Team comparison query
  const { 
    data: teamComparison, 
    isLoading: teamComparisonLoading 
  } = useQuery(
    'team-comparison',
    dashboardService.getTeamComparison,
    {
      staleTime: 300000, // 5 minutes
      onError: (error) => {
        console.error('Error fetching team comparison:', error);
      },
    }
  );

  // Manual refresh function with data collection
  const handleRefresh = async () => {
    setIsRefreshing(true);
    try {
      toast('Veri toplama ve güncelleme başlatıldı...');
      
      // Sırayla çalıştır ki hangi adımda hata aldığımızı görebilelim
      try {
        await refetchDashboard();
        toast('Dashboard verileri güncellendi');
      } catch (error) {
        console.error('Dashboard fetch error:', error);
        toast.error('Dashboard güncellenirken hata: ' + error.message);
      }

      try {
        await collectRedditData();
        toast('Reddit verileri toplandı');
      } catch (error) {
        console.error('Reddit collection error:', error);
        toast.error('Reddit verileri toplanırken hata: ' + error.message);
      }

      try {
        await collectYouTubeData();
        toast('YouTube verileri toplandı');
      } catch (error) {
        console.error('YouTube collection error:', error);
        toast.error('YouTube verileri toplanırken hata: ' + error.message);
      }

      try {
        await redditLiveService.restartStream();
        toast('Reddit stream yenilendi');
      } catch (error) {
        console.error('Reddit stream restart error:', error);
        toast.error('Reddit stream yenilenirken hata: ' + error.message);
      }

      toast.success('Tüm işlemler tamamlandı!');
    } catch (error) {
      console.error('General refresh error:', error);
      toast.error('Genel hata: ' + error.message);
    } finally {
      setIsRefreshing(false);
    }
  };

  // Reddit veri toplama fonksiyonu
  const collectRedditData = async () => {
    try {
      const response = await fetch(`${import.meta.env.VITE_API_URL}/reddit/collect`, {
        method: 'POST',
        headers: {
          'X-API-Key': import.meta.env.VITE_API_KEY,
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        throw new Error(`Reddit API error: ${response.status}`);
      }
      
      const result = await response.json();
      console.log('Reddit data collected:', result);
      return result;
    } catch (error) {
      console.error('Error collecting Reddit data:', error);
      throw error;
    }
  };

  // YouTube veri toplama fonksiyonu
  const collectYouTubeData = async () => {
    try {
      const response = await fetch(`${import.meta.env.VITE_API_URL}/youtube/collect`, {
        method: 'POST',
        headers: {
          'X-API-Key': import.meta.env.VITE_API_KEY,
          'Content-Type': 'application/json',
        },
      });
      
      if (!response.ok) {
        throw new Error(`YouTube API error: ${response.status}`);
      }
      
      const result = await response.json();
      console.log('YouTube data collected:', result);
      return result;
    } catch (error) {
      console.error('Error collecting YouTube data:', error);
      throw error;
    }
  };

  // Auto-refresh every 5 minutes
  useEffect(() => {
    const interval = setInterval(() => {
      refetchDashboard();
    }, 300000); // 5 minutes

    return () => clearInterval(interval);
  }, [refetchDashboard]);

  const isLoading = sentimentLoading || dashboardLoading || commentStatsLoading;

  // Stats cards data
  const statsCards = [
    {
      title: 'Toplam Yorum',
      value: commentStats?.total_comments || 0,
      change: '+12%',
      trend: 'up',
      icon: MessageSquare,
      color: 'blue',
    },
    {
      title: 'İşlenen Yorum',
      value: commentStats?.processed_comments || 0,
      change: '+8%',
      trend: 'up',
      icon: Activity,
      color: 'green',
    },
    {
      title: 'Duygu Skoru',
      value: overallSentimentScore || 50,
      change: sentimentTrend === 'improving' ? '+5%' : sentimentTrend === 'declining' ? '-3%' : '0%',
      trend: sentimentTrend === 'improving' ? 'up' : sentimentTrend === 'declining' ? 'down' : 'stable',
      icon: TrendingUp,
      color: overallSentimentScore >= 70 ? 'green' : overallSentimentScore >= 40 ? 'yellow' : 'red',
      isPercentage: true,
    },
    {
      title: 'Aktif Takım',
      value: teamComparison?.teams?.length || 0,
      change: '0%',
      trend: 'stable',
      icon: Users,
      color: 'purple',
    },
  ];

  // Chart data for sentiment distribution
  const sentimentChartData = {
    labels: ['Pozitif', 'Nötr', 'Negatif'],
    datasets: [
      {
        data: [
          sentimentBreakdown?.positive.count || 0,
          sentimentBreakdown?.neutral.count || 0,
          sentimentBreakdown?.negative.count || 0,
        ],
        backgroundColor: ['#10B981', '#6B7280', '#EF4444'],
        borderColor: ['#059669', '#4B5563', '#DC2626'],
        borderWidth: 2,
      },
    ],
  };

  // Recent trends data
  const trendsData = stats?.recent_trends || [];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Dashboard</h2>
          <p className="text-gray-600">
            Taraftar duygu analizi sistemi genel görünümü
          </p>
        </div>
        <button
          onClick={handleRefresh}
          disabled={isRefreshing}
          className="btn btn-primary flex items-center space-x-2"
        >
          <RefreshCw className={`h-4 w-4 ${isRefreshing ? 'animate-spin' : ''}`} />
          <span>{isRefreshing ? 'Veri Topluyor...' : 'Veri Topla & Yenile'}</span>
        </button>
      </div>

      {/* Stats Cards */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
        {statsCards.map((card) => {
          const Icon = card.icon;
          return (
            <div key={card.title} className="stat-card">
              <div className="flex items-center justify-between">
                <div>
                  <p className="stat-title">{card.title}</p>
                  <p className="stat-value">
                    {card.isPercentage 
                      ? `${card.value}%` 
                      : card.value.toLocaleString()
                    }
                  </p>
                  <div 
                    className={`stat-change ${
                      card.trend === 'up' 
                        ? 'stat-change-positive' 
                        : card.trend === 'down'
                        ? 'stat-change-negative'
                        : 'text-gray-500'
                    }`}
                  >
                    {card.trend === 'up' && <TrendingUp className="h-4 w-4 mr-1" />}
                    {card.trend === 'down' && <TrendingDown className="h-4 w-4 mr-1" />}
                    <span>{card.change}</span>
                  </div>
                </div>
                <div className={`p-3 rounded-lg bg-${card.color}-100`}>
                  <Icon className={`h-6 w-6 text-${card.color}-600`} />
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Sentiment Distribution */}
        <div className="card">
          <div className="card-header">
            <h3 className="text-lg font-medium text-gray-900">
              Duygu Dağılımı
            </h3>
          </div>
          <div className="card-content">
            {sentimentBreakdown ? (
              <SentimentChart 
                data={sentimentChartData}
                type="pie"
                height={250}
              />
            ) : (
              <div className="flex items-center justify-center h-64">
                <div className="text-center">
                  <AlertCircle className="h-12 w-12 text-gray-400 mx-auto mb-2" />
                  <p className="text-gray-500">Henüz duygu analizi verisi yok</p>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Sentiment Trends */}
        <div className="card">
          <div className="card-header">
            <h3 className="text-lg font-medium text-gray-900">
              Duygu Trendleri (7 Gün)
            </h3>
          </div>
          <div className="card-content">
            {trendsData.length > 0 ? (
              <SentimentChart
                data={{
                  labels: trendsData.map(d => new Date(d.date).toLocaleDateString('tr-TR', { day: 'numeric', month: 'short' })),
                  datasets: [
                    {
                      label: 'Pozitif',
                      data: trendsData.map(d => d.positive),
                      borderColor: '#10B981',
                      backgroundColor: 'rgba(16, 185, 129, 0.1)',
                    },
                    {
                      label: 'Negatif',
                      data: trendsData.map(d => d.negative),
                      borderColor: '#EF4444',
                      backgroundColor: 'rgba(239, 68, 68, 0.1)',
                    },
                    {
                      label: 'Nötr',
                      data: trendsData.map(d => d.neutral),
                      borderColor: '#6B7280',
                      backgroundColor: 'rgba(107, 114, 128, 0.1)',
                    },
                  ],
                }}
                type="line"
                height={250}
              />
            ) : (
              <div className="flex items-center justify-center h-64">
                <div className="text-center">
                  <TrendingUp className="h-12 w-12 text-gray-400 mx-auto mb-2" />
                  <p className="text-gray-500">Trend verisi yükleniyor...</p>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Team Comparison */}
      <div className="card">
        <div className="card-header">
          <h3 className="text-lg font-medium text-gray-900">
            Takım Karşılaştırması
          </h3>
        </div>
        <div className="card-content">
          {teamComparisonLoading ? (
            <div className="space-y-3">
              {[...Array(4)].map((_, i) => (
                <div key={i} className="animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                  <div className="h-2 bg-gray-100 rounded"></div>
                </div>
              ))}
            </div>
          ) : teamComparison?.teams?.length > 0 ? (
            <div className="space-y-4">
              {teamComparison.teams.slice(0, 5).map((team, index) => (
                <div key={team.team_id} className="flex items-center justify-between p-3 rounded-lg border border-gray-200">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center space-x-2">
                      <div className={`w-6 h-6 rounded-full flex items-center justify-center text-white font-bold text-xs ${
                        index === 0 ? 'bg-yellow-500' :
                        index === 1 ? 'bg-gray-400' :
                        index === 2 ? 'bg-amber-600' :
                        'bg-gray-300'
                      }`}>
                        {index + 1}
                      </div>
                      <TeamLogo teamName={team.team_name} size={32} />
                    </div>
                    <div>
                      <h4 className="font-medium text-gray-900">{team.team_name}</h4>
                      <p className="text-sm text-gray-500">
                        {team.total_comments} yorum
                      </p>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                      team.avg_sentiment >= 0.7 
                        ? 'bg-green-100 text-green-800'
                        : team.avg_sentiment >= 0.4
                        ? 'bg-yellow-100 text-yellow-800'
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {(team.avg_sentiment * 100).toFixed(0)}%
                    </div>
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center py-8">
              <Users className="h-12 w-12 text-gray-400 mx-auto mb-2" />
              <p className="text-gray-500">Takım karşılaştırma verisi yükleniyor...</p>
            </div>
          )}
        </div>
      </div>

      {/* Loading overlay */}
      {isLoading && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-30 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 flex items-center space-x-3">
            <div className="loading-spinner"></div>
            <span className="text-gray-900">Veriler yükleniyor...</span>
          </div>
        </div>
      )}
    </div>
  );
};