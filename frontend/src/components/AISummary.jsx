import React, { useState, useEffect } from 'react';
import { groqService } from '../services/groqService';
import { 
  SparklesIcon, 
  ClockIcon, 
  ArrowPathIcon,
  ExclamationTriangleIcon,
  ChatBubbleLeftRightIcon
} from '@heroicons/react/24/outline';
import toast from 'react-hot-toast';

const AISummary = ({ teamId, teamName = "Tüm Takımlar" }) => {
  const [summary, setSummary] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [lastGenerated, setLastGenerated] = useState(null);

  useEffect(() => {
    // Component yüklendiğinde otomatik özet almayacak
    // Kullanıcı butona tıklamalı
  }, [teamId]);

  const generateSummary = async () => {
    // teamId null ise "Tüm Takımlar" için özet oluştur
    console.log('AISummary generateSummary called with teamId:', teamId, 'teamName:', teamName);
    
    try {
      setLoading(true);
      setError(null);
      
      // teamId null olsa bile API çağrısı yap - backend null'ı handle eder
      const response = await groqService.generateDailySummary(teamId);
      setSummary(response);
      setLastGenerated(new Date());
      
      toast.success('AI Özeti başarıyla oluşturuldu! 🤖');
    } catch (err) {
      console.error('Summary generation error:', err);
      setError(err.message);
      toast.error('Özet oluşturulamadı: ' + err.message);
    } finally {
      setLoading(false);
    }
  };

  const formatDate = (date) => {
    return new Intl.DateTimeFormat('tr-TR', {
      day: 'numeric',
      month: 'long',
      hour: '2-digit',
      minute: '2-digit'
    }).format(date);
  };

  return (
    <div className="bg-white rounded-lg shadow-lg p-6">
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center">
          <SparklesIcon className="h-6 w-6 text-purple-500 mr-2" />
          <h3 className="text-xl font-semibold text-gray-800">AI Yorum Özeti</h3>
          <span className="ml-2 px-2 py-1 bg-purple-100 text-purple-800 text-xs rounded-full">
            Grok AI
          </span>
        </div>

        <div className="flex space-x-2">
          <button
            onClick={async () => {
              try {
                toast.loading('Yorumlar toplanıyor...', { id: 'collect' });
                await fetch('http://localhost:8060/api/v1/youtube/collect', { method: 'POST' });
                toast.success('Yorumlar toplandı!', { id: 'collect' });
              } catch (err) {
                toast.error('Yorum toplama başarısız', { id: 'collect' });
              }
            }}
            className="flex items-center px-3 py-2 bg-green-500 text-white rounded-lg hover:bg-green-600 text-sm"
          >
            📥 Yorum Topla
          </button>
          
          <button
            onClick={generateSummary}
            disabled={loading}
            className={`flex items-center px-4 py-2 rounded-lg font-medium transition-colors ${
              loading
                ? 'bg-gray-100 text-gray-400 cursor-not-allowed'
                : 'bg-purple-500 text-white hover:bg-purple-600 active:bg-purple-700'
            }`}
          >
            {loading ? (
              <ArrowPathIcon className="h-4 w-4 mr-2 animate-spin" />
            ) : (
              <ChatBubbleLeftRightIcon className="h-4 w-4 mr-2" />
            )}
            {loading ? 'Oluşturuluyor...' : 'Özet Oluştur'}
          </button>
        </div>
      </div>

      {/* Team Info */}
      <div className="mb-4 p-3 bg-gray-50 rounded-lg">
        <div className="text-sm text-gray-600">
          <strong>Takım:</strong> {teamName}
        </div>
        <div className="text-sm text-gray-600 mt-1">
          <strong>Dönem:</strong> Son 24 Saat
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="text-center py-8">
          <div className="animate-pulse">
            <div className="h-4 bg-gray-200 rounded mb-2"></div>
            <div className="h-4 bg-gray-200 rounded mb-2 w-3/4 mx-auto"></div>
            <div className="h-4 bg-gray-200 rounded w-1/2 mx-auto"></div>
          </div>
          <p className="mt-4 text-purple-600">
            🤖 Grok AI yorumları analiz ediyor...
          </p>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="text-center py-8">
          <ExclamationTriangleIcon className="h-12 w-12 text-red-400 mx-auto mb-2" />
          <p className="text-red-600 mb-4">{error}</p>
          <button
            onClick={generateSummary}
            className="px-4 py-2 bg-red-500 text-white rounded hover:bg-red-600"
          >
            Tekrar Dene
          </button>
        </div>
      )}

      {/* Summary Content */}
      {summary && (
        <div className="space-y-6">
          {/* Summary Stats */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="text-center p-4 bg-blue-50 rounded-lg">
              <div className="text-2xl font-bold text-blue-600">
                {summary.total_comments}
              </div>
              <div className="text-sm text-blue-600">Analiz Edilen Yorum</div>
            </div>

            <div className="text-center p-4 bg-green-50 rounded-lg">
              <div className="text-2xl font-bold text-green-600">
                {summary.main_topics ? summary.main_topics.length : 0}
              </div>
              <div className="text-sm text-green-600">Ana Konu</div>
            </div>

            <div className="text-center p-4 bg-purple-50 rounded-lg">
              <div className="text-2xl font-bold text-purple-600">AI</div>
              <div className="text-sm text-purple-600">Grok Analizi</div>
            </div>
          </div>

          {/* AI Summary Text */}
          <div className="p-6 bg-gradient-to-r from-purple-50 to-pink-50 rounded-lg border border-purple-200">
            <div className="flex items-start mb-3">
              <SparklesIcon className="h-5 w-5 text-purple-500 mr-2 mt-0.5 flex-shrink-0" />
              <h4 className="text-lg font-medium text-purple-800">AI Özeti</h4>
            </div>
            
            <div className="prose prose-sm max-w-none">
              <p className="text-gray-700 leading-relaxed whitespace-pre-wrap">
                {summary.summary}
              </p>
            </div>
          </div>

          {/* Main Topics */}
          {summary.main_topics && summary.main_topics.length > 0 && (
            <div>
              <h4 className="text-md font-medium text-gray-800 mb-3">Ana Konular</h4>
              <div className="flex flex-wrap gap-2">
                {summary.main_topics.map((topic, index) => (
                  <span
                    key={index}
                    className="px-3 py-1 bg-indigo-100 text-indigo-800 text-sm rounded-full"
                  >
                    {topic}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Generated Info */}
          <div className="flex items-center justify-between pt-4 border-t border-gray-200">
            <div className="flex items-center text-sm text-gray-500">
              <ClockIcon className="h-4 w-4 mr-1" />
              {lastGenerated ? (
                <span>Son güncelleme: {formatDate(lastGenerated)}</span>
              ) : (
                <span>Henüz özet oluşturulmadı</span>
              )}
            </div>

            <div className="text-xs text-gray-400">
              Powered by Grok AI 🤖
            </div>
          </div>
        </div>
      )}

      {/* Empty State */}
      {!summary && !loading && !error && (
        <div className="text-center py-12">
          <ChatBubbleLeftRightIcon className="h-16 w-16 text-gray-300 mx-auto mb-4" />
          <h4 className="text-lg font-medium text-gray-600 mb-2">
            AI Özeti Hazır Değil
          </h4>
          <p className="text-gray-500 mb-4">
            Son 24 saatin yorumlarından akıllı bir özet oluşturmak için butona tıklayın
          </p>
          <button
            onClick={generateSummary}
            className="px-6 py-3 bg-purple-500 text-white rounded-lg hover:bg-purple-600 font-medium"
          >
            🤖 AI Özeti Oluştur
          </button>
        </div>
      )}
    </div>
  );
};

export default AISummary;