import React, { useState, useEffect } from 'react';
import { useQuery } from 'react-query';
import { dashboardService } from '../services/dashboardService';
import { Activity, Wifi, WifiOff, RefreshCw } from 'lucide-react';

export const RealtimeMonitor = () => {
  const [isConnected, setIsConnected] = useState(true);
  const [lastUpdate, setLastUpdate] = useState(new Date());

  // Real-time data polling
  const { data: realtimeData, refetch } = useQuery(
    'realtime-updates',
    dashboardService.getRealTimeUpdates,
    {
      refetchInterval: 5000, // 5 seconds
      refetchIntervalInBackground: true,
      onSuccess: (data) => {
        if (data) {
          setLastUpdate(new Date());
          setIsConnected(true);
        }
      },
      onError: () => {
        setIsConnected(false);
      },
    }
  );

  const formatLastUpdate = () => {
    return lastUpdate.toLocaleTimeString('tr-TR');
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Canlı İzleme</h2>
          <p className="text-gray-600">Gerçek zamanlı sistem durumu ve veriler</p>
        </div>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            {isConnected ? (
              <Wifi className="h-5 w-5 text-green-600" />
            ) : (
              <WifiOff className="h-5 w-5 text-red-600" />
            )}
            <span className="text-sm text-gray-600">
              {isConnected ? 'Bağlı' : 'Bağlantı Kesildi'}
            </span>
          </div>
          <button
            onClick={() => refetch()}
            className="btn btn-secondary flex items-center space-x-2"
          >
            <RefreshCw className="h-4 w-4" />
            <span>Yenile</span>
          </button>
        </div>
      </div>

      {/* Status Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="card p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Toplam Yorum</p>
              <p className="text-xl font-bold text-gray-900">
                {realtimeData?.comments?.total?.toLocaleString() || 0}
              </p>
            </div>
            <Activity className="h-8 w-8 text-blue-600" />
          </div>
        </div>

        <div className="card p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">İşlenmemiş</p>
              <p className="text-xl font-bold text-gray-900">
                {realtimeData?.comments?.unprocessed?.toLocaleString() || 0}
              </p>
            </div>
            <div className="w-3 h-3 bg-yellow-500 rounded-full animate-pulse"></div>
          </div>
        </div>

        <div className="card p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Pozitif</p>
              <p className="text-xl font-bold text-green-600">
                {realtimeData?.sentiments?.breakdown?.POSITIVE?.toLocaleString() || 0}
              </p>
            </div>
            <div className="w-3 h-3 bg-green-500 rounded-full"></div>
          </div>
        </div>

        <div className="card p-4">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm text-gray-500">Negatif</p>
              <p className="text-xl font-bold text-red-600">
                {realtimeData?.sentiments?.breakdown?.NEGATIVE?.toLocaleString() || 0}
              </p>
            </div>
            <div className="w-3 h-3 bg-red-500 rounded-full"></div>
          </div>
        </div>
      </div>

      {/* System Status */}
      <div className="card">
        <div className="card-header">
          <h3 className="text-lg font-medium text-gray-900">Sistem Durumu</h3>
        </div>
        <div className="card-content">
          <div className="space-y-4">
            <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                <span className="font-medium text-green-800">API Sunucusu</span>
              </div>
              <span className="text-green-600 text-sm">Aktif</span>
            </div>

            <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                <span className="font-medium text-green-800">MongoDB</span>
              </div>
              <span className="text-green-600 text-sm">Bağlı</span>
            </div>

            <div className="flex items-center justify-between p-3 bg-yellow-50 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="w-3 h-3 bg-yellow-500 rounded-full animate-pulse"></div>
                <span className="font-medium text-yellow-800">Reddit API</span>
              </div>
              <span className="text-yellow-600 text-sm">Rate Limited</span>
            </div>

            <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                <span className="font-medium text-green-800">HuggingFace API</span>
              </div>
              <span className="text-green-600 text-sm">Aktif</span>
            </div>

            <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                <span className="font-medium text-green-800">YouTube API</span>
              </div>
              <span className="text-green-600 text-sm">Bağlı</span>
            </div>

            <div className="flex items-center justify-between p-3 bg-green-50 rounded-lg">
              <div className="flex items-center space-x-3">
                <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                <span className="font-medium text-green-800">Grok AI</span>
              </div>
              <span className="text-green-600 text-sm">Aktif</span>
            </div>
          </div>
        </div>
      </div>

      {/* Last Update Info */}
      <div className="text-center text-sm text-gray-500">
        Son güncelleme: {formatLastUpdate()}
      </div>
    </div>
  );
};