import React, { useState, useEffect } from 'react';
import { Line, Bar, Doughnut } from 'react-chartjs-2';
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js';
import { 
  TrendingUp, 
  TrendingDown, 
  Activity, 
  BarChart3,
  Calendar,
  Users,
  MessageSquare,
  Award,
  AlertTriangle,
  RefreshCw,
  Download,
  Lightbulb
} from 'lucide-react';
import { TeamLogo } from '../utils/teamLogos.jsx';
import { apiClient } from '../services/api';
import toast from 'react-hot-toast';

// Chart.js kayıt
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  BarElement,
  ArcElement,
  Title,
  Tooltip,
  Legend,
  Filler
);

export const TrendAnalysis = () => {
  const [trendData, setTrendData] = useState(null);
  const [insights, setInsights] = useState([]);
  const [selectedPeriod, setSelectedPeriod] = useState('7d');
  const [isLoading, setIsLoading] = useState(true);
  const [activeTab, setActiveTab] = useState('trends');
  const [isGeneratingReport, setIsGeneratingReport] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(null);

  // Veri yükleme - Cache bypass için timestamp ekle
  const loadTrendData = async () => {
    setIsLoading(true);
    try {
      const timestamp = Date.now();
      console.log('[TrendAnalysis] Loading data with timestamp:', timestamp, 'Period:', selectedPeriod);
      
      const [trendsResponse, insightsResponse] = await Promise.all([
        apiClient.get(`/trends/analysis?period=${selectedPeriod}&t=${timestamp}`),
        apiClient.get(`/trends/insights?period=${selectedPeriod}&t=${timestamp}`)
      ]);
      
      console.log('[TrendAnalysis] Received trends data:', trendsResponse);
      console.log('[TrendAnalysis] Received insights data:', insightsResponse);
      
      setTrendData(trendsResponse);
      setInsights(insightsResponse.insights || []);
      setLastUpdated(new Date());
      toast.success(`Trend analizi güncellendi - ${new Date().toLocaleTimeString('tr-TR')}`);
    } catch (error) {
      console.error('Trend data loading error:', error);
      toast.error('Trend verileri yüklenirken hata oluştu');
    } finally {
      setIsLoading(false);
    }
  };

  // Rapor indirme fonksiyonu
  const downloadReport = async (format = 'html') => {
    setIsGeneratingReport(true);
    try {
      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/reports/executive/download?period=${selectedPeriod}&format=${format}`,
        {
          method: 'GET',
          headers: {
            'X-API-Key': import.meta.env.VITE_API_KEY,
          },
        }
      );

      if (!response.ok) {
        throw new Error('Rapor oluşturulamadı');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.style.display = 'none';
      a.href = url;
      a.download = `executive-report-${selectedPeriod}.${format === 'json' ? 'json' : 'html'}`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast.success(`${format.toUpperCase()} raporu başarıyla indirildi`);
    } catch (error) {
      console.error('Report download error:', error);
      toast.error('Rapor indirilemedi: ' + error.message);
    } finally {
      setIsGeneratingReport(false);
    }
  };

  useEffect(() => {
    loadTrendData();
  }, [selectedPeriod]);

  // Otomatik yenileme - her 30 saniyede bir
  useEffect(() => {
    const interval = setInterval(() => {
      loadTrendData();
    }, 30000); // 30 saniye

    return () => clearInterval(interval);
  }, [selectedPeriod]);

  // Chart.js konfigürasyonları
  const chartOptions = {
    responsive: true,
    plugins: {
      legend: {
        position: 'top',
      },
      tooltip: {
        mode: 'index',
        intersect: false,
      },
    },
    scales: {
      x: {
        display: true,
        title: {
          display: true,
          text: 'Tarih'
        }
      },
      y: {
        display: true,
        title: {
          display: true,
          text: 'Yorum Sayısı'
        }
      }
    },
    interaction: {
      mode: 'nearest',
      axis: 'x',
      intersect: false
    },
  };

  // Günlük yorum sayısı grafiği
  const getCommentsChart = () => {
    if (!trendData?.teams) {
      console.log('[TrendAnalysis] No trend data available for chart');
      return null;
    }

    console.log('[TrendAnalysis] Generating chart for', trendData.teams.length, 'teams');
    
    const dates = trendData.teams[0]?.data?.map(d => 
      new Date(d.date).toLocaleDateString('tr-TR', { month: 'short', day: 'numeric' })
    ) || [];

    const datasets = trendData.teams.map((team, index) => {
      const colors = ['#f59e0b', '#3b82f6', '#10b981', '#ef4444'];
      const dataPoints = team.data.map(d => d.total);
      console.log(`[TrendAnalysis] Team ${team.team_name}: ${dataPoints.length} data points, total comments: ${dataPoints.reduce((a, b) => a + b, 0)}`);
      
      return {
        label: team.team_name,
        data: dataPoints,
        borderColor: colors[index % colors.length],
        backgroundColor: colors[index % colors.length] + '20',
        fill: true,
        tension: 0.4,
      };
    });

    return {
      labels: dates,
      datasets: datasets
    };
  };

  // Takım karşılaştırma grafiği
  const getTeamComparisonChart = () => {
    if (!trendData?.teams) {
      console.log('[TrendAnalysis] No team data available for comparison chart');
      return null;
    }

    const teamData = trendData.teams.map(t => ({
      name: t.team_name,
      comments: t.overall.total_comments
    }));
    
    console.log('[TrendAnalysis] Team comparison data:', teamData);

    return {
      labels: trendData.teams.map(t => t.team_name),
      datasets: [{
        label: 'Toplam Yorum',
        data: trendData.teams.map(t => t.overall.total_comments),
        backgroundColor: [
          '#f59e0b',
          '#3b82f6', 
          '#10b981',
          '#ef4444'
        ],
        borderWidth: 2,
        borderColor: '#ffffff'
      }]
    };
  };

  // Trend yönü ikonu
  const getTrendIcon = (direction) => {
    switch (direction) {
      case 'up':
        return <TrendingUp className="h-4 w-4 text-green-500" />;
      case 'down':
        return <TrendingDown className="h-4 w-4 text-red-500" />;
      default:
        return <Activity className="h-4 w-4 text-gray-500" />;
    }
  };

  // Insight ikonu
  const getInsightIcon = (type) => {
    switch (type) {
      case 'improvement':
        return <TrendingUp className="h-5 w-5 text-green-500" />;
      case 'decline':
        return <TrendingDown className="h-5 w-5 text-red-500" />;
      case 'spike':
        return <Award className="h-5 w-5 text-blue-500" />;
      default:
        return <Lightbulb className="h-5 w-5 text-yellow-500" />;
    }
  };

  // Insight rengi
  const getInsightColor = (severity) => {
    switch (severity) {
      case 'high':
        return 'border-red-200 bg-red-50';
      case 'medium':
        return 'border-yellow-200 bg-yellow-50';
      default:
        return 'border-blue-200 bg-blue-50';
    }
  };

  if (isLoading) {
    return (
      <div className="max-w-7xl mx-auto px-4 py-8">
        <div className="animate-pulse">
          <div className="h-8 bg-gray-200 rounded w-64 mb-8"></div>
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            {[1, 2, 3, 4].map(i => (
              <div key={i} className="bg-gray-200 rounded-lg h-32"></div>
            ))}
          </div>
          <div className="bg-gray-200 rounded-lg h-96"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="max-w-7xl mx-auto px-4 py-8">
      {/* Header */}
      <div className="mb-8 flex flex-col sm:flex-row justify-between items-start sm:items-center">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            📊 Trend Analizi
          </h1>
          <p className="text-gray-600">
            Taraftar duygularının zaman içindeki değişimini analiz edin
          </p>
          {lastUpdated && (
            <p className="text-xs text-gray-500 mt-1">
              Son güncelleme: {lastUpdated.toLocaleString('tr-TR')}
            </p>
          )}
        </div>
        
        <div className="flex items-center space-x-4 mt-4 sm:mt-0">
          {/* Period Selector */}
          <select 
            value={selectedPeriod}
            onChange={(e) => setSelectedPeriod(e.target.value)}
            className="select"
          >
            <option value="7d">Son 7 Gün</option>
            <option value="30d">Son 30 Gün</option>
            <option value="90d">Son 90 Gün</option>
          </select>

          <button
            onClick={loadTrendData}
            disabled={isLoading}
            className="btn btn-primary flex items-center space-x-2"
          >
            <RefreshCw className={`h-4 w-4 ${isLoading ? 'animate-spin' : ''}`} />
            <span>{isLoading ? 'Yenileniyor...' : 'Yenile'}</span>
          </button>

          {/* Report Download Dropdown */}
          <div className="relative inline-block text-left">
            <div className="group">
              <button
                type="button"
                className="btn btn-secondary flex items-center space-x-2"
                disabled={isGeneratingReport}
              >
                <Download className={`h-4 w-4 ${isGeneratingReport ? 'animate-spin' : ''}`} />
                <span>{isGeneratingReport ? 'Oluşturuluyor...' : 'Rapor İndir'}</span>
              </button>
              
              {!isGeneratingReport && (
                <div className="absolute right-0 z-10 mt-2 w-48 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200">
                  <div className="py-1">
                    <button
                      onClick={() => downloadReport('html')}
                      className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    >
                      <Download className="h-4 w-4 mr-2" />
                      HTML Raporu
                    </button>
                    <button
                      onClick={() => downloadReport('json')}
                      className="flex items-center w-full px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                    >
                      <Download className="h-4 w-4 mr-2" />
                      JSON Raporu
                    </button>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
        <div className="card">
          <div className="card-body">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Toplam Yorum</p>
                <p className="text-2xl font-bold text-gray-900">
                  {trendData?.summary?.total_comments?.toLocaleString() || 0}
                </p>
              </div>
              <MessageSquare className="h-8 w-8 text-blue-500" />
            </div>
          </div>
        </div>

        <div className="card">
          <div className="card-body">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Günlük Ortalama</p>
                <p className="text-2xl font-bold text-gray-900">
                  {Math.round(trendData?.summary?.average_daily || 0)}
                </p>
              </div>
              <Activity className="h-8 w-8 text-green-500" />
            </div>
          </div>
        </div>

        <div className="card">
          <div className="card-body">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">En Pozitif Takım</p>
                <p className="text-lg font-bold text-green-600">
                  {trendData?.summary?.most_positive_team || '-'}
                </p>
              </div>
              <Award className="h-8 w-8 text-yellow-500" />
            </div>
          </div>
        </div>

        <div className="card">
          <div className="card-body">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-gray-600">Analiz Dönemi</p>
                <p className="text-lg font-bold text-gray-900">
                  {selectedPeriod}
                </p>
              </div>
              <Calendar className="h-8 w-8 text-purple-500" />
            </div>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="flex space-x-8">
          <button
            onClick={() => setActiveTab('trends')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'trends'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            📈 Trend Grafikleri
          </button>
          <button
            onClick={() => setActiveTab('reports')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'reports'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            📊 Executive Raporları
          </button>
          <button
            onClick={() => setActiveTab('insights')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'insights'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            🤖 AI İçgörüler
          </button>
          <button
            onClick={() => setActiveTab('debug')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'debug'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            🔧 Debug
          </button>
        </nav>
      </div>

      {/* Content */}
      {activeTab === 'trends' ? (
        <div className="space-y-8">
          {/* Günlük Trend Grafiği */}
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-semibold">📈 Günlük Yorum Trendi</h3>
              <p className="text-sm text-gray-600">
                Son {selectedPeriod} içindeki günlük yorum sayısı değişimi
              </p>
            </div>
            <div className="card-body">
              <div className="h-96">
                {getCommentsChart() && (
                  <Line 
                    key={`line-chart-${selectedPeriod}-${Date.now()}`}
                    data={getCommentsChart()} 
                    options={{
                      ...chartOptions,
                      animation: {
                        duration: 0 // Disable animation for faster updates
                      },
                      responsive: true,
                      maintainAspectRatio: false
                    }} 
                  />
                )}
              </div>
            </div>
          </div>

          {/* Takım Karşılaştırma */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <div className="card">
              <div className="card-header">
                <h3 className="text-lg font-semibold">📊 Takım Karşılaştırma</h3>
              </div>
              <div className="card-body">
                <div className="h-64">
                  {getTeamComparisonChart() && (
                    <Doughnut 
                      key={`doughnut-chart-${selectedPeriod}-${Date.now()}`}
                      data={getTeamComparisonChart()} 
                      options={{
                        responsive: true,
                        maintainAspectRatio: false,
                        animation: {
                          duration: 0 // Disable animation for faster updates
                        },
                        plugins: {
                          legend: {
                            position: 'bottom',
                          }
                        }
                      }} 
                    />
                  )}
                </div>
              </div>
            </div>

            {/* Takım Detayları */}
            <div className="card">
              <div className="card-header">
                <h3 className="text-lg font-semibold">🏆 Takım Performansı</h3>
              </div>
              <div className="card-body">
                <div className="space-y-3">
                  {trendData?.teams?.map((team) => (
                    <div key={team.team_id} className="flex items-center justify-between p-2 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
                      <div className="flex items-center space-x-2">
                        <div className="w-6 h-6 flex-shrink-0">
                          <TeamLogo teamName={team.team_name} size={16} />
                        </div>
                        <div>
                          <p className="font-medium text-sm">{team.team_name}</p>
                          <p className="text-xs text-gray-500">
                            {team.overall.total_comments} yorum
                          </p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-1">
                        {getTrendIcon(team.overall.trend_direction)}
                        <span className={`text-xs font-medium ${
                          team.overall.weekly_change > 0 
                            ? 'text-green-600' 
                            : team.overall.weekly_change < 0 
                            ? 'text-red-600' 
                            : 'text-gray-600'
                        }`}>
                          {team.overall.weekly_change > 0 ? '+' : ''}
                          {team.overall.weekly_change.toFixed(1)}%
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : activeTab === 'reports' ? (
        <div className="space-y-6">
          {/* Executive Reports */}
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-semibold flex items-center space-x-2">
                <Download className="h-5 w-5 text-blue-500" />
                <span>📊 Executive Raporları</span>
              </h3>
              <p className="text-sm text-gray-600">
                Yöneticiler için detaylı analiz raporları oluşturun ve indirin
              </p>
            </div>
            <div className="card-body">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="space-y-4">
                  <h4 className="font-medium text-gray-900">Rapor Türleri</h4>
                  
                  <div className="space-y-3">
                    <div className="border rounded-lg p-4 hover:bg-gray-50 transition-colors">
                      <h5 className="font-medium text-gray-900 flex items-center space-x-2">
                        <div className="w-2 h-2 bg-blue-500 rounded-full"></div>
                        <span>HTML Executive Raporu</span>
                      </h5>
                      <p className="text-sm text-gray-600 mt-1">
                        Görsel grafikler ve detaylı analizlerle profesyonel rapor
                      </p>
                      <button
                        onClick={() => downloadReport('html')}
                        disabled={isGeneratingReport}
                        className="mt-2 btn btn-primary btn-sm"
                      >
                        {isGeneratingReport ? 'Oluşturuluyor...' : 'HTML İndir'}
                      </button>
                    </div>

                    <div className="border rounded-lg p-4 hover:bg-gray-50 transition-colors">
                      <h5 className="font-medium text-gray-900 flex items-center space-x-2">
                        <div className="w-2 h-2 bg-green-500 rounded-full"></div>
                        <span>JSON Veri Raporu</span>
                      </h5>
                      <p className="text-sm text-gray-600 mt-1">
                        API entegrasyonu ve programatik kullanım için yapılandırılmış veri
                      </p>
                      <button
                        onClick={() => downloadReport('json')}
                        disabled={isGeneratingReport}
                        className="mt-2 btn btn-secondary btn-sm"
                      >
                        {isGeneratingReport ? 'Oluşturuluyor...' : 'JSON İndir'}
                      </button>
                    </div>
                  </div>
                </div>

                <div className="space-y-4">
                  <h4 className="font-medium text-gray-900">Rapor İçeriği</h4>
                  <div className="space-y-2 text-sm text-gray-600">
                    <div className="flex items-center space-x-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full"></div>
                      <span>Executive Summary - Genel özetler ve metrikler</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full"></div>
                      <span>Key Findings - Ana bulgular ve içgörüler</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full"></div>
                      <span>Team Analysis - Takım bazlı karşılaştırmalar</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full"></div>
                      <span>Recommendations - Actionable öneriler</span>
                    </div>
                    <div className="flex items-center space-x-2">
                      <div className="w-1.5 h-1.5 bg-gray-400 rounded-full"></div>
                      <span>Data Sources - Veri kaynakları ve güvenilirlik</span>
                    </div>
                  </div>
                  
                  <div className="mt-4 p-3 bg-blue-50 rounded-lg">
                    <div className="flex items-center space-x-2">
                      <Lightbulb className="h-4 w-4 text-blue-600" />
                      <span className="text-sm font-medium text-blue-900">LinkedIn İpucu</span>
                    </div>
                    <p className="text-sm text-blue-700 mt-1">
                      Bu raporları LinkedIn profilinizde paylaşarak veri analizi yetkinliklerinizi sergileyin
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : activeTab === 'debug' ? (
        <div className="space-y-6">
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-semibold flex items-center space-x-2">
                <span>🔧 Data Debug Information</span>
              </h3>
            </div>
            <div className="card-body">
              <div className="space-y-4">
                <div>
                  <h4 className="font-medium text-gray-900 mb-2">Current Period: {selectedPeriod}</h4>
                  <p className="text-sm text-gray-600">Last Updated: {lastUpdated ? lastUpdated.toLocaleString('tr-TR') : 'Never'}</p>
                </div>
                
                <div>
                  <h4 className="font-medium text-gray-900 mb-2">Raw Trend Data:</h4>
                  <pre className="bg-gray-100 p-3 rounded text-xs overflow-auto max-h-96">
                    {JSON.stringify(trendData, null, 2)}
                  </pre>
                </div>
                
                <div>
                  <h4 className="font-medium text-gray-900 mb-2">Raw Insights Data:</h4>
                  <pre className="bg-gray-100 p-3 rounded text-xs overflow-auto max-h-96">
                    {JSON.stringify(insights, null, 2)}
                  </pre>
                </div>
                
                <div>
                  <h4 className="font-medium text-gray-900 mb-2">Data Summary:</h4>
                  <ul className="text-sm text-gray-600 space-y-1">
                    <li>Teams Count: {trendData?.teams?.length || 0}</li>
                    <li>Total Comments: {trendData?.summary?.total_comments || 0}</li>
                    <li>Period: {trendData?.period || 'N/A'}</li>
                    <li>Start Date: {trendData?.start_date || 'N/A'}</li>
                    <li>End Date: {trendData?.end_date || 'N/A'}</li>
                    <li>Insights Count: {insights?.length || 0}</li>
                  </ul>
                </div>
              </div>
            </div>
          </div>
        </div>
      ) : (
        <div className="space-y-6">
          {/* AI Insights */}
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-semibold flex items-center space-x-2">
                <Lightbulb className="h-5 w-5 text-yellow-500" />
                <span>🤖 AI İçgörüler</span>
              </h3>
              <p className="text-sm text-gray-600">
                Otomatik analiz ile önemli eğilimler ve değişiklikler
              </p>
            </div>
            <div className="card-body">
              {insights.length > 0 ? (
                <div className="space-y-4">
                  {insights.map((insight, index) => (
                    <div key={index} className={`border rounded-lg p-4 ${getInsightColor(insight.severity)}`}>
                      <div className="flex items-start space-x-3">
                        {getInsightIcon(insight.type)}
                        <div className="flex-1">
                          <div className="flex items-center justify-between">
                            <h4 className="font-medium text-gray-900">
                              {insight.team_name}
                            </h4>
                            <span className="text-sm font-semibold text-blue-600">
                              {insight.value}
                            </span>
                          </div>
                          <p className="text-sm text-gray-700 mt-1">
                            {insight.description}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <Lightbulb className="h-12 w-12 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-lg font-medium text-gray-900 mb-2">
                    Henüz İçgörü Yok
                  </h3>
                  <p className="text-gray-600">
                    Daha fazla veri toplandığında AI içgörüler burada görünecek
                  </p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};