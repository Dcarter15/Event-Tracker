import React from 'react';
import { useWebSocket } from '../contexts/WebSocketContext';

const WebSocketStatus = () => {
  const { isConnected, reconnectAttempts, maxReconnectAttempts } = useWebSocket();

  return (
    <div
      style={{
        position: 'fixed',
        top: '10px',
        left: '10px',
        background: isConnected ? '#28a745' : '#dc3545',
        color: 'white',
        padding: '8px 12px',
        borderRadius: '4px',
        fontSize: '12px',
        zIndex: 9999
      }}
    >
      WebSocket: {isConnected ? 'Connected' : 'Disconnected'}
      {reconnectAttempts > 0 && ` (Retries: ${reconnectAttempts}/${maxReconnectAttempts})`}
    </div>
  );
};

export default WebSocketStatus;