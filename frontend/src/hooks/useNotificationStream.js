import { useEffect, useState, useCallback } from 'react';
import { useQueryClient } from '@tanstack/react-query';

export function useNotificationStream(addToast) {
  const [notifications, setNotifications] = useState([]);
  const [unreadCount, setUnreadCount] = useState(0);
  const queryClient = useQueryClient();

  useEffect(() => {
    const eventSource = new EventSource('http://localhost:8080/api/v1/notifications/stream', {
      withCredentials: true,
    });

    const handleMessage = (event) => {
      if (!event.data || event.data.startsWith(':')) return;
      try {
        const payload = JSON.parse(event.data);
        const newNotif = {
          id: Date.now() + Math.random(),
          title: payload.title || 'Notification',
          message: payload.message || '',
          type: payload.type || 'INFO',
          timestamp: new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
          read: false,
        };

        setNotifications((prev) => [newNotif, ...prev]);
        setUnreadCount((prev) => prev + 1);

        // Auto-refresh React Query data across dashboard
        queryClient.invalidateQueries({ queryKey: ['wallets'] });
        queryClient.invalidateQueries({ queryKey: ['transactions'] });
        queryClient.invalidateQueries({ queryKey: ['cards'] });

        if (addToast) {
          const toastType = payload.type === 'CARD_PAYMENT' || payload.type === 'TRANSACTION' ? 'success' : 'info';
          addToast(payload.message || payload.title, toastType);
        }
      } catch (err) {
        console.error('Failed to parse SSE notification payload:', err);
      }
    };

    eventSource.onmessage = handleMessage;

    eventSource.onerror = (err) => {
      console.warn('Notification SSE connection notice:', err);
    };

    return () => {
      eventSource.close();
    };
  }, [queryClient, addToast]);

  const markAllAsRead = useCallback(() => {
    setUnreadCount(0);
    setNotifications((prev) => prev.map((n) => ({ ...n, read: true })));
  }, []);

  const clearNotifications = useCallback(() => {
    setNotifications([]);
    setUnreadCount(0);
  }, []);

  return {
    notifications,
    unreadCount,
    markAllAsRead,
    clearNotifications,
  };
}

