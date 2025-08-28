import React, { useState } from 'react';
import { Settings as SettingsIcon, Database, Key, Bell, Download } from 'lucide-react';

export const Settings = () => {
  const [activeTab, setActiveTab] = useState('general');

  const tabs = [
    { id: 'general', name: 'Genel', icon: SettingsIcon },
    { id: 'database', name: 'Veritabanı', icon: Database },
    { id: 'api', name: 'API Ayarları', icon: Key },
    { id: 'notifications', name: 'Bildirimler', icon: Bell },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-gray-900">Ayarlar</h2>
        <p className="text-gray-600">Sistem ayarlarını yönetin</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar */}
        <div className="lg:col-span-1">
          <nav className="space-y-1">
            {tabs.map((tab) => {
              const Icon = tab.icon;
              return (
                <button
                  key={tab.id}
                  onClick={() => setActiveTab(tab.id)}
                  className={`w-full flex items-center px-3 py-2 text-sm font-medium rounded-md ${
                    activeTab === tab.id
                      ? 'bg-primary-100 text-primary-700'
                      : 'text-gray-600 hover:bg-gray-50 hover:text-gray-900'
                  }`}
                >
                  <Icon className="mr-3 h-4 w-4" />
                  {tab.name}
                </button>
              );
            })}
          </nav>
        </div>

        {/* Content */}
        <div className="lg:col-span-3">
          {activeTab === 'general' && (
            <div className="card">
              <div className="card-header">
                <h3 className="text-lg font-medium text-gray-900">Genel Ayarlar</h3>
              </div>
              <div className="card-content space-y-6">
                <div>
                  <label className="label">Uygulama Adı</label>
                  <input
                    type="text"
                    defaultValue="Taraftar Duygu Analizi"
                    className="input"
                  />
                </div>
                <div>
                  <label className="label">Zaman Dilimi</label>
                  <select className="select">
                    <option value="Europe/Istanbul">Türkiye (UTC+3)</option>
                    <option value="UTC">UTC</option>
                  </select>
                </div>
                <div>
                  <label className="label">Dil</label>
                  <select className="select">
                    <option value="tr">Türkçe</option>
                    <option value="en">English</option>
                  </select>
                </div>
                <button className="btn btn-primary">Kaydet</button>
              </div>
            </div>
          )}

          {activeTab === 'database' && (
            <div className="card">
              <div className="card-header">
                <h3 className="text-lg font-medium text-gray-900">Veritabanı Ayarları</h3>
              </div>
              <div className="card-content space-y-6">
                <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                  <p className="text-yellow-800">
                    Veritabanı ayarları environment dosyasından yönetilir.
                  </p>
                </div>
                <div>
                  <label className="label">MongoDB Bağlantı Durumu</label>
                  <div className="flex items-center space-x-2">
                    <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                    <span className="text-green-700">Bağlı</span>
                  </div>
                </div>
                <button className="btn btn-secondary">
                  <Download className="h-4 w-4 mr-2" />
                  Veritabanı Yedeği Al
                </button>
              </div>
            </div>
          )}

          {activeTab === 'api' && (
            <div className="space-y-6">
              <div className="card">
                <div className="card-header">
                  <h3 className="text-lg font-medium text-gray-900">API Anahtarları</h3>
                </div>
                <div className="card-content space-y-6">
                  <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <p className="text-blue-800">
                      API anahtarları güvenlik nedeniyle .env dosyasından yönetilir.
                    </p>
                  </div>
                  
                  <div>
                    <h4 className="font-medium text-gray-900 mb-3">API Durumları</h4>
                    <div className="space-y-3">
                      <div className="flex items-center justify-between p-3 border rounded-lg">
                        <span className="font-medium">Reddit API</span>
                        <span className="text-green-600">Aktif</span>
                      </div>
                      <div className="flex items-center justify-between p-3 border rounded-lg">
                        <span className="font-medium">HuggingFace API</span>
                        <span className="text-green-600">Aktif</span>
                      </div>
                      <div className="flex items-center justify-between p-3 border rounded-lg">
                        <span className="font-medium">YouTube API</span>
                        <span className="text-gray-400">Yapılandırılmamış</span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'notifications' && (
            <div className="card">
              <div className="card-header">
                <h3 className="text-lg font-medium text-gray-900">Bildirim Ayarları</h3>
              </div>
              <div className="card-content space-y-6">
                <div className="space-y-4">
                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900">Sistem Bildirimleri</h4>
                      <p className="text-sm text-gray-500">Sistem durumu güncellemeleri</p>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" className="sr-only peer" defaultChecked />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                    </label>
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900">Yeni Yorum Bildirimleri</h4>
                      <p className="text-sm text-gray-500">Yeni yorumlar geldiğinde bildir</p>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" className="sr-only peer" />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                    </label>
                  </div>

                  <div className="flex items-center justify-between">
                    <div>
                      <h4 className="font-medium text-gray-900">Analiz Tamamlama</h4>
                      <p className="text-sm text-gray-500">Duygu analizi tamamlandığında bildir</p>
                    </div>
                    <label className="relative inline-flex items-center cursor-pointer">
                      <input type="checkbox" className="sr-only peer" defaultChecked />
                      <div className="w-11 h-6 bg-gray-200 peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-blue-300 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-gray-300 after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-blue-600"></div>
                    </label>
                  </div>
                </div>

                <button className="btn btn-primary">Ayarları Kaydet</button>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};