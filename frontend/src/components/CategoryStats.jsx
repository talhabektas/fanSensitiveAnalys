import React, { useState, useEffect } from 'react';
import { Doughnut, Bar } from 'react-chartjs-2';
import { groqService, groqUtils } from '../services/groqService';
import { 
  ChartBarIcon, 
  TagIcon, 
  ShieldExclamationIcon,
  InformationCircleIcon
} from '@heroicons/react/24/outline';

const CategoryStats = ({ teamId = null }) => {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);

  useEffect(() => {
    loadCategoryStats();
  }, [teamId]);

  const loadCategoryStats = async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await groqService.getCategoryStats(teamId);
      setData(response);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-white rounded-lg shadow-lg p-6">
        <div className="animate-pulse">
          <div className="h-6 bg-gray-200 rounded mb-4"></div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-white rounded-lg shadow-lg p-6">
        <div className="text-center py-8">
          <InformationCircleIcon className="h-12 w-12 text-gray-400 mx-auto mb-2" />
          <p className="text-gray-500">Kategori verileri yüklenemedi</p>
          <button 
            onClick={loadCategoryStats}
            className="mt-2 px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
          >
            Tekrar Dene
          </button>
        </div>
      </div>
    );
  }

  if (!data || !data.categories) {
    return null;
  }

  // Doughnut chart verileri
  const doughnutData = {
    labels: data.categories.map(cat => cat.category),
    datasets: [{
      data: data.categories.map(cat => cat.percentage),
      backgroundColor: data.categories.map(cat => groqUtils.getCategoryColor(cat.category)),
      borderWidth: 2,
      borderColor: '#ffffff'
    }]
  };

  // Bar chart verileri (sentiment dağılımı)
  const barData = {
    labels: data.categories.map(cat => cat.category),
    datasets: [{
      label: 'Ortalama Sentiment',
      data: data.categories.map(cat => cat.avg_sentiment),
      backgroundColor: data.categories.map(cat => {
        const sentiment = cat.avg_sentiment;
        if (sentiment > 0.6) return '#10B981'; // Pozitif - yeşil
        if (sentiment < 0.4) return '#EF4444'; // Negatif - kırmızı
        return '#F59E0B'; // Nötr - turuncu
      }),
      borderRadius: 4,
    }]
  };

  const chartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
      legend: {
        position: 'bottom',
        labels: {
          padding: 20,
          usePointStyle: true,
        }
      }
    }
  };

  const barOptions = {
    ...chartOptions,
    scales: {
      y: {
        beginAtZero: true,
        max: 1,
        ticks: {
          callback: function(value) {
            return (value * 100).toFixed(0) + '%';
          }
        }
      }
    },
    plugins: {
      ...chartOptions.plugins,
      tooltip: {
        callbacks: {
          label: function(context) {
            return `Ortalama Sentiment: ${(context.parsed.y * 100).toFixed(1)}%`;
          }
        }
      }
    }
  };

  return (
    <div className="space-y-6">
      {/* Kategori Dağılımı */}
      <div className="bg-white rounded-lg shadow-lg p-6">
        <div className="flex items-center mb-4">
          <TagIcon className="h-5 w-5 text-blue-500 mr-2" />
          <h3 className="text-lg font-semibold text-gray-800">Yorum Kategorileri</h3>
        </div>
        
        <div className="h-64 mb-4">
          <Doughnut data={doughnutData} options={chartOptions} />
        </div>

        {/* Kategori Detayları */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {data.categories.map((category, index) => (
            <div 
              key={index}
              className="p-3 rounded-lg border border-gray-200 bg-gray-50"
            >
              <div className="flex items-center justify-between mb-2">
                <span 
                  className="inline-block w-3 h-3 rounded-full mr-2"
                  style={{ backgroundColor: groqUtils.getCategoryColor(category.category) }}
                ></span>
                <span className="font-medium text-gray-800">{category.category}</span>
                <span className="text-sm text-gray-500">
                  {category.percentage}%
                </span>
              </div>
              
              <div className="text-xs text-gray-600">
                <div>{category.count} yorum</div>
                <div>
                  Sentiment: {(category.avg_sentiment * 100).toFixed(1)}%
                </div>
              </div>

              {/* Anahtar Kelimeler */}
              {category.keywords && category.keywords.length > 0 && (
                <div className="mt-2">
                  <div className="flex flex-wrap gap-1">
                    {category.keywords.slice(0, 3).map((keyword, idx) => (
                      <span 
                        key={idx}
                        className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded"
                      >
                        {keyword}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </div>

      {/* Kategori Sentiment Analizi */}
      <div className="bg-white rounded-lg shadow-lg p-6">
        <div className="flex items-center mb-4">
          <ChartBarIcon className="h-5 w-5 text-green-500 mr-2" />
          <h3 className="text-lg font-semibold text-gray-800">Kategori Bazlı Sentiment</h3>
        </div>
        
        <div className="h-64">
          <Bar data={barData} options={barOptions} />
        </div>
      </div>

      {/* Toksiklik İstatistikleri */}
      {data.toxicity && (
        <div className="bg-white rounded-lg shadow-lg p-6">
          <div className="flex items-center mb-4">
            <ShieldExclamationIcon className="h-5 w-5 text-red-500 mr-2" />
            <h3 className="text-lg font-semibold text-gray-800">Toksiklik Analizi</h3>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="text-center p-4 bg-gray-50 rounded-lg">
              <div className="text-2xl font-bold text-gray-800">
                {data.toxicity.total_scanned}
              </div>
              <div className="text-sm text-gray-600">Toplam Tarama</div>
            </div>

            <div className="text-center p-4 bg-green-50 rounded-lg">
              <div className="text-2xl font-bold text-green-600">
                {data.toxicity.low_toxicity}
              </div>
              <div className="text-sm text-green-600">Düşük Toksiklik</div>
            </div>

            <div className="text-center p-4 bg-yellow-50 rounded-lg">
              <div className="text-2xl font-bold text-yellow-600">
                {data.toxicity.medium_toxicity}
              </div>
              <div className="text-sm text-yellow-600">Orta Toksiklik</div>
            </div>

            <div className="text-center p-4 bg-red-50 rounded-lg">
              <div className="text-2xl font-bold text-red-600">
                {data.toxicity.high_toxicity}
              </div>
              <div className="text-sm text-red-600">Yüksek Toksiklik</div>
            </div>
          </div>

          <div className="mt-4 p-3 bg-blue-50 rounded-lg">
            <div className="text-sm text-blue-800">
              <strong>Ortalama Toksiklik Skoru: </strong>
              {(data.toxicity.average_toxicity * 100).toFixed(1)}%
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default CategoryStats;