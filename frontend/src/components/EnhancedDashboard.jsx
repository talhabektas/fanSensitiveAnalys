import React, { useState, useEffect } from 'react';

import { Tab } from '@headlessui/react';
import { 
  ChartBarIcon,
  SparklesIcon,
  LightBulbIcon,
  BeakerIcon,
  TagIcon,
  ChatBubbleLeftRightIcon
} from '@heroicons/react/24/outline';
import CategoryStats from './CategoryStats';
import AISummary from './AISummary';
import TrendInsights from './TrendInsights';
import GroqTest from './GroqTest';
import { groqService } from '../services/groqService';

function classNames(...classes) {
  return classes.filter(Boolean).join(' ');
}

const EnhancedDashboard = ({ selectedTeam }) => {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [teams, setTeams] = useState([]);
  const [currentTeamId, setCurrentTeamId] = useState(selectedTeam?.id || null);
  const [currentTeamName, setCurrentTeamName] = useState(selectedTeam?.name || 'Tüm Takımlar');

  // Takımları yükle
  useEffect(() => {
    const fetchTeams = async () => {
      try {
        console.log('Fetching teams...');
        const response = await fetch('http://localhost:8060/api/v1/teams');
        if (response.ok) {
          const data = await response.json();
          console.log('Teams API response:', data);
          console.log('Teams array:', data.teams);
          // API response: {teams: [...], count: x} formatında
          setTeams(data.teams || []);
        }
      } catch (err) {
        console.error('Failed to fetch teams:', err);
      }
    };
    fetchTeams();
  }, []);

  // selectedTeam prop değiştiğinde güncelle
  useEffect(() => {
    setCurrentTeamId(selectedTeam?.id || null);
    setCurrentTeamName(selectedTeam?.name || 'Tüm Takımlar');
  }, [selectedTeam]);

  const tabs = [
    {
      name: 'Kategori Analizi',
      icon: TagIcon,
      component: CategoryStats,
      color: 'text-blue-500',
      description: 'Yorumları kategorilere ayır ve analiz et'
    },
    {
      name: 'AI Özeti',
      icon: ChatBubbleLeftRightIcon,
      component: AISummary,
      color: 'text-purple-500',
      description: 'Günlük yorumların akıllı özeti'
    },
    {
      name: 'Trend İçgörüleri',
      icon: LightBulbIcon,
      component: TrendInsights,
      color: 'text-yellow-500',
      description: 'AI destekli trend analizi'
    },
    {
      name: 'Grok AI Test',
      icon: BeakerIcon,
      component: GroqTest,
      color: 'text-green-500',
      description: 'Grok AI özelliklerini test et'
    }
  ];

  return (
    <div className="w-full">
      {/* Header */}
      <div className="mb-8">
        <div className="flex items-center mb-2">
          <SparklesIcon className="h-8 w-8 text-indigo-500 mr-3" />
          <h1 className="text-3xl font-bold text-gray-900">Gelişmiş AI Dashboard</h1>
          <span className="ml-3 px-3 py-1 bg-indigo-100 text-indigo-800 text-sm rounded-full">
            Powered by Grok AI 🤖
          </span>
        </div>
        <p className="text-gray-600">
          Hibrit yapay zeka teknolojisi ile güçlendirilmiş sentiment analizi ve içgörüler
        </p>
      </div>

      {/* Team Selection */}
      <div className="mb-6 p-4 bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg border border-blue-200">
        <div className="flex items-center justify-between">
          <div className="flex items-center">
            <ChartBarIcon className="h-5 w-5 text-blue-500 mr-2" />
            <span className="font-medium text-blue-800">Analiz Kapsamı:</span>
          </div>
          
          <select
            value={currentTeamId || ''}
            onChange={(e) => {
              const selectedTeamId = e.target.value || null;
              const selectedTeam = Array.isArray(teams) ? teams.find(t => t._id === selectedTeamId) : null;
              const teamName = selectedTeam ? selectedTeam.name : 'Tüm Takımlar';
              
              console.log('Team selected:', { selectedTeamId, selectedTeam, teamName });
              
              setCurrentTeamId(selectedTeamId);
              setCurrentTeamName(teamName);
            }}
            className="ml-3 px-3 py-2 border border-blue-300 rounded-md text-sm font-medium text-blue-800 bg-white focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">Tüm Takımlar</option>
            {Array.isArray(teams) && teams.map((team, index) => {
              console.log('Team option:', team);
              return (
                <option key={team._id || team.id || index} value={team._id || team.id}>
                  {team.name}
                </option>
              );
            })}
          </select>
        </div>
      </div>

      {/* Tabs */}
      <Tab.Group selectedIndex={selectedIndex} onChange={setSelectedIndex}>
        <Tab.List className="flex space-x-1 rounded-xl bg-blue-900/20 p-1 mb-8">
          {tabs.map((tab, index) => {
            const Icon = tab.icon;
            return (
              <Tab
                key={tab.name}
                className={({ selected }) =>
                  classNames(
                    'w-full rounded-lg py-2.5 text-sm font-medium leading-5',
                    'ring-white ring-opacity-60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2',
                    selected
                      ? 'bg-white text-blue-700 shadow'
                      : 'text-blue-100 hover:bg-white/[0.12] hover:text-white'
                  )
                }
              >
                <div className="flex items-center justify-center space-x-2">
                  <Icon className="h-5 w-5" />
                  <span className="hidden sm:inline">{tab.name}</span>
                </div>
              </Tab>
            );
          })}
        </Tab.List>

        <Tab.Panels className="mt-2">
          {tabs.map((tab, index) => {
            const Component = tab.component;
            return (
              <Tab.Panel
                key={index}
                className={classNames(
                  'rounded-xl bg-white p-3',
                  'ring-white ring-opacity-60 ring-offset-2 ring-offset-blue-400 focus:outline-none focus:ring-2'
                )}
              >
                {/* Tab Description */}
                <div className="mb-4 p-3 bg-gray-50 rounded-lg">
                  <div className="flex items-center">
                    {React.createElement(tab.icon, { className: `h-5 w-5 ${tab.color} mr-2` })}
                    <span className="text-gray-700 text-sm">{tab.description}</span>
                  </div>
                </div>

                {/* Tab Content */}
                <Component 
                  teamId={currentTeamId} 
                  teamName={currentTeamName}
                />
              </Tab.Panel>
            );
          })}
        </Tab.Panels>
      </Tab.Group>

      {/* Footer Info */}
      <div className="mt-8 p-4 bg-gray-50 rounded-lg">
        <div className="flex items-center justify-between text-sm text-gray-600">
          <div className="flex items-center">
            <SparklesIcon className="h-4 w-4 mr-1" />
            <span>Bu özellikler Grok AI hibrit teknolojisi ile güçlendirilmiştir</span>
          </div>
          <div className="flex items-center space-x-4">
            <span>HuggingFace BERT</span>
            <span>+</span>
            <span>Grok AI</span>
            <span>=</span>
            <span className="font-medium text-indigo-600">%30 Daha Doğru</span>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EnhancedDashboard;