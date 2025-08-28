import React, { useState, useEffect } from 'react';
import { groqService } from '../services/groqService';
import { 
  ArrowTrendingUpIcon, 
  ClockIcon,
  LightBulbIcon,
  ArrowPathIcon,
  ExclamationTriangleIcon,
  FireIcon,
  ArrowTrendingDownIcon
} from '@heroicons/react/24/outline';

const TrendInsights = ({ teamId, teamName = "Genel" }) => {
  const [insights, setInsights] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [refreshing, setRefreshing] = useState(false);

  useEffect(() => {
    loadTrendInsights();
  }, [teamId]);

  const loadTrendInsights = async () => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await groqService.getTrendInsights(teamId);
      setInsights(response || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const refreshInsights = async () => {
    try {
      setRefreshing(true);
      const response = await groqService.getTrendInsights(teamId);
      setInsights(response || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setRefreshing(false);
    }
  };

  const getTrendIcon = (trendType) => {
    switch (trendType) {
      case 'positive':
        return <ArrowTrendingUpIcon className="h-5 w-5 text-green-500" />;
      case 'negative':
        return <ArrowTrendingDownIcon className="h-5 w-5 text-red-500" />;
      case 'topic':
        return <FireIcon className="h-5 w-5 text-orange-500" />;
      default:
        return <LightBulbIcon className="h-5 w-5 text-blue-500" />;
    }
  };

  const getTrendColor = (trendType) => {
    switch (trendType) {
      case 'positive':
        return 'border-green-200 bg-green-50';
      case 'negative':
        return 'border-red-200 bg-red-50';
      case 'topic':
        return 'border-orange-200 bg-orange-50';
      default:
        return 'border-blue-200 bg-blue-50';
    }
  };

  const formatDate = (dateStr) => {
    return new Intl.DateTimeFormat('tr-TR', {
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit'
    }).format(new Date(dateStr));
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-lg p-6">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded mb-4"></div>
          <div className="space-y-3">
            <div className="h-20 bg-gray-200 rounded"></div>
            <div className="h-20 bg-gray-200 rounded"></div>
            <div className="h-20 bg-gray-200 rounded"></div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center">
          <LightBulbIcon className="h-6 w-6 text-yellow-500 mr-2" />
          <h3 className="text-xl font-semibold text-gray-800">AI Trend İçgörüleri</h3>
          <span className="ml-2 px-2 py-1 bg-yellow-100 text-yellow-800 text-xs rounded-full">
            Son 7 Gün
          </span>
        </div>

        <button
          onClick={refreshInsights}
          disabled={refreshing}
          className={`flex items-center px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
            refreshing
              ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
              : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
          }`}
        >
          <ArrowPathIcon className={`h-4 w-4 mr-1 ${refreshing ? 'animate-spin' : ''}`} />
          Yenile
        </button>
      </div>

      {/* Team Info */}
      <div className="mb-4 p-3 bg-gray-50 rounded-lg">
        <div className="text-sm text-gray-600">
          <strong>Analiz Kapsamı:</strong> {teamName}
        </div>
      </div>

      {/* Error State */}
      {error && (
        <div className="text-center py-8">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-400 mx-auto mb-2" />
          <p className="text-red-600 mb-4">{error}</p>
          <button
            onClick={loadTrendInsights}
            className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Tekrar Dene
          </button>
        </div>
      )}

      {/* Insights List */}
      {insights.length > 0 ? (
        <div className="space-y-4">
          {insights.map((insight, index) => (
            <div
              key={insight.id || index}
              className={`border rounded-lg p-4 transition-all hover:shadow-md ${getTrendColor(insight.trend_type)}`}
            >
              {/* Insight Header */}
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center">
                  {getTrendIcon(insight.trend_type)}
                  <h4 className="ml-2 font-medium text-gray-800">
                    {insight.title}
                  </h4>
                </div>
                
                <div className="flex items-center text-xs text-gray-500">
                  <ClockIcon className="h-3 w-3 mr-1" />
                  {formatDate(insight.created_at)}
                </div>
              </div>

              {/* Insight Description */}
              <div className="mb-3">
                <p className="text-gray-700 text-sm leading-relaxed whitespace-pre-wrap">
                  {insight.description}
                </p>
              </div>

              {/* Insight Footer */}
              <div className="flex items-center justify-between">
                {/* Keywords */}
                {insight.keywords && insight.keywords.length > 0 && (
                  <div className="flex flex-wrap gap-1">
                    {insight.keywords.slice(0, 4).map((keyword, idx) => (
                      <span
                        key={idx}
                        className="px-2 py-1 bg-white bg-opacity-70 text-gray-700 text-xs rounded"
                      >
                        {keyword}
                      </span>
                    ))}
                  </div>
                )}

                {/* Confidence */}
                <div className="flex items-center text-xs text-gray-600">
                  <span className="mr-1">Güven:</span>
                  <div className={`px-2 py-1 rounded ${
                    insight.confidence >= 0.8 
                      ? 'bg-green-100 text-green-800' 
                      : insight.confidence >= 0.6
                      ? 'bg-yellow-100 text-yellow-800'
                      : 'bg-red-100 text-red-800'
                  }`}>
                    {(insight.confidence * 100).toFixed(0)}%
                  </div>
                </div>
              </div>
            </div>
          ))}
        </div>
      ) : !loading && !error && (
        /* Empty State */
        <div className="text-center py-12">
          <LightBulbIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
          <h4 className="text-lg font-medium text-gray-600 mb-2">
            Henüz Trend İçgörüsü Yok
          </h4>
          <p className="text-gray-500 mb-4">
            Bu dönem için henüz yeterli veri birikimi olmadığı için trend analizi yapılamıyor
          </p>
          <button
            onClick={loadTrendInsights}
            className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Kontrol Et
          </button>
        </div>
      )}

      {/* Info Footer */}
      {insights.length > 0 && (
        <div className="mt-6 pt-4 border-t border-gray-200">
          <div className="flex items-center justify-between text-xs text-gray-500">
            <div>
              🤖 Grok AI tarafından analiz edildi
            </div>
            <div>
              {insights.length} içgörü bulundu
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default TrendInsights;