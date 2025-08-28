// MongoDB initialization script for Docker
// This script runs when the MongoDB container starts for the first time

db = db.getSiblingDB('taraftar_analizi');

// Create initial collections
db.createCollection('teams');
db.createCollection('comments');
db.createCollection('sentiments');

// Insert Turkish teams
db.teams.insertMany([
  {
    name: 'Galatasaray',
    slug: 'galatasaray',
    league: 'Süper Lig',
    country: 'Turkey',
    colors: ['#FFA500', '#8B0000'],
    keywords: ['galatasaray', 'gala', 'gs', 'aslan', 'sarı-kırmızı'],
    subreddits: ['galatasaray'],
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Fenerbahçe',
    slug: 'fenerbahce',
    league: 'Süper Lig',
    country: 'Turkey',
    colors: ['#FFFF00', '#000080'],
    keywords: ['fenerbahçe', 'fenerbahce', 'fener', 'fb', 'kanarya', 'sarı-lacivert'],
    subreddits: ['fenerbahce'],
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Beşiktaş',
    slug: 'besiktas',
    league: 'Süper Lig',
    country: 'Turkey',
    colors: ['#000000', '#FFFFFF'],
    keywords: ['beşiktaş', 'besiktas', 'bjk', 'kartal', 'siyah-beyaz'],
    subreddits: ['besiktas'],
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  },
  {
    name: 'Trabzonspor',
    slug: 'trabzonspor',
    league: 'Süper Lig',
    country: 'Turkey',
    colors: ['#800080', '#000080'],
    keywords: ['trabzonspor', 'trabzon', 'ts', 'bordo-mavi'],
    subreddits: ['trabzonspor'],
    is_active: true,
    created_at: new Date(),
    updated_at: new Date()
  }
]);

// Create indexes for better performance
db.comments.createIndexes([
  { 'source_id': 1, 'source': 1 },
  { 'created_at': -1 },
  { 'team_id': 1 },
  { 'sentiment.label': 1 },
  { 'is_processed': 1 },
  { 'has_sentiment': 1 }
]);

db.sentiments.createIndexes([
  { 'comment_id': 1 },
  { 'team_id': 1, 'created_at': -1 },
  { 'label': 1 },
  { 'confidence': 1 }
]);

db.teams.createIndexes([
  { 'slug': 1 },
  { 'is_active': 1 }
]);

print('MongoDB initialization completed successfully!');
print('Created collections: teams, comments, sentiments');
print('Inserted Turkish football teams');
print('Created performance indexes');