import { apiClient } from './api';
import toast from 'react-hot-toast';

// 🚀 Grok AI Enhanced Services

export const groqService = {
  // 1. Gelişmiş İstatistikler
  getEnhancedStats: async (teamId = null) => {
    try {
      const url = teamId 
        ? `/sentiments/enhanced-stats/${teamId}`
        : '/sentiments/enhanced-stats';
        
      const response = await apiClient.get(url);
      return response.data;
    } catch (error) {
      toast.error('Gelişmiş istatistikler alınamadı');
      throw error;
    }
  },

  // 2. Günlük Özet Oluştur
  generateDailySummary: async (teamId) => {
    try {
      toast.loading('Özet oluşturuluyor...', { id: 'summary-loading' });
      
      // teamId null ise query parameter olarak gönder, yoksa path parameter
      const url = teamId 
        ? `/sentiments/summary/generate/${teamId}`
        : `/sentiments/summary/generate`;
      
      console.log('Making API call to:', url);
      const response = await apiClient.post(url);
      
      console.log('API response for summary:', response);
      console.log('Summary data:', response.data);
      
      toast.success('Özet başarıyla oluşturuldu!', { id: 'summary-loading' });
      return response.data;
    } catch (error) {
      console.error('API Error:', error);
      toast.error('Özet oluşturulamadı', { id: 'summary-loading' });
      throw error;
    }
  },

  // 3. Trend İçgörüleri
  getTrendInsights: async (teamId = null) => {
    try {
      const url = teamId 
        ? `/sentiments/trends/insights/${teamId}`
        : '/sentiments/trends/insights';
        
      const response = await apiClient.get(url);
      return response.data;
    } catch (error) {
      toast.error('Trend içgörüleri alınamadı');
      throw error;
    }
  },

  // 4. Kategori İstatistikleri
  getCategoryStats: async (teamId = null) => {
    try {
      const params = teamId ? `?team_id=${teamId}` : '';
      const response = await apiClient.get(`/sentiments/categories/stats${params}`);
      return {
        categories: response.categories,
        toxicity: response.toxicity
      };
    } catch (error) {
      toast.error('Kategori istatistikleri alınamadı');
      throw error;
    }
  },

  // 5. Grok AI Test
  testGrokAI: async (text) => {
    try {
      toast.loading('Grok AI test ediliyor...', { id: 'grok-test' });
      
      const response = await apiClient.post('/sentiments/test-grok', { text });
      
      toast.success('Grok AI test başarılı!', { id: 'grok-test' });
      return response.results;
    } catch (error) {
      toast.error('Grok AI testi başarısız', { id: 'grok-test' });
      throw error;
    }
  },

  // 6. Hibrit Analiz
  analyzeWithAI: async (text) => {
    try {
      const response = await apiClient.post('/sentiments/analyze', { text });
      return response.result;
    } catch (error) {
      toast.error('AI analizi başarısız');
      throw error;
    }
  }
};

// Kategori renk haritası
export const categoryColors = {
  'Takım Performansı': '#3B82F6',
  'Oyuncu Eleştirisi': '#EF4444', 
  'Hakem Kararları': '#F59E0B',
  'Transfer Haberleri': '#10B981',
  'Teknik Direktör': '#8B5CF6',
  'Genel': '#6B7280',
};

// Toksiklik seviyeleri
export const toxicityLevels = {
  low: { label: 'Düşük', color: '#10B981', threshold: [0, 0.3] },
  medium: { label: 'Orta', color: '#F59E0B', threshold: [0.3, 0.7] },
  high: { label: 'Yüksek', color: '#EF4444', threshold: [0.7, 1.0] }
};

// Güven seviyeleri
export const confidenceLevels = {
  low: { label: 'Düşük Güven', color: '#EF4444', threshold: [0, 0.6] },
  medium: { label: 'Orta Güven', color: '#F59E0B', threshold: [0.6, 0.8] },
  high: { label: 'Yüksek Güven', color: '#10B981', threshold: [0.8, 1.0] }
};

// Yardımcı fonksiyonlar
export const groqUtils = {
  // Toksiklik seviyesini belirle
  getToxicityLevel: (score) => {
    if (score >= 0.7) return toxicityLevels.high;
    if (score >= 0.3) return toxicityLevels.medium;
    return toxicityLevels.low;
  },

  // Güven seviyesini belirle
  getConfidenceLevel: (score) => {
    if (score >= 0.8) return confidenceLevels.high;
    if (score >= 0.6) return confidenceLevels.medium;
    return confidenceLevels.low;
  },

  // Kategori rengini al
  getCategoryColor: (category) => {
    return categoryColors[category] || categoryColors['Genel'];
  },

  // Yüzdelik dilim hesapla
  calculatePercentage: (value, total) => {
    return total > 0 ? ((value / total) * 100).toFixed(1) : 0;
  },

  // Model tipini format la
  formatModelName: (modelUsed) => {
    const modelMap = {
      'hybrid-consensus': 'Hibrit Konsensüs',
      'hybrid-hf-primary': 'Hibrit (HF Öncelik)',
      'hybrid-groq-primary': 'Hibrit (Grok Öncelik)',
      'groq-only': 'Sadece Grok AI',
      'hf-only': 'Sadece HuggingFace',
      'groq-fallback': 'Grok Yedek'
    };
    return modelMap[modelUsed] || modelUsed;
  }
};

export default groqService;