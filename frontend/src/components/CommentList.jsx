import React, { useState } from 'react';
import { useComments } from '../hooks/useComments';
import { useSentiment } from '../hooks/useSentiment';
import { TeamLogo } from '../utils/teamLogos.jsx';
import { 
  Search, 
  Filter, 
  Download, 
  CheckSquare, 
  Square,
  Eye,
  Calendar,
  MessageSquare,
  TrendingUp
} from 'lucide-react';
import { format } from 'date-fns';
import { tr } from 'date-fns/locale';

export const CommentList = () => {
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTeam, setSelectedTeam] = useState('');
  const [selectedSource, setSelectedSource] = useState('');
  const [selectedSentiment, setSelectedSentiment] = useState('');
  const [showFilters, setShowFilters] = useState(false);

  const {
    comments,
    stats,
    isLoading,
    totalComments,
    totalPages,
    filters,
    selectedComments,
    setPage,
    setLimit,
    updateFilters,
    toggleCommentSelection,
    selectAllComments,
    clearSelection,
    markSelectedAsProcessed,
    isBulkUpdating,
  } = useComments();

  const { formatSentimentLabel, getSentimentColor } = useSentiment();

  // Handle search
  const handleSearch = (e) => {
    e.preventDefault();
    updateFilters({ search: searchQuery });
  };

  // Handle filter changes
  const handleFilterChange = () => {
    updateFilters({
      team_id: selectedTeam || undefined,
      source: selectedSource || undefined,
      sentiment: selectedSentiment || undefined,
    });
  };

  // Clear all filters
  const clearAllFilters = () => {
    setSearchQuery('');
    setSelectedTeam('');
    setSelectedSource('');
    setSelectedSentiment('');
    updateFilters({
      search: undefined,
      team_id: undefined,
      source: undefined,
      sentiment: undefined,
    });
  };

  // Format date
  const formatDate = (dateString) => {
    return format(new Date(dateString), 'dd MMM yyyy, HH:mm', { locale: tr });
  };

  // Get sentiment badge classes
  const getSentimentBadgeClass = (sentiment) => {
    if (!sentiment) return 'badge bg-gray-100 text-gray-800';
    
    switch (sentiment.label?.toUpperCase()) {
      case 'POSITIVE':
        return 'badge-positive';
      case 'NEGATIVE':
        return 'badge-negative';
      case 'NEUTRAL':
        return 'badge-neutral';
      default:
        return 'badge bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Yorumlar</h2>
          <p className="text-gray-600">
            Toplam {totalComments.toLocaleString()} yorum
          </p>
        </div>
        <div className="flex items-center space-x-3">
          <button
            onClick={() => setShowFilters(!showFilters)}
            className="btn btn-secondary flex items-center space-x-2"
          >
            <Filter className="h-4 w-4" />
            <span>Filtreler</span>
          </button>
          <button className="btn btn-secondary flex items-center space-x-2">
            <Download className="h-4 w-4" />
            <span>Dışa Aktar</span>
          </button>
        </div>
      </div>

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
          <div className="stat-card">
            <div className="flex items-center">
              <MessageSquare className="h-8 w-8 text-blue-600" />
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-500">Toplam</p>
                <p className="text-xl font-semibold text-gray-900">
                  {stats.total_comments?.toLocaleString() || 0}
                </p>
              </div>
            </div>
          </div>
          <div className="stat-card">
            <div className="flex items-center">
              <CheckSquare className="h-8 w-8 text-green-600" />
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-500">İşlenen</p>
                <p className="text-xl font-semibold text-gray-900">
                  {stats.processed_comments?.toLocaleString() || 0}
                </p>
              </div>
            </div>
          </div>
          <div className="stat-card">
            <div className="flex items-center">
              <Eye className="h-8 w-8 text-yellow-600" />
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-500">Bekleyen</p>
                <p className="text-xl font-semibold text-gray-900">
                  {stats.unprocessed_comments?.toLocaleString() || 0}
                </p>
              </div>
            </div>
          </div>
          <div className="stat-card">
            <div className="flex items-center">
              <TrendingUp className="h-8 w-8 text-purple-600" />
              <div className="ml-3">
                <p className="text-sm font-medium text-gray-500">Analiz Edilen</p>
                <p className="text-xl font-semibold text-gray-900">
                  {Object.values(stats.sentiment_breakdown || {})
                    .reduce((sum, val) => sum + val, 0)
                    .toLocaleString()}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Search and Filters */}
      <div className="card">
        <div className="card-content">
          <form onSubmit={handleSearch} className="flex space-x-4">
            <div className="flex-1">
              <div className="relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400 h-4 w-4" />
                <input
                  type="text"
                  placeholder="Yorumlarda ara..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="input pl-10"
                />
              </div>
            </div>
            <button type="submit" className="btn btn-primary">
              Ara
            </button>
          </form>

          {showFilters && (
            <div className="mt-4 pt-4 border-t border-gray-200">
              <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                <div>
                  <label className="label">Takım</label>
                  <select
                    value={selectedTeam}
                    onChange={(e) => setSelectedTeam(e.target.value)}
                    className="select"
                  >
                    <option value="">Tüm Takımlar</option>
                    <option value="galatasaray">Galatasaray</option>
                    <option value="fenerbahce">Fenerbahçe</option>
                    <option value="besiktas">Beşiktaş</option>
                    <option value="trabzonspor">Trabzonspor</option>
                  </select>
                </div>
                <div>
                  <label className="label">Kaynak</label>
                  <select
                    value={selectedSource}
                    onChange={(e) => setSelectedSource(e.target.value)}
                    className="select"
                  >
                    <option value="">Tüm Kaynaklar</option>
                    <option value="reddit">Reddit</option>
                    <option value="twitter">Twitter</option>
                    <option value="youtube">YouTube</option>
                  </select>
                </div>
                <div>
                  <label className="label">Duygu</label>
                  <select
                    value={selectedSentiment}
                    onChange={(e) => setSelectedSentiment(e.target.value)}
                    className="select"
                  >
                    <option value="">Tüm Duygular</option>
                    <option value="POSITIVE">Pozitif</option>
                    <option value="NEUTRAL">Nötr</option>
                    <option value="NEGATIVE">Negatif</option>
                  </select>
                </div>
                <div className="flex items-end space-x-2">
                  <button
                    type="button"
                    onClick={handleFilterChange}
                    className="btn btn-primary flex-1"
                  >
                    Filtrele
                  </button>
                  <button
                    type="button"
                    onClick={clearAllFilters}
                    className="btn btn-secondary"
                  >
                    Temizle
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Bulk Actions */}
      {selectedComments.length > 0 && (
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
          <div className="flex items-center justify-between">
            <p className="text-blue-800">
              {selectedComments.length} yorum seçildi
            </p>
            <div className="flex items-center space-x-3">
              <button
                onClick={clearSelection}
                className="text-blue-600 hover:text-blue-800 text-sm"
              >
                Seçimi Kaldır
              </button>
              <button
                onClick={markSelectedAsProcessed}
                disabled={isBulkUpdating}
                className="btn btn-success"
              >
                {isBulkUpdating ? 'İşleniyor...' : 'İşlendi Olarak İşaretle'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Comments Table */}
      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <table className="table">
            <thead className="table-header">
              <tr>
                <th className="table-header-cell w-12">
                  <button
                    onClick={selectAllComments}
                    className="text-gray-500 hover:text-gray-700"
                  >
                    <CheckSquare className="h-4 w-4" />
                  </button>
                </th>
                <th className="table-header-cell">Yorum</th>
                <th className="table-header-cell">Takım</th>
                <th className="table-header-cell">Yazar</th>
                <th className="table-header-cell">Kaynak</th>
                <th className="table-header-cell">Duygu</th>
                <th className="table-header-cell">Link</th>
                <th className="table-header-cell">Tarih</th>
              </tr>
            </thead>
            <tbody className="table-body">
              {isLoading ? (
                // Loading skeleton
                [...Array(5)].map((_, i) => (
                  <tr key={i} className="animate-pulse">
                    <td className="table-cell"><div className="h-4 w-4 bg-gray-200 rounded"></div></td>
                    <td className="table-cell"><div className="h-4 bg-gray-200 rounded w-3/4"></div></td>
                    <td className="table-cell"><div className="h-6 w-6 bg-gray-200 rounded-full"></div></td>
                    <td className="table-cell"><div className="h-4 bg-gray-200 rounded w-20"></div></td>
                    <td className="table-cell"><div className="h-4 bg-gray-200 rounded w-16"></div></td>
                    <td className="table-cell"><div className="h-4 bg-gray-200 rounded w-16"></div></td>
                    <td className="table-cell"><div className="h-4 bg-gray-200 rounded w-8"></div></td>
                    <td className="table-cell"><div className="h-4 bg-gray-200 rounded w-24"></div></td>
                  </tr>
                ))
              ) : comments.length > 0 ? (
                comments.map((comment) => (
                  <tr key={comment.id} className="hover:bg-gray-50">
                    <td className="table-cell">
                      <button
                        onClick={() => toggleCommentSelection(comment.id)}
                        className="text-gray-500 hover:text-gray-700"
                      >
                        {selectedComments.includes(comment.id) ? (
                          <CheckSquare className="h-4 w-4 text-blue-600" />
                        ) : (
                          <Square className="h-4 w-4" />
                        )}
                      </button>
                    </td>
                    <td className="table-cell">
                      <div className="max-w-xs">
                        <p className="truncate-2-lines text-sm text-gray-900">
                          {comment.text}
                        </p>
                        {comment.score !== undefined && (
                          <p className="text-xs text-gray-500 mt-1">
                            Skor: {comment.score}
                          </p>
                        )}
                      </div>
                    </td>
                    <td className="table-cell">
                      <div className="flex items-center">
                        <TeamLogo teamName={comment.team_name} size={24} />
                      </div>
                    </td>
                    <td className="table-cell">
                      <span className="text-sm text-gray-900">
                        {comment.author || 'Anonim'}
                      </span>
                    </td>
                    <td className="table-cell">
                      <span className="badge bg-gray-100 text-gray-800 capitalize">
                        {comment.source}
                      </span>
                    </td>
                    <td className="table-cell">
                      {comment.sentiment ? (
                        <span className={getSentimentBadgeClass(comment.sentiment)}>
                          {formatSentimentLabel(comment.sentiment.label)}
                        </span>
                      ) : (
                        <span className="badge bg-gray-100 text-gray-500">
                          Analiz Edilmemiş
                        </span>
                      )}
                    </td>
                    <td className="table-cell">
                      {comment.url ? (
                        <a
                          href={comment.url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-600 hover:text-blue-800 text-sm"
                          title="Kaynağa git"
                        >
                          🔗
                        </a>
                      ) : (
                        <span className="text-gray-400 text-sm">-</span>
                      )}
                    </td>
                    <td className="table-cell">
                      <span className="text-sm text-gray-500">
                        {formatDate(comment.created_at)}
                      </span>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan="8" className="text-center py-8">
                    <MessageSquare className="h-12 w-12 text-gray-400 mx-auto mb-2" />
                    <p className="text-gray-500">Yorum bulunamadı</p>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Pagination */}
        {totalPages > 1 && (
          <div className="border-t border-gray-200 px-6 py-3 flex items-center justify-between">
            <div className="flex items-center text-sm text-gray-500">
              Toplam {totalComments.toLocaleString()} yorumdan {' '}
              {((filters.page - 1) * filters.limit + 1).toLocaleString()}-
              {Math.min(filters.page * filters.limit, totalComments).toLocaleString()} arası gösteriliyor
            </div>
            <div className="flex items-center space-x-2">
              <button
                onClick={() => setPage(Math.max(1, filters.page - 1))}
                disabled={filters.page === 1}
                className="btn btn-secondary disabled:opacity-50"
              >
                Önceki
              </button>
              <span className="text-sm text-gray-700">
                {filters.page} / {totalPages}
              </span>
              <button
                onClick={() => setPage(Math.min(totalPages, filters.page + 1))}
                disabled={filters.page === totalPages}
                className="btn btn-secondary disabled:opacity-50"
              >
                Sonraki
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};