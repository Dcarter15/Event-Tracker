import React, { createContext, useContext, useEffect, useState, useRef } from 'react';
import { toast } from 'react-toastify';

const WebSocketContext = createContext();

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error('useWebSocket must be used within a WebSocketProvider');
  }
  return context;
};

export const WebSocketProvider = ({ children }) => {
  const [socket, setSocket] = useState(null);
  const [isConnected, setIsConnected] = useState(false);
  const [reconnectAttempts, setReconnectAttempts] = useState(0);
  const [notificationCount, setNotificationCount] = useState(0);
  const reconnectTimeoutRef = useRef(null);
  const connectionRef = useRef(false); // Track if we're already trying to connect
  const maxReconnectAttempts = 5;
  const reconnectDelay = 3000; // 3 seconds

  const connect = () => {
    // Prevent duplicate connections
    if (connectionRef.current || socket) {
      console.log('🔌 Connection already in progress or exists, skipping...');
      return;
    }

    connectionRef.current = true;

    try {
      const wsUrl = `ws://localhost:8081/ws`;
      console.log('🔌 Attempting to connect to WebSocket:', wsUrl);
      const newSocket = new WebSocket(wsUrl);

      newSocket.onopen = () => {
        console.log('✅ WebSocket connected successfully');
        setIsConnected(true);
        setReconnectAttempts(0);
        setSocket(newSocket);
        connectionRef.current = false;
      };

      newSocket.onmessage = (event) => {
        try {
          const notification = JSON.parse(event.data);
          handleNotification(notification);
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      newSocket.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason);
        setIsConnected(false);
        setSocket(null);
        connectionRef.current = false;

        // Attempt to reconnect if not manually closed
        if (event.code !== 1000 && reconnectAttempts < maxReconnectAttempts) {
          reconnectTimeoutRef.current = setTimeout(() => {
            console.log(`Attempting to reconnect... (${reconnectAttempts + 1}/${maxReconnectAttempts})`);
            setReconnectAttempts(prev => prev + 1);
            connect();
          }, reconnectDelay);
        }
      };

      newSocket.onerror = (error) => {
        console.error('WebSocket error:', error);
        connectionRef.current = false;
      };

    } catch (error) {
      console.error('Error creating WebSocket connection:', error);
      connectionRef.current = false;
    }
  };

  const handleNotification = (notification) => {
    console.log('🔔 Received WebSocket notification:', notification);

    // Handle notification count updates
    if (notification.type === 'notification_count') {
      console.log('📊 Updating notification count:', notification.count);
      setNotificationCount(notification.count);
      return;
    }

    const { type, action, message, priority, entity_name } = notification;

    // Determine toast type based on priority
    let toastType = 'info';
    if (priority === 'critical') {
      toastType = 'error';
    } else if (priority === 'normal') {
      toastType = 'success';
    }

    // Show toast notification
    console.log('📢 Showing toast notification:', message, 'Type:', toastType);
    toast[toastType](message, {
      position: "top-right",
      autoClose: priority === 'critical' ? 8000 : 5000,
      hideProgressBar: false,
      closeOnClick: true,
      pauseOnHover: true,
      draggable: true,
    });
  };

  const disconnect = () => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
    }
    if (socket) {
      socket.close(1000, 'Manual disconnect');
    }
    connectionRef.current = false;
  };

  useEffect(() => {
    connect();

    // Cleanup on unmount
    return () => {
      disconnect();
    };
  }, []);

  const value = {
    socket,
    isConnected,
    reconnectAttempts,
    maxReconnectAttempts,
    notificationCount,
    connect,
    disconnect,
  };

  return (
    <WebSocketContext.Provider value={value}>
      {children}
    </WebSocketContext.Provider>
  );
};