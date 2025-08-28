# 🚀 N8N Workflow Entegrasyonu

Bu klasör, **FanSensitive Analytics** projesi için N8N otomatizasyon workflow'larını içerir.

## 📁 Workflow Dosyaları

### 1. 🔄 `youtube-auto-collector.json`
**Otomatik YouTube Yorum Toplayıcı**
- **Çalışma Sıklığı:** Her 2 saatte bir
- **İşlevler:**
  - YouTube API'den futbol yorumlarını otomatik toplama
  - Groq AI ile sentiment analizi
  - Başarılı toplamada Slack bildirimi
  - Hata durumunda email uyarısı
- **Entegrasyonlar:** Slack, Email, Backend API

### 2. 🚨 `sentiment-alert-system.json`
**Akıllı Sentiment Uyarı Sistemi**
- **Çalışma Sıklığı:** Her 30 dakikada bir
- **İşlevler:**
  - Sentiment seviyelerini sürekli izleme
  - Negatif sentiment %30'u geçince Discord uyarısı
  - Pozitif sentiment %50'yi geçince kutlama mesajı
  - Trend analizi ve içgörü oluşturma
- **Entegrasyonlar:** Discord, Backend API

### 3. 🔬 `advanced-analytics-workflow.json`
**Gelişmiş Analytics Otomasyonu**
- **Çalışma Sıklığı:** Her 6 saatte bir
- **İşlevler:**
  - Kapsamlı performans analizi
  - Takım sıralaması ve toksiklik analizi
  - AI destekli trend öngörüleri
  - Akıllı uyarı sistemi (kritik durumlar için Discord)
  - Düzenli raporlama (Slack'e 6 saatlik özet)
- **Entegrasyonlar:** Discord, Slack, Backend API

## 🛠 Kurulum Adımları

### 1. N8N Kurulumu
```bash
# Docker ile N8N başlatma
docker run -it --rm --name n8n -p 5678:5678 n8nio/n8n

# Veya npm ile global kurulum
npm install n8n -g
n8n start
```

### 2. Workflow Import Etme
1. N8N arayüzünü aç: `http://localhost:5678`
2. **Import** butonuna tıkla
3. JSON dosyalarını tek tek import et
4. Her workflow için gerekli credentials'ları ayarla

### 3. Credential Ayarları

#### Slack Integration
- **Webhook URL:** Slack app'ten webhook URL'ini al
- **Bot Token:** `xoxb-` ile başlayan bot token
- **Channel ID:** Bildirim gönderilecek kanal

#### Discord Integration
- **Bot Token:** Discord developer portal'dan bot token
- **Channel ID:** Mesaj gönderilecek kanal ID'si

#### LinkedIn API
- **Client ID:** LinkedIn Developer Console'dan
- **Client Secret:** LinkedIn app secret
- **Redirect URI:** N8N callback URL'i

#### Email Settings
- **SMTP Server:** Email provider ayarları
- **Credentials:** Email gönderim için kullanıcı bilgileri

## 🎯 Proje Showcase Özellikleri

N8N entegrasyonu sayesinde projen şu özellikleri kazanır:
- 🔄 **Otomatik Veri Toplama:** YouTube API'den sürekli güncel futbol yorumları
- 🚨 **Akıllı Uyarılar:** Kritik sentiment değişimlerinde anlık bildirimler  
- 🔬 **Gelişmiş Analytics:** 6 saatlik performans raporları ve trend analizleri
- 🤖 **AI Destekli İçgörüler:** Grok AI ile gerçek zamanlı futbol analizi
- 📊 **Multi-Platform Bildirimler:** Discord, Slack entegrasyonları

## 📱 Monitoring & Alerts

### Discord Uyarıları
- 🚨 **Negatif Alert:** Sentiment < -30%
- 🎉 **Pozitif Alert:** Sentiment > +50%
- 📊 **Trend Değişimi:** Ani sentiment değişimlerinde

### Slack Bildirimleri
- ✅ **Başarılı İşlemler:** Veri toplama, analiz tamamlanması
- ❌ **Hata Durumları:** API hatası, bağlantı sorunları
- 📈 **LinkedIn Paylaşımı:** Otomatik post başarısı

## 🔧 Özelleştirme

### Zaman Ayarları
- Cron expression'larını değiştirerek çalışma saatlerini özelleştirin
- Daha sık veya daha seyrek çalıştırma seçenekleri

### Threshold Değerleri
- Sentiment eşik değerlerini projenize göre ayarlayın
- Alert seviyelerini takım ihtiyaçlarına göre özelleştirin

### İçerik Şablonları
- LinkedIn post şablonlarını markanıza göre güncelleyin
- Discord/Slack mesaj formatlarını özelleştirin

## 📈 Analytics & Reporting

N8N execution history'den:
- Workflow çalışma başarı oranları
- Ortalama işlem süreleri  
- Hata logları ve debug bilgileri
- Performance metrikleri

## 🚨 Troubleshooting

### Yaygın Sorunlar
1. **API Rate Limits:** YouTube/LinkedIn API limitlerini aştığınızda bekleme süreleri
2. **Token Expiry:** OAuth tokenlarının yenilenmesi gerektiğinde
3. **Network Errors:** Backend API'ye erişim sorunlarında

### Debug Modları
- N8N workflow'larında debug mode aktifleştirin
- Console logları ile API response'larını inceleyin
- Test modunda manual execution yapın

---

**🎯 Sonuç:** Bu N8N workflow'ları ile projeniz tamamen otomatik çalışan, akıllı uyarılar gönderen ve LinkedIn'de profesyonel görünüm sağlayan bir sistem haline gelir!