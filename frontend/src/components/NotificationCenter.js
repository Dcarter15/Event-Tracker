import React, { useState, useEffect } from 'react';
import { Dropdown, Badge, ListGroup, Button, Spinner } from 'react-bootstrap';
import { useWebSocket } from '../contexts/WebSocketContext';
import './NotificationCenter.css';

const NotificationCenter = () => {
  const [notifications, setNotifications] = useState([]);
  const [loading, setLoading] = useState(false);
  const [show, setShow] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [offset, setOffset] = useState(0);
  const limit = 20;

  // Get notification count from WebSocket context
  const { notificationCount } = useWebSocket();


  // Fetch notifications with pagination
  const fetchNotifications = async (reset = false) => {
    setLoading(true);
    try {
      const currentOffset = reset ? 0 : offset;
      const response = await fetch(`/api/notifications?limit=${limit}&offset=${currentOffset}`);
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

  // No need for initial notification count fetch since it comes from WebSocket

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
        <h6 className="mb-0">Notifications</h6>
        <small className="text-muted">{notificationCount} total</small>
      </div>

      <div className="notification-list">
        {loading && notifications.length === 0 ? (
          <div className="text-center p-3">
            <Spinner animation="border" size="sm" />
            <div>Loading notifications...</div>
          </div>
        ) : notifications.length === 0 ? (
          <div className="text-center text-muted p-3">
            No notifications yet
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
                      <small className="text-muted">
                        {formatTimestamp(notification.created_at)}
                      </small>
                      {notification.priority === 'critical' && (
                        <Badge bg="danger" size="sm" className="ms-2">Critical</Badge>
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