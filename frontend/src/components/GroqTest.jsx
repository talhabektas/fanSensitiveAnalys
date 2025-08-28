import React, { useState } from 'react';
import { groqService, groqUtils } from '../services/groqService';
import { 
  BeakerIcon, 
  ArrowPathIcon, 
  CheckCircleIcon,
  ExclamationCircleIcon,
  SparklesIcon
} from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

const GroqTest = () => {
  const [testText, setTestText] = useState('Galatasaray bu maçı harika oynadı, Icardi muhteşemdi!');
  const [result, setResult] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const runTest = async () => {
    if (!testText.trim()) {
      toast.error('Lütfen bir metin girin');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      setResult(null);

      const response = await groqService.testGrokAI(testText);
      setResult(response);
      
      // Başarı mesajı gösterilmesi service'de zaten yapılıyor
    } catch (err) {
      setError(err.message);
      // Hata mesajı gösterilmesi service'de zaten yapılıyor
    } finally {
      setLoading(false);
    }
  };

  const getSentimentColor = (label) => {
    switch (label) {
      case 'POSITIVE':
        return 'text-green-600 bg-green-50 border-green-200';
      case 'NEGATIVE':
        return 'text-red-600 bg-red-50 border-red-200';
      case 'NEUTRAL':
        return 'text-gray-600 bg-gray-50 border-gray-200';
      default:
        return 'text-gray-600 bg-gray-50 border-gray-200';
    }
  };

  const getSentimentIcon = (label) => {
    switch (label) {
      case 'POSITIVE':
        return '😊';
      case 'NEGATIVE':
        return '😞';
      case 'NEUTRAL':
        return '😐';
      default:
        return '❓';
    }
  };

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      {/* Header */}
      <div className="flex items-center mb-6">
        <BeakerIcon className="h-6 w-6 text-blue-500 mr-2" />
        <h3 className="text-xl font-semibold text-gray-800">Grok AI Test Aracı</h3>
        <span className="ml-2 px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
          Beta
        </span>
      </div>

      {/* Test Input */}
      <div className="mb-6">
        <label className="block text-sm font-medium text-gray-700 mb-2">
          Test Metni
        </label>
        <textarea
          value={testText}
          onChange={(e) => setTestText(e.target.value)}
          placeholder="Analiz edilecek Türkçe futbol yorumunu buraya yazın..."
          className="w-full p-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none"
          rows={3}
          maxLength={500}
        />
        <div className="mt-1 text-xs text-gray-500">
          {testText.length}/500 karakter
        </div>
      </div>

      {/* Test Button */}
      <div className="mb-6">
        <button
          onClick={runTest}
          disabled={loading || !testText.trim()}
          className={`flex items-center px-6 py-3 rounded-lg font-medium transition-all ${
            loading || !testText.trim()
              ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
              : 'bg-blue-500 text-white hover:bg-blue-600 active:bg-blue-700 hover:shadow-md'
          }`}
        >
          {loading ? (
            <ArrowPathIcon className="h-5 w-5 mr-2 animate-spin" />
          ) : (
            <SparklesIcon className="h-5 w-5 mr-2" />
          )}
          {loading ? 'Grok AI Analiz Ediyor...' : '🤖 Grok AI ile Analiz Et'}
        </button>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="text-center py-8">
          <div className="animate-pulse">
            <div className="h-4 bg-gray-200 rounded mb-2"></div>
            <div className="h-4 bg-gray-200 rounded mb-2 w-3/4 mx-auto"></div>
            <div className="h-4 bg-gray-200 rounded w-1/2 mx-auto"></div>
          </div>
          <p className="mt-4 text-blue-600">
            🚀 Grok AI hibrit analiz yapıyor...
          </p>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
          <div className="flex items-center">
            <ExclamationCircleIcon className="h-5 w-5 text-red-500 mr-2" />
            <span className="text-red-700 font-medium">Test Başarısız</span>
          </div>
          <p className="text-red-600 text-sm mt-1">{error}</p>
        </div>
      )}

      {/* Results */}
      {result && (
        <div className="space-y-6">
          {/* Success Indicator */}
          <div className="flex items-center p-3 bg-green-50 border border-green-200 rounded-lg">
            <CheckCircleIcon className="h-5 w-5 text-green-500 mr-2" />
            <span className="text-green-700 font-medium">Grok AI Analizi Başarılı!</span>
          </div>

          {/* Sentiment Analysis Results */}
          <div className="p-4 border border-gray-200 rounded-lg">
            <h4 className="font-semibold text-gray-800 mb-3 flex items-center">
              <SparklesIcon className="h-4 w-4 mr-1" />
              Sentiment Analizi
            </h4>
            
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
              {/* Sentiment */}
              <div className={`p-3 rounded-lg border ${getSentimentColor(result.sentiment_analysis?.label)}`}>
                <div className="text-center">
                  <div className="text-2xl mb-1">
                    {getSentimentIcon(result.sentiment_analysis?.label)}
                  </div>
                  <div className="font-medium">
                    {result.sentiment_analysis?.label || 'N/A'}
                  </div>
                  <div className="text-sm opacity-75">Sentiment</div>
                </div>
              </div>

              {/* Confidence */}
              <div className="p-3 rounded-lg border border-blue-200 bg-blue-50">
                <div className="text-center">
                  <div className="text-xl font-bold text-blue-600">
                    {result.sentiment_analysis?.confidence 
                      ? (result.sentiment_analysis.confidence * 100).toFixed(1) + '%'
                      : 'N/A'
                    }
                  </div>
                  <div className="text-sm text-blue-600">Güven Skoru</div>
                </div>
              </div>

              {/* Toxicity */}
              <div className="p-3 rounded-lg border border-purple-200 bg-purple-50">
                <div className="text-center">
                  <div className="text-xl font-bold text-purple-600">
                    {result.sentiment_analysis?.toxicity 
                      ? (result.sentiment_analysis.toxicity * 100).toFixed(1) + '%'
                      : '0%'
                    }
                  </div>
                  <div className="text-sm text-purple-600">Toksiklik</div>
                </div>
              </div>

              {/* Model */}
              <div className="p-3 rounded-lg border border-gray-200 bg-gray-50">
                <div className="text-center">
                  <div className="text-sm font-medium text-gray-800">
                    {groqUtils.formatModelName(result.sentiment_analysis?.model)}
                  </div>
                  <div className="text-xs text-gray-600">Model</div>
                </div>
              </div>
            </div>
          </div>

          {/* Categorization Results */}
          <div className="p-4 border border-gray-200 rounded-lg">
            <h4 className="font-semibold text-gray-800 mb-3">Kategorizasyon</h4>
            
            <div className="flex items-center justify-between mb-3">
              <span className="text-gray-700">Kategori:</span>
              <span 
                className="px-3 py-1 rounded-full text-sm font-medium"
                style={{ 
                  backgroundColor: `${groqUtils.getCategoryColor(result.categorization?.category)}20`,
                  color: groqUtils.getCategoryColor(result.categorization?.category)
                }}
              >
                {result.categorization?.category || 'Belirtilmedi'}
              </span>
            </div>

            {/* Keywords */}
            {result.categorization?.keywords && result.categorization.keywords.length > 0 && (
              <div>
                <span className="text-gray-700 text-sm">Anahtar Kelimeler:</span>
                <div className="flex flex-wrap gap-1 mt-2">
                  {result.categorization.keywords.map((keyword, index) => (
                    <span
                      key={index}
                      className="px-2 py-1 bg-indigo-100 text-indigo-800 text-xs rounded"
                    >
                      {keyword}
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Summary */}
          {result.sentiment_analysis?.summary && (
            <div className="p-4 border border-gray-200 rounded-lg">
              <h4 className="font-semibold text-gray-800 mb-2">AI Özeti</h4>
              <p className="text-gray-700 text-sm">
                {result.sentiment_analysis.summary}
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default GroqTest;