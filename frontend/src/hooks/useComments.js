import { useState, useEffect, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import { commentService } from '../services/commentService';
import toast from 'react-hot-toast';

export const useComments = (initialFilters = {}) => {
  const [filters, setFilters] = useState({
    page: 1,
    limit: 20,
    sort_by: 'created_at',
    sort_order: 'desc',
    ...initialFilters,
  });
  
  const [selectedComments, setSelectedComments] = useState([]);
  const queryClient = useQueryClient();

  // Fetch comments with current filters
  const {
    data: commentsData,
    isLoading,
    error,
    refetch,
  } = useQuery(
    ['comments', filters],
    () => commentService.getComments(filters),
    {
      keepPreviousData: true,
      staleTime: 30000, // 30 seconds
      onError: (error) => {
        console.error('Error fetching comments:', error);
        toast.error('Yorumlar yüklenirken hata oluştu');
      },
    }
  );

  // Fetch comment statistics
  const {
    data: stats,
    isLoading: statsLoading,
    error: statsError,
  } = useQuery(
    'comment-stats',
    commentService.getCommentStats,
    {
      staleTime: 60000, // 1 minute
      onError: (error) => {
        console.error('Error fetching comment stats:', error);
      },
    }
  );

  // Create comment mutation
  const createCommentMutation = useMutation(commentService.createComment, {
    onSuccess: (data) => {
      toast.success('Yorum başarıyla eklendi');
      queryClient.invalidateQueries('comments');
      queryClient.invalidateQueries('comment-stats');
    },
    onError: (error) => {
      console.error('Error creating comment:', error);
      toast.error('Yorum eklenirken hata oluştu');
    },
  });

  // Update comment mutation
  const updateCommentMutation = useMutation(
    ({ id, data }) => commentService.updateComment(id, data),
    {
      onSuccess: () => {
        toast.success('Yorum başarıyla güncellendi');
        queryClient.invalidateQueries('comments');
        queryClient.invalidateQueries('comment-stats');
      },
      onError: (error) => {
        console.error('Error updating comment:', error);
        toast.error('Yorum güncellenirken hata oluştu');
      },
    }
  );

  // Bulk update processed mutation
  const bulkUpdateProcessedMutation = useMutation(
    commentService.bulkUpdateProcessed,
    {
      onSuccess: (data) => {
        toast.success(`${data.updated_count} yorum işlendi olarak işaretlendi`);
        queryClient.invalidateQueries('comments');
        queryClient.invalidateQueries('comment-stats');
        setSelectedComments([]);
      },
      onError: (error) => {
        console.error('Error bulk updating comments:', error);
        toast.error('Yorumlar güncellenirken hata oluştu');
      },
    }
  );

  // Filter functions
  const updateFilters = useCallback((newFilters) => {
    setFilters(prev => ({ ...prev, ...newFilters, page: 1 }));
  }, []);

  const setPage = useCallback((page) => {
    setFilters(prev => ({ ...prev, page }));
  }, []);

  const setLimit = useCallback((limit) => {
    setFilters(prev => ({ ...prev, limit, page: 1 }));
  }, []);

  const setSorting = useCallback((sortBy, sortOrder = 'desc') => {
    setFilters(prev => ({ ...prev, sort_by: sortBy, sort_order: sortOrder }));
  }, []);

  const filterByTeam = useCallback((teamId) => {
    updateFilters({ team_id: teamId });
  }, [updateFilters]);

  const filterBySource = useCallback((source) => {
    updateFilters({ source });
  }, [updateFilters]);

  const filterBySentiment = useCallback((sentiment) => {
    updateFilters({ sentiment });
  }, [updateFilters]);

  const filterByDateRange = useCallback((startDate, endDate) => {
    updateFilters({ start_date: startDate, end_date: endDate });
  }, [updateFilters]);

  const searchComments = useCallback((query) => {
    updateFilters({ search: query });
  }, [updateFilters]);

  const clearFilters = useCallback(() => {
    setFilters({
      page: 1,
      limit: 20,
      sort_by: 'created_at',
      sort_order: 'desc',
    });
  }, []);

  // Selection functions
  const toggleCommentSelection = useCallback((commentId) => {
    setSelectedComments(prev => 
      prev.includes(commentId)
        ? prev.filter(id => id !== commentId)
        : [...prev, commentId]
    );
  }, []);

  const selectAllComments = useCallback(() => {
    if (commentsData?.comments) {
      setSelectedComments(commentsData.comments.map(comment => comment.id));
    }
  }, [commentsData]);

  const clearSelection = useCallback(() => {
    setSelectedComments([]);
  }, []);

  // Actions
  const createComment = useCallback((commentData) => {
    return createCommentMutation.mutate(commentData);
  }, [createCommentMutation]);

  const updateComment = useCallback((id, data) => {
    return updateCommentMutation.mutate({ id, data });
  }, [updateCommentMutation]);

  const markAsProcessed = useCallback((commentIds) => {
    const ids = Array.isArray(commentIds) ? commentIds : [commentIds];
    return bulkUpdateProcessedMutation.mutate(ids);
  }, [bulkUpdateProcessedMutation]);

  const markSelectedAsProcessed = useCallback(() => {
    if (selectedComments.length > 0) {
      return bulkUpdateProcessedMutation.mutate(selectedComments);
    }
  }, [selectedComments, bulkUpdateProcessedMutation]);

  // Computed values
  const comments = commentsData?.comments || [];
  const totalComments = commentsData?.total || 0;
  const totalPages = commentsData?.total_pages || 0;
  const hasNextPage = filters.page < totalPages;
  const hasPreviousPage = filters.page > 1;

  // Loading states
  const isCreating = createCommentMutation.isLoading;
  const isUpdating = updateCommentMutation.isLoading;
  const isBulkUpdating = bulkUpdateProcessedMutation.isLoading;

  // Error states
  const hasError = error || statsError;

  return {
    // Data
    comments,
    stats,
    totalComments,
    totalPages,
    filters,
    selectedComments,
    
    // Loading states
    isLoading,
    statsLoading,
    isCreating,
    isUpdating,
    isBulkUpdating,
    
    // Error states
    error,
    statsError,
    hasError,
    
    // Pagination
    hasNextPage,
    hasPreviousPage,
    setPage,
    setLimit,
    
    // Filtering
    updateFilters,
    setSorting,
    filterByTeam,
    filterBySource,
    filterBySentiment,
    filterByDateRange,
    searchComments,
    clearFilters,
    
    // Selection
    toggleCommentSelection,
    selectAllComments,
    clearSelection,
    
    // Actions
    createComment,
    updateComment,
    markAsProcessed,
    markSelectedAsProcessed,
    refetch,
  };
};