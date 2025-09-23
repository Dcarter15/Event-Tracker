import React, { useState, useEffect } from 'react';
import { Dropdown, Badge, ListGroup, Button, Spinner } from 'react-bootstrap';
import { useWebSocket } from '../contexts/WebSocketContext';
import './NotificationCenter.css';

// Generate or get session ID for consistent user identification
const getSessionId = () => {
  let sessionId = localStorage.getItem('notification-session-id');
  if (!sessionId) {
    sessionId = 'user_' + Math.random().toString(36).substr(2, 16);
    localStorage.setItem('notification-session-id', sessionId);
  }
  return sessionId;
};

const NotificationCenter = () => {
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(false);
  const [show, setShow] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [offset, setOffset] = useState(0);
  const [viewMode, setViewMode] = useState('unread'); // 'unread' or 'read'
  const limit = 20;

  // Get notification count from WebSocket context
  const { notificationCount } = useWebSocket();


  // Fetch notifications with pagination
  const fetchNotifications = async (reset = false) => {
    setLoading(true);
    try {
      const currentOffset = reset ? 0 : offset;
      const endpoint = viewMode === 'unread' ? '/api/notifications' : '/api/notifications/read';
      const response = await fetch(`${endpoint}?limit=${limit}&offset=${currentOffset}`, {
        headers: {
          'X-Session-ID': getSessionId(),
        },
      });
      const data = await response.json();

      if (reset) {
        setNotifications(data);
        setOffset(data.length);
      } else {
        setNotifications(prev => [...prev, ...data]);
        setOffset(prev => prev + data.length);
      }

      // Check if there are more notifications
      setHasMore(data.length === limit);
    } catch (error) {
      console.error('Error fetching notifications:', error);
    } finally {
      setLoading(false);
    }
  };

  // Load more notifications
  const loadMore = () => {
    if (!loading && hasMore) {
      fetchNotifications(false);
    }
  };

  // Clear all notifications for current user
  const clearNotifications = async () => {
    try {
      const response = await fetch('/api/notifications/clear', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Session-ID': getSessionId(),
        },
      });

      if (response.ok) {
        const result = await response.json();
        console.log(`Cleared ${result.cleared} notifications`);

        // Reset the notification list and refresh the count
        setNotifications([]);
        setOffset(0);
        setHasMore(false);

        // The WebSocket will automatically update the count
      } else {
        console.error('Failed to clear notifications');
      }
    } catch (error) {
      console.error('Error clearing notifications:', error);
    }
  };

  // Mark individual notification as read
  const markNotificationAsRead = async (notificationId) => {
    try {
      const response = await fetch('/api/notifications/mark-read', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Session-ID': getSessionId(),
        },
        body: JSON.stringify({ notification_id: notificationId }),
      });

      if (response.ok) {
        const result = await response.json();
        console.log(`Marked notification ${notificationId} as read:`, result.marked);

        // Remove the notification from the current list
        setNotifications(prev => prev.filter(n => n.id !== notificationId));

        // The WebSocket will automatically update the count
      } else {
        console.error('Failed to mark notification as read');
      }
    } catch (error) {
      console.error('Error marking notification as read:', error);
    }
  };

  // No need for initial notification count fetch since it comes from WebSocket

  // Switch between unread and read notifications
  const switchViewMode = (mode) => {
    if (mode !== viewMode) {
      setViewMode(mode);
      setNotifications([]);
      setOffset(0);
      setHasMore(true);
      // Fetch notifications for the new mode
      setTimeout(() => fetchNotifications(true), 0);
    }
  };

  // Fetch notifications when dropdown is opened
  const handleToggle = (isOpen) => {
    setShow(isOpen);
    if (isOpen && notifications.length === 0) {
      fetchNotifications(true);
    }
  };

  // Format timestamp for display
  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffInHours = (now - date) / (1000 * 60 * 60);

    if (diffInHours < 1) {
      return 'Just now';
    } else if (diffInHours < 24) {
      return `${Math.floor(diffInHours)}h ago`;
    } else {
      return date.toLocaleDateString();
    }
  };

  // Get priority class for styling
  const getPriorityClass = (priority) => {
    switch (priority) {
      case 'critical': return 'notification-critical';
      case 'normal': return 'notification-normal';
      default: return 'notification-normal';
    }
  };

  const CustomToggle = React.forwardRef(({ onClick }, ref) => (
    <Button
      ref={ref}
      variant="outline-primary"
      onClick={onClick}
      className="notification-toggle"
    >
      🔔
      {notificationCount > 0 && (
        <Badge bg="danger" className="notification-badge">
          {notificationCount > 99 ? '99+' : notificationCount}
        </Badge>
      )}
    </Button>
  ));

  const CustomMenu = React.forwardRef(({ children, style, className, 'aria-labelledby': labeledBy }, ref) => (
    <div
      ref={ref}
      style={style}
      className={`${className} notification-menu`}
      aria-labelledby={labeledBy}
    >
      <div className="notification-header">
        <div className="d-flex align-items-center justify-content-between w-100">
          <h6 className="mb-0">Notifications</h6>
          <div className="d-flex align-items-center gap-2">
            <div className="btn-group btn-group-sm">
              <Button
                variant={viewMode === 'unread' ? 'primary' : 'outline-primary'}
                size="sm"
                onClick={() => switchViewMode('unread')}
                className="view-mode-btn"
              >
                Unread ({notificationCount})
              </Button>
              <Button
                variant={viewMode === 'read' ? 'primary' : 'outline-primary'}
                size="sm"
                onClick={() => switchViewMode('read')}
                className="view-mode-btn"
              >
                Read
              </Button>
            </div>
            {viewMode === 'unread' && notificationCount > 0 && (
              <Button
                variant="outline-secondary"
                size="sm"
                onClick={clearNotifications}
                className="clear-notifications-btn"
              >
                Clear All
              </Button>
            )}
          </div>
        </div>
      </div>

      <div className="notification-list">
        {loading && notifications.length === 0 ? (
          <div className="text-center p-3">
            <Spinner animation="border" size="sm" />
            <div>Loading notifications...</div>
          </div>
        ) : notifications.length === 0 ? (
          <div className="text-center text-muted p-3">
            {viewMode === 'unread' ? 'No unread notifications' : 'No read notifications'}
          </div>
        ) : (
          <>
            <ListGroup variant="flush">
              {notifications.map((notification) => (
                <ListGroup.Item
                  key={notification.id}
                  className={`notification-item ${getPriorityClass(notification.priority)}`}
                >
                  <div className="notification-content">
                    <div className="notification-message">
                      {notification.message}
                    </div>
                    <div className="notification-meta">
                      <div className="d-flex align-items-center gap-2">
                        <small className="text-muted">
                          {formatTimestamp(notification.created_at)}
                        </small>
                        {notification.priority === 'critical' && (
                          <Badge bg="danger" size="sm">Critical</Badge>
                        )}
                      </div>
                      {viewMode === 'unread' && (
                        <Button
                          variant="outline-secondary"
                          size="sm"
                          onClick={() => markNotificationAsRead(notification.id)}
                          className="dismiss-btn"
                          title="Mark as read"
                        >
                          ✕
                        </Button>
                      )}
                    </div>
                  </div>
                </ListGroup.Item>
              ))}
            </ListGroup>

            {hasMore && (
              <div className="text-center p-2">
                <Button
                  variant="outline-secondary"
                  size="sm"
                  onClick={loadMore}
                  disabled={loading}
                >
                  {loading ? (
                    <>
                      <Spinner animation="border" size="sm" className="me-2" />
                      Loading...
                    </>
                  ) : (
                    'Load More'
                  )}
                </Button>
              </div>
            )}
          </>
        )}
      </div>
    </div>
  ));

  return (
    <Dropdown show={show} onToggle={handleToggle} className="notification-dropdown">
      <Dropdown.Toggle as={CustomToggle} />
      <Dropdown.Menu as={CustomMenu} />
    </Dropdown>
  );
};

export default NotificationCenter;