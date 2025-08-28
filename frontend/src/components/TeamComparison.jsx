import React, { useState } from 'react';
import { useQuery } from 'react-query';
import { teamService } from '../services/teamService';
import { TeamComparisonChart, SentimentDistributionChart } from './SentimentChart';
import { TeamLogo, getTeamColors } from '../utils/teamLogos.jsx';
import { Users, TrendingUp, Calendar } from 'lucide-react';

export const TeamComparison = () => {
  const [selectedPeriod, setSelectedPeriod] = useState('7');

  const { data: teamComparison, isLoading } = useQuery(
    ['team-comparison', selectedPeriod],
    () => teamService.getTeamComparison(parseInt(selectedPeriod)),
    {
      staleTime: 300000,
    }
  );

  const periods = [
    { value: '7', label: 'Son 7 Gün' },
    { value: '30', label: 'Son 30 Gün' },
    { value: '90', label: 'Son 3 Ay' },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Takım Karşılaştırması</h2>
          <p className="text-gray-600">Takımların duygu durumu karşılaştırması</p>
        </div>
        <div className="flex items-center space-x-3">
          <Calendar className="h-5 w-5 text-gray-400" />
          <select
            value={selectedPeriod}
            onChange={(e) => setSelectedPeriod(e.target.value)}
            className="select"
          >
            {periods.map(period => (
              <option key={period.value} value={period.value}>
                {period.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {[...Array(4)].map((_, i) => (
            <div key={i} className="card animate-pulse">
              <div className="h-64 bg-gray-200 rounded"></div>
            </div>
          ))}
        </div>
      ) : (
        <div className="space-y-6">
          <div className="card">
            <div className="card-header">
              <h3 className="text-lg font-medium text-gray-900">Genel Karşılaştırma</h3>
            </div>
            <div className="card-content">
              <TeamComparisonChart data={teamComparison?.teams || []} height={400} />
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-4 gap-6">
            {teamComparison?.teams?.map((team) => {
              const teamColors = getTeamColors(team.team_name);
              return (
                <div key={team.team_id} className="card relative overflow-hidden">
                  <div 
                    className="absolute top-0 left-0 right-0 h-1"
                    style={{ background: teamColors.background }}
                  ></div>
                  <div className="card-header">
                    <div className="flex items-center space-x-3">
                      <TeamLogo teamName={team.team_name} size={40} />
                      <div>
                        <h4 className="font-medium text-gray-900">{team.team_name}</h4>
                        <div className="text-xs text-gray-500">#{team.ranking} sırada</div>
                      </div>
                    </div>
                  </div>
                  <div className="card-content">
                    <div className="text-center mb-4">
                      <div 
                        className="text-3xl font-bold mb-1"
                        style={{ color: teamColors.primary }}
                      >
                        {(team.avg_sentiment * 100).toFixed(0)}%
                      </div>
                      <div className="text-sm text-gray-500">Ortalama Duygu Skoru</div>
                    </div>
                    <div className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <span className="text-gray-600">Toplam Yorum:</span>
                        <span className="font-medium">{team.total_comments}</span>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-gray-600">Trend:</span>
                        <span className={`font-medium ${
                          team.avg_sentiment >= 0.7 ? 'text-green-600' :
                          team.avg_sentiment >= 0.4 ? 'text-yellow-600' :
                          'text-red-600'
                        }`}>
                          {team.avg_sentiment >= 0.7 ? '📈 Pozitif' :
                           team.avg_sentiment >= 0.4 ? '➡️ Nötr' :
                           '📉 Negatif'}
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}
    </div>
  );
};