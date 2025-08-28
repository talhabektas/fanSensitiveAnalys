import React, { useState } from 'react';
import { useSentiment } from '../hooks/useSentiment';
import { TrendingUp, Type, Zap, History } from 'lucide-react';

export const SentimentAnalyzer = () => {
  const [text, setText] = useState('');
  const [result, setResult] = useState(null);

  const { 
    analyzeText, 
    processUnprocessedComments,
    analysisHistory,
    clearAnalysisHistory,
    isAnalyzing, 
    isProcessing,
    formatSentimentLabel,
    getSentimentColor,
    getSentimentIcon
  } = useSentiment();

  const handleAnalyze = async () => {
    if (text.trim()) {
      try {
        const analysisResult = await analyzeText(text);
        setResult(analysisResult?.result);
      } catch (error) {
        console.error('Analysis failed:', error);
      }
    }
  };

  const handleProcessComments = () => {
    processUnprocessedComments();
  };

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Duygu Analizi</h2>
        <p className="text-gray-600">Metin analizi yapın veya bekleyen yorumları işleyin</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Manual Analysis */}
        <div className="lg:col-span-2 space-y-6">
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-medium text-gray-900 flex items-center">
                <Type className="h-5 w-5 mr-2" />
                Manuel Analiz
              </h3>
            </div>
            <div className="card-content space-y-4">
              <div>
                <label className="label">Analiz edilecek metin</label>
                <textarea
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  placeholder="Analiz etmek istediğiniz metni buraya yazın..."
                  rows={6}
                  className="input resize-none"
                />
              </div>
              <button
                onClick={handleAnalyze}
                disabled={!text.trim() || isAnalyzing}
                className="btn btn-primary w-full"
              >
                {isAnalyzing ? 'Analiz Ediliyor...' : 'Analiz Et'}
              </button>

              {result && (
                <div className="mt-4 p-4 rounded-lg border-2" style={{ borderColor: getSentimentColor(result.label) }}>
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-gray-900">Analiz Sonucu</span>
                    <span className="text-2xl">{getSentimentIcon(result.label)}</span>
                  </div>
                  <div className="space-y-2">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Duygu:</span>
                      <span className="font-medium" style={{ color: getSentimentColor(result.label) }}>
                        {formatSentimentLabel(result.label)}
                      </span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Güven Skoru:</span>
                      <span className="font-medium">{(result.confidence * 100).toFixed(1)}%</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Model:</span>
                      <span className="font-medium text-xs">{result.model_used}</span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Batch Processing */}
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-medium text-gray-900 flex items-center">
                <Zap className="h-5 w-5 mr-2" />
                Toplu İşleme
              </h3>
            </div>
            <div className="card-content">
              <p className="text-gray-600 mb-4">
                Bekleyen yorumları otomatik olarak analiz edin
              </p>
              <button
                onClick={handleProcessComments}
                disabled={isProcessing}
                className="btn btn-success w-full"
              >
                {isProcessing ? 'İşleniyor...' : 'Bekleyen Yorumları İşle'}
              </button>
            </div>
          </div>
        </div>

        {/* Analysis History */}
        <div className="space-y-6">
          <div className="card">
            <div className="card-header flex items-center justify-between">
              <h3 className="text-lg font-medium text-gray-900 flex items-center">
                <History className="h-5 w-5 mr-2" />
                Analiz Geçmişi
              </h3>
              {analysisHistory.length > 0 && (
                <button
                  onClick={clearAnalysisHistory}
                  className="text-sm text-gray-500 hover:text-gray-700"
                >
                  Temizle
                </button>
              )}
            </div>
            <div className="card-content">
              {analysisHistory.length > 0 ? (
                <div className="space-y-3 max-h-96 overflow-y-auto">
                  {analysisHistory.map((item) => (
                    <div key={item.id} className="border border-gray-200 rounded-lg p-3">
                      <div className="text-sm text-gray-900 truncate-2-lines mb-2">
                        {item.text}
                      </div>
                      <div className="flex items-center justify-between text-xs">
                        <span 
                          className="px-2 py-1 rounded-full text-white"
                          style={{ backgroundColor: getSentimentColor(item.result.label) }}
                        >
                          {formatSentimentLabel(item.result.label)}
                        </span>
                        <span className="text-gray-500">
                          {new Date(item.timestamp).toLocaleTimeString('tr-TR')}
                        </span>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center py-8">
                  <History className="h-12 w-12 text-gray-400 mx-auto mb-2" />
                  <p className="text-gray-500 text-sm">Henüz analiz geçmişi yok</p>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};