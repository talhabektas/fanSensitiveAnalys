import React from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Layout } from './components/Layout';
import { Dashboard } from './components/Dashboard';
import { CommentList } from './components/CommentList';
import { TeamComparison } from './components/TeamComparison';
import { SentimentAnalyzer } from './components/SentimentAnalyzer';
import { RealtimeMonitor } from './components/RealtimeMonitor';
import { TrendAnalysis } from './components/TrendAnalysis';
import { Settings } from './components/Settings';
import EnhancedDashboard from './components/EnhancedDashboard';

function App() {
  return (
    <Layout>
      <Routes>
        <Route path="/" element={<Navigate to="/dashboard" replace />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/ai-dashboard" element={<EnhancedDashboard />} />
        <Route path="/comments" element={<CommentList />} />
        <Route path="/teams" element={<TeamComparison />} />
        <Route path="/analyze" element={<SentimentAnalyzer />} />
        <Route path="/monitor" element={<RealtimeMonitor />} />
        <Route path="/trends" element={<TrendAnalysis />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Layout>
  );
}

export default App;