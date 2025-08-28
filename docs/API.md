# 📡 API Documentation

Taraftar Duygu Analizi platformunun REST API dokümantasyonu.

## Base URL

```
Production: https://your-railway-backend.railway.app/api/v1
Local: http://localhost:8080/api/v1
```

## Authentication

n8n webhook endpoint'leri API key gerektirir:

```bash
Headers:
X-API-Key: your_api_secret_key
```

## Response Format

Tüm API yanıtları JSON formatındadır:

```json
{
  "data": "...",
  "message": "Success message",
  "error": "Error message (if any)"
}
```

## Endpoints

### System

#### Health Check
```http
GET /health
```

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "healthy",
    "huggingface": "configured", 
    "reddit": "configured"
  },
  "version": "1.0.0"
}
```

---

### Dashboard

#### Get Dashboard Data
```http
GET /dashboard/data
```

**Response:**
```json
{
  "overall_stats": {
    "total_analyzed": 1500,
    "overall_sentiment": 0.65,
    "sentiment_breakdown": {
      "POSITIVE": 800,
      "NEUTRAL": 400,
      "NEGATIVE": 300
    }
  },
  "team_comparison": {
    "teams": [
      {
        "team_id": "...",
        "team_name": "Galatasaray",
        "avg_sentiment": 0.72,
        "total_comments": 450,
        "ranking": 1
      }
    ]
  },
  "recent_comments": [...],
  "daily_trends": [...]
}
```

#### Get Overall Statistics
```http
GET /dashboard/stats
```

#### Get Team Comparison
```http
GET /dashboard/comparison
```

---

### Comments

#### List Comments
```http
GET /comments?page=1&limit=20&team_id=...&source=reddit&sentiment=POSITIVE
```

**Query Parameters:**
- `page` (int): Page number (default: 1)
- `limit` (int): Items per page (default: 20, max: 100)
- `team_id` (string): Filter by team ObjectID
- `source` (string): Filter by source (reddit, youtube, twitter)
- `sentiment` (string): Filter by sentiment (POSITIVE, NEGATIVE, NEUTRAL)
- `author` (string): Filter by author name
- `language` (string): Filter by language (default: tr)
- `is_processed` (bool): Filter by processing status
- `has_sentiment` (bool): Filter by sentiment analysis status
- `start_date` (string): Filter from date (YYYY-MM-DD)
- `end_date` (string): Filter to date (YYYY-MM-DD)
- `search` (string): Search in comment text
- `sort_by` (string): Sort field (default: created_at)
- `sort_order` (string): Sort order (asc, desc - default: desc)

**Response:**
```json
{
  "comments": [
    {
      "id": "...",
      "source_id": "abc123",
      "source": "reddit",
      "team_id": "...",
      "author": "taraftar_user",
      "text": "Harika bir maç oldu!",
      "url": "https://reddit.com/...",
      "score": 15,
      "subreddit": "galatasaray",
      "language": "tr",
      "is_processed": true,
      "has_sentiment": true,
      "sentiment": {
        "label": "POSITIVE",
        "score": 0.85,
        "confidence": 0.85,
        "model_used": "savasy/bert-base-turkish-sentiment-cased",
        "processed_at": "2024-01-01T12:00:00Z"
      },
      "metadata": {
        "platform": "reddit",
        "post_id": "...",
        "is_reply": false,
        "quality": "high",
        "tags": ["post", "selftext"]
      },
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 1500,
  "page": 1,
  "limit": 20,
  "total_pages": 75
}
```

#### Create Comment
```http
POST /comments
```

**Request Body:**
```json
{
  "source_id": "abc123",
  "source": "reddit",
  "team_id": "optional_team_id",
  "author": "taraftar_user",
  "text": "Maç çok güzeldi!",
  "url": "https://reddit.com/...",
  "score": 10,
  "parent_id": "optional_parent",
  "subreddit": "galatasaray",
  "language": "tr",
  "metadata": {
    "platform": "reddit",
    "quality": "high"
  }
}
```

#### Get Comment Statistics
```http
GET /comments/stats
```

**Response:**
```json
{
  "total_comments": 1500,
  "processed_comments": 1200,
  "unprocessed_comments": 300,
  "sentiment_breakdown": {
    "POSITIVE": 600,
    "NEGATIVE": 300,
    "NEUTRAL": 300
  },
  "source_breakdown": {
    "reddit": 1000,
    "youtube": 400,
    "twitter": 100
  },
  "language_breakdown": {
    "tr": 1450,
    "en": 50
  },
  "daily_stats": [
    {
      "date": "2024-01-01",
      "count": 150,
      "positive": 80,
      "negative": 30,
      "neutral": 40
    }
  ]
}
```

#### Get Unprocessed Comments
```http
GET /comments/unprocessed?limit=50
```

#### Update Comment
```http
PUT /comments/{id}
```

**Request Body:**
```json
{
  "is_processed": true,
  "has_sentiment": true,
  "sentiment": {
    "label": "POSITIVE",
    "score": 0.85,
    "confidence": 0.85,
    "model_used": "...",
    "processed_at": "2024-01-01T12:00:00Z"
  }
}
```

#### Bulk Update Processed
```http
POST /comments/bulk/processed
```

**Request Body:**
```json
{
  "comment_ids": ["id1", "id2", "id3"]
}
```

---

### Sentiment Analysis

#### Analyze Text
```http
POST /sentiments/analyze
```

**Request Body:**
```json
{
  "text": "Bu maç çok güzeldi!",
  "language": "tr"
}
```

**Response:**
```json
{
  "message": "Analysis completed successfully",
  "result": {
    "label": "POSITIVE",
    "score": 0.85,
    "confidence": 0.85,
    "model_used": "savasy/bert-base-turkish-sentiment-cased",
    "processed_at": "2024-01-01T12:00:00Z"
  }
}
```

#### Batch Analyze
```http
POST /sentiments/analyze/batch
```

**Request Body:**
```json
{
  "texts": [
    "Bu maç harika!",
    "Çok kötü oynadılar.",
    "Normal bir performans."
  ]
}
```

**Response:**
```json
{
  "message": "Batch analysis completed",
  "results": [
    {
      "label": "POSITIVE",
      "score": 0.89,
      "confidence": 0.89,
      "model_used": "...",
      "processed_at": "..."
    },
    {
      "label": "NEGATIVE", 
      "score": 0.78,
      "confidence": 0.78,
      "model_used": "...",
      "processed_at": "..."
    },
    null
  ],
  "total_texts": 3,
  "success_count": 2,
  "failed_count": 1
}
```

#### Get Sentiment Statistics
```http
GET /sentiments/stats
```

#### Get Team Report
```http
GET /sentiments/report/{teamId}?start_date=2024-01-01&end_date=2024-01-31
```

**Response:**
```json
{
  "team_id": "...",
  "team_name": "Galatasaray",
  "period": {
    "start_date": "2024-01-01T00:00:00Z",
    "end_date": "2024-01-31T23:59:59Z",
    "label": "2024-01-01 - 2024-01-31"
  },
  "total_analyzed": 500,
  "sentiment_counts": {
    "POSITIVE": 300,
    "NEGATIVE": 100,
    "NEUTRAL": 100
  },
  "average_sentiment": 0.68,
  "trend_analysis": {
    "direction": "improving",
    "change_percent": 12.5
  },
  "top_keywords": [],
  "hourly_distribution": {},
  "source_breakdown": {},
  "generated_at": "2024-01-01T12:00:00Z"
}
```

---

### Teams

#### List Teams
```http
GET /teams
```

**Response:**
```json
{
  "teams": [
    {
      "id": "...",
      "name": "Galatasaray",
      "slug": "galatasaray",
      "league": "Süper Lig",
      "country": "Turkey",
      "logo": "",
      "colors": ["#FFA500", "#8B0000"],
      "keywords": ["galatasaray", "gala", "gs", "aslan"],
      "subreddits": ["galatasaray"],
      "is_active": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ],
  "count": 4
}
```

#### Create Team
```http
POST /teams
```

**Request Body:**
```json
{
  "name": "Yeni Takım",
  "slug": "yeni-takim",
  "league": "Süper Lig",
  "country": "Turkey",
  "colors": ["#FF0000", "#0000FF"],
  "keywords": ["yeni", "takım", "yt"],
  "subreddits": ["yenitakim"]
}
```

#### Get Team
```http
GET /teams/{id}
```

#### Update Team
```http
PUT /teams/{id}
```

#### Get Team Sentiment
```http
GET /teams/{id}/sentiment?start_date=2024-01-01&end_date=2024-01-31
```

#### Get Team Stats
```http
GET /teams/{id}/stats
```

#### Seed Turkish Teams
```http
POST /teams/seed
```

---

### Webhooks (n8n)

#### Create Comment via Webhook
```http
POST /webhook/comment
Headers: X-API-Key: your_api_secret
```

#### Save Sentiment via Webhook  
```http
POST /webhook/sentiment
Headers: X-API-Key: your_api_secret
```

#### Get Unprocessed Comments
```http
GET /webhook/unprocessed?limit=50
Headers: X-API-Key: your_api_secret
```

---

### Reddit Integration

#### Collect from All Subreddits
```http
POST /reddit/collect
```

#### Collect from Specific Subreddit
```http
POST /reddit/subreddit/{name}
```

---

## Error Codes

- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden  
- `404` - Not Found
- `409` - Conflict (Duplicate)
- `429` - Too Many Requests
- `500` - Internal Server Error

## Rate Limiting

- **Public endpoints**: 100 requests/minute per IP
- **Webhook endpoints**: 1000 requests/minute with API key
- **Analysis endpoints**: 50 requests/minute per IP

## Examples

### JavaScript/Node.js
```javascript
const axios = require('axios');

// Get comments
const response = await axios.get('http://localhost:8080/api/v1/comments', {
  params: {
    page: 1,
    limit: 20,
    sentiment: 'POSITIVE'
  }
});

// Analyze text
const analysis = await axios.post('http://localhost:8080/api/v1/sentiments/analyze', {
  text: 'Bu maç harika geçti!',
  language: 'tr'
});
```

### cURL
```bash
# Get dashboard data
curl -X GET "http://localhost:8080/api/v1/dashboard/data"

# Analyze sentiment
curl -X POST "http://localhost:8080/api/v1/sentiments/analyze" \
  -H "Content-Type: application/json" \
  -d '{"text": "Maç çok güzeldi!", "language": "tr"}'

# Create comment via webhook
curl -X POST "http://localhost:8080/api/v1/webhook/comment" \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your_api_secret" \
  -d '{"source_id": "abc123", "source": "reddit", "text": "Test comment", "author": "test_user"}'
```

### Python
```python
import requests

# Get team comparison
response = requests.get('http://localhost:8080/api/v1/dashboard/comparison')
data = response.json()

# Batch analyze
response = requests.post('http://localhost:8080/api/v1/sentiments/analyze/batch', 
  json={'texts': ['Harika maç!', 'Kötü oynadılar.']})
results = response.json()
```