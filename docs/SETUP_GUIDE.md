# 🚀 Detaylı Kurulum Rehberi

Bu rehber, Taraftar Duygu Analizi platformunu sıfırdan kurmak için adım adım talimatlar içerir.

## 📋 İçindekiler

1. [Gerekli Hesapların Oluşturulması](#gerekli-hesapların-oluşturulması)
2. [Lokal Geliştirme Ortamı](#lokal-geliştirme-ortamı)  
3. [Production Deployment](#production-deployment)
4. [n8n Workflow Kurulumu](#n8n-workflow-kurulumu)
5. [Troubleshooting](#troubleshooting)

## 🔐 Gerekli Hesapların Oluşturulması

### 1. MongoDB Atlas (Veritabanı) - ÜCRETSİZ

**Adım 1: Hesap Oluşturma**
1. [MongoDB Atlas](https://mongodb.com/atlas) adresine gidin
2. "Try Free" butonuna tıklayın
3. Email ile hesap oluşturun veya Google/GitHub ile giriş yapın

**Adım 2: Cluster Oluşturma**
1. "Build a Database" seçeneğini seçin
2. **M0 FREE** cluster'ı seçin (512MB - MVP için yeterli)
3. Provider: **AWS** 
4. Region: **Europe (Frankfurt)** - Türkiye'ye en yakın
5. Cluster Name: `taraftar-cluster` (veya istediğiniz isim)
6. "Create Cluster" butonuna tıklayın

**Adım 3: Database User Oluşturma**
1. Security → Database Access menüsüne gidin
2. "Add New Database User" butonuna tıklayın
3. Authentication Method: **Password**
4. Username: `taraftar_user`
5. Password: Güçlü bir şifre oluşturun ve kaydedin
6. Database User Privileges: **Built-in Role** → **Read and write to any database**
7. "Add User" butonuna tıklayın

**Adım 4: Network Access Konfigürasyonu**
1. Security → Network Access menüsüne gidin
2. "Add IP Address" butonuna tıklayın
3. **Allow Access from Anywhere** seçeneğini seçin
4. IP Address: `0.0.0.0/0` (otomatik doldurulur)
5. Comment: `Allow all IPs for development`
6. "Confirm" butonuna tıklayın

**Adım 5: Connection String Alma**
1. Database → Connect butonuna tıklayın
2. **Connect your application** seçeneğini seçin
3. Driver: **Go** / Version: **1.10 or later**
4. Connection string'i kopyalayın:
   ```
   mongodb+srv://taraftar_user:<password>@taraftar-cluster.xxx.mongodb.net/
   ```
5. `<password>` kısmını gerçek şifrenizle değiştirin

### 2. Reddit API - ÜCRETSİZ

**Adım 1: Reddit Hesabı**
1. [Reddit](https://reddit.com) hesabınız yoksa oluşturun
2. Email adresinizi doğrulayın

**Adım 2: App Oluşturma**
1. [Reddit App Preferences](https://reddit.com/prefs/apps) sayfasına gidin
2. "Create App" veya "Create Another App" butonuna tıklayın
3. **Name**: `Taraftar Sentiment Analyzer`
4. **App type**: **script** seçeneğini işaretleyin
5. **Description**: `Turkish football fan sentiment analysis tool`
6. **About URL**: Boş bırakabilirsiniz
7. **Redirect URI**: `http://localhost:8080`
8. "Create app" butonuna tıklayın

**Adım 3: Credentials Alma**
1. Oluşturulan app'in altında:
   - **Client ID**: App adının altındaki küçük yazı (örn: `abc123xyz`)
   - **Client Secret**: "secret" yazan kısmın karşısındaki uzun kod
2. Bu bilgileri kaydedin

### 3. HuggingFace (AI Model) - ÜCRETSİZ

**Adım 1: Hesap Oluşturma**
1. [HuggingFace](https://huggingface.co) adresine gidin
2. "Sign Up" ile hesap oluşturun (GitHub/Google ile de olur)

**Adım 2: Access Token Oluşturma**
1. Sağ üst köşedeki profil resminize tıklayın
2. **Settings** menüsüne gidin
3. Sol menüden **Access Tokens** seçeneğini seçin
4. "New token" butonuna tıklayın
5. **Name**: `taraftar-analizi`
6. **Type**: **Read** seçeneğini seçin
7. "Generate a token" butonuna tıklayın
8. Oluşan token'ı kopyalayın ve güvenli bir yere kaydedin

### 4. YouTube Data API (Opsiyonel) - ÜCRETSİZ

**Adım 1: Google Cloud Console**
1. [Google Cloud Console](https://console.cloud.google.com) adresine gidin
2. Google hesabınızla giriş yapın
3. "New Project" ile yeni proje oluşturun
4. Project Name: `taraftar-sentiment-analysis`

**Adım 2: API Etkinleştirme**
1. Sol menüden **APIs & Services** → **Library** seçeneğine gidin
2. "YouTube Data API v3" aratın
3. API'ye tıklayın ve **Enable** butonuna basın

**Adım 3: API Key Oluşturma**
1. **APIs & Services** → **Credentials** menüsüne gidin
2. "**+ CREATE CREDENTIALS**" → **API key** seçeneğini seçin
3. API key oluşturulacak
4. **Restrict Key** butonuna tıklayın
5. **API restrictions** kısmında **Restrict key** seçeneğini işaretleyin
6. **YouTube Data API v3**'ü seçin
7. **Save** butonuna tıklayın

### 5. Railway (Backend Hosting) - ÜCRETSİZ

**Adım 1: Hesap Oluşturma**
1. [Railway](https://railway.app) adresine gidin
2. **Login with GitHub** ile giriş yapın
3. GitHub hesabınızı bağlayın

**Adım 2: GitHub Repository Hazırlama**
1. Bu projeyi fork edin veya kendi repository'nize push edin
2. Repository'nin public olduğundan emin olun

### 6. Vercel (Frontend Hosting) - ÜCRETSİZ

**Adım 1: Hesap Oluşturma**
1. [Vercel](https://vercel.com) adresine gidin
2. **Continue with GitHub** ile giriş yapın
3. GitHub hesabınızı bağlayın

## 🛠 Lokal Geliştirme Ortamı

### Ön Gereksinimler

```bash
# Go 1.21+
go version

# Node.js 18+
node --version

# Docker (opsiyonel ama önerilen)
docker --version

# Git
git --version
```

### 1. Repository'yi Klonlama

```bash
git clone https://github.com/yourusername/fanSensitiveAnalys.git
cd fanSensitiveAnalys
```

### 2. Environment Dosyalarını Oluşturma

**Backend Environment**
```bash
cp backend/.env.example backend/.env
```

`backend/.env` dosyasını düzenleyin:
```bash
# Server Configuration
PORT=8080
GIN_MODE=debug
API_SECRET=dev_secret_key_change_in_production

# MongoDB Configuration (yukarıda aldığınız bilgiler)
MONGODB_URI=mongodb+srv://taraftar_user:yourpassword@taraftar-cluster.xxx.mongodb.net/
MONGODB_DATABASE=taraftar_analizi

# Reddit API Credentials (yukarıda aldığınız bilgiler)
REDDIT_CLIENT_ID=your_reddit_client_id
REDDIT_CLIENT_SECRET=your_reddit_client_secret
REDDIT_USERNAME=your_reddit_username
REDDIT_PASSWORD=your_reddit_password

# HuggingFace API (yukarıda aldığınız token)
HUGGINGFACE_TOKEN=your_huggingface_token

# YouTube API (opsiyonel)
YOUTUBE_API_KEY=your_youtube_api_key

# External URLs
FRONTEND_URL=http://localhost:3000
N8N_WEBHOOK_URL=http://localhost:5678/webhook
```

**Frontend Environment**
```bash
cp frontend/.env.example frontend/.env
```

`frontend/.env` dosyasını düzenleyin:
```bash
VITE_API_URL=http://localhost:8080/api/v1
VITE_WS_URL=ws://localhost:8080/ws
VITE_APP_NAME=Taraftar Duygu Analizi
VITE_APP_VERSION=1.0.0
```

**n8n Environment**
```bash
cp n8n.env.example n8n.env
```

`n8n.env` dosyasını düzenleyin:
```bash
# n8n Configuration
N8N_BASIC_AUTH_ACTIVE=true
N8N_BASIC_AUTH_USER=admin
N8N_BASIC_AUTH_PASSWORD=admin123

# API Keys
REDDIT_CLIENT_ID=your_reddit_client_id
REDDIT_CLIENT_SECRET=your_reddit_client_secret
REDDIT_USERNAME=your_reddit_username
REDDIT_PASSWORD=your_reddit_password
HUGGINGFACE_TOKEN=your_huggingface_token
YOUTUBE_API_KEY=your_youtube_api_key

# Backend URL
BACKEND_URL=http://backend:8080
API_SECRET=dev_secret_key_change_in_production
```

### 3. Docker ile Çalıştırma (Önerilen)

```bash
# Tüm servisleri başlat
docker-compose up -d

# Servislerin durumunu kontrol et
docker-compose ps

# Logları izle
docker-compose logs -f

# Sadece backend logları
docker-compose logs -f backend

# Servisleri durdur
docker-compose down
```

### 4. Manuel Kurulum

**Terminal 1 - Backend:**
```bash
cd backend
go mod tidy
go run main.go
```

**Terminal 2 - Frontend:**
```bash
cd frontend
npm install
npm run dev
```

**Terminal 3 - n8n:**
```bash
npx n8n start
```

### 5. İlk Veri Yükleme

**Takımları yükle:**
```bash
curl -X POST http://localhost:8080/api/v1/teams/seed
```

**Health check:**
```bash
curl http://localhost:8080/health
```

## 🌐 Production Deployment

### 1. Railway Backend Deployment

**Adım 1: Railway'da Yeni Proje**
1. Railway dashboard'a gidin
2. "**New Project**" butonuna tıklayın
3. "**Deploy from GitHub repo**" seçeneğini seçin
4. Repository'nizi seçin

**Adım 2: Konfigürasyon**
1. **Root Directory**: `backend` yazın
2. **Build Command**: Otomatik algılanır
3. **Start Command**: `./main`

**Adım 3: Environment Variables**
Settings → Variables kısmında şunları ekleyin:
```
PORT=8080
GIN_MODE=release
API_SECRET=production_secret_key_very_secure
MONGODB_URI=your_mongodb_atlas_connection_string
MONGODB_DATABASE=taraftar_analizi
REDDIT_CLIENT_ID=your_reddit_client_id
REDDIT_CLIENT_SECRET=your_reddit_client_secret
REDDIT_USERNAME=your_reddit_username
REDDIT_PASSWORD=your_reddit_password
HUGGINGFACE_TOKEN=your_huggingface_token
YOUTUBE_API_KEY=your_youtube_api_key
FRONTEND_URL=https://your-vercel-app.vercel.app
N8N_WEBHOOK_URL=your_n8n_url_if_deployed
```

**Adım 4: Deploy**
1. "**Deploy**" butonuna tıklayın
2. Build tamamlandıktan sonra URL'i kopyalayın
3. URL'e `/health` ekleyerek test edin

### 2. Vercel Frontend Deployment

**Adım 1: Vercel'da Yeni Proje**
1. Vercel dashboard'a gidin
2. "**New Project**" butonuna tıklayın
3. GitHub repository'nizi import edin

**Adım 2: Konfigürasyon**
1. **Framework Preset**: Vite
2. **Root Directory**: `frontend`
3. **Build Command**: `npm run build`
4. **Output Directory**: `dist`
5. **Install Command**: `npm install`

**Adım 3: Environment Variables**
```
VITE_API_URL=https://your-railway-backend.railway.app/api/v1
VITE_WS_URL=wss://your-railway-backend.railway.app/ws
VITE_APP_NAME=Taraftar Duygu Analizi
VITE_APP_VERSION=1.0.0
```

**Adım 4: Deploy**
1. "**Deploy**" butonuna tıklayın
2. Build tamamlandıktan sonra URL'i test edin

### 3. Production'da İlk Veri Yükleme

```bash
# Takımları yükle
curl -X POST https://your-railway-backend.railway.app/api/v1/teams/seed

# Health check
curl https://your-railway-backend.railway.app/health
```

## 🔄 n8n Workflow Kurulumu

### 1. n8n Kurulumu

**Docker ile (Önerilen):**
```bash
docker run -it --rm \
  --name n8n \
  -p 5678:5678 \
  -e N8N_BASIC_AUTH_ACTIVE=true \
  -e N8N_BASIC_AUTH_USER=admin \
  -e N8N_BASIC_AUTH_PASSWORD=password123 \
  -v n8n_data:/home/node/.n8n \
  n8nio/n8n
```

**npm ile:**
```bash
npm install -g n8n
n8n start
```

### 2. Workflow'ları İçe Aktarma

1. n8n arayüzüne gidin: `http://localhost:5678`
2. Kullanıcı adı: `admin`, Şifre: `password123`
3. Sol menüden **Workflows** seçeneğine gidin
4. "**Import from file**" butonuna tıklayın

Şu sırayla workflow'ları import edin:

**1. Reddit Collector (`n8n-workflows/reddit-collector.json`)**
- Her 2 saatte bir Reddit'ten yorum toplar
- Settings → Variables kısmında API key'leri kontrol edin
- **Activate** butonuna tıklayın

**2. Sentiment Analyzer (`n8n-workflows/sentiment-analyzer.json`)**  
- Webhook ile tetiklenir
- İşlenmemiş yorumları analiz eder
- Webhook URL'ini kaydedin: `http://localhost:5678/webhook/sentiment-analysis`

**3. YouTube Collector (`n8n-workflows/youtube-collector.json`)**
- Günlük sabah 6'da YouTube yorumları toplar
- YouTube API key gerekli
- **Activate** butonuna tıklayın

**4. Daily Report (`n8n-workflows/daily-report.json`)**
- Günlük sabah 8'de rapor gönderir
- Email/Telegram ayarları opsiyonel
- **Activate** butonuna tıklayın

### 3. Workflow Test Etme

**Reddit Collector Test:**
1. Workflow'u açın
2. Sağ üstteki "**Execute Workflow**" butonuna tıklayın
3. Başarılı olursa yeşil tick işaretleri göreceksiniz

**Sentiment Analyzer Test:**
```bash
curl -X POST http://localhost:5678/webhook/sentiment-analysis \
  -H "Content-Type: application/json" \
  -d '{}'
```

## 🔧 Troubleshooting

### Yaygın Sorunlar ve Çözümleri

#### 1. MongoDB Connection Error
```
Error: failed to connect to MongoDB
```

**Çözüm:**
- MongoDB Atlas Network Access ayarlarını kontrol edin
- IP adresi `0.0.0.0/0` olarak ayarlandığından emin olun
- Connection string'deki şifrenin doğru olduğunu kontrol edin
- Database user'ın read/write yetkisi olduğunu kontrol edin

#### 2. Reddit API Authentication Error
```
Error: 401 Unauthorized - Reddit API
```

**Çözüm:**
- Reddit app tipinin "script" olduğundan emin olun
- Client ID ve Secret'ın doğru olduğunu kontrol edin
- Reddit kullanıcı adı ve şifresinin doğru olduğunu kontrol edin
- Redirect URI'nin `http://localhost:8080` olduğunu kontrol edin

#### 3. HuggingFace API Rate Limit
```
Error: Model is currently loading, please retry
```

**Çözüm:**
- Birkaç dakika bekleyin (model ilk kullanımda yükleniyor)
- n8n workflow'da batch size'ı azaltın
- Request frequency'sini azaltın

#### 4. Frontend API Connection Error
```
Network Error / CORS Error
```

**Çözüm:**
- Backend'in çalıştığından emin olun
- `VITE_API_URL` environment variable'ının doğru olduğunu kontrol edin
- CORS ayarlarını kontrol edin (backend/middleware/cors.go)

#### 5. Docker Build Error
```
Error: Docker build failed
```

**Çözüm:**
```bash
# Cache'i temizle
docker system prune -a

# Docker compose'u yeniden build et
docker-compose build --no-cache

# Container'ları tamamen temizle
docker-compose down -v
docker-compose up -d
```

### Log İnceleme

**Backend logları:**
```bash
docker-compose logs -f backend
```

**n8n logları:**
```bash
docker-compose logs -f n8n
```

**Tüm servis logları:**
```bash
docker-compose logs -f
```

### Performance Monitoring

**Backend health check:**
```bash
curl http://localhost:8080/health
```

**Database stats:**
```bash
curl http://localhost:8080/api/v1/comments/stats
```

**Sentiment stats:**
```bash
curl http://localhost:8080/api/v1/sentiments/stats
```

## 📞 Yardım Alma

Eğer sorunlarınız devam ediyorsa:

1. **GitHub Issues**: [Create Issue](https://github.com/yourusername/fanSensitiveAnalys/issues)
2. **Logs**: Sorun bildirirken ilgili logları da ekleyin
3. **Environment**: Hangi ortamda çalıştığınızı belirtin (Docker/Manuel)
4. **Steps**: Sorunu reproduce etmek için gerekli adımları yazın

---

✅ **Kurulum tamamlandığında sisteminiz tamamen otomatik çalışacak!**