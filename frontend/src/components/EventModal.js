import React, { useState, useEffect } from 'react';
import { Modal, Button, Form, Row, Col, Alert } from 'react-bootstrap';

const EventModal = ({ show, handleClose, exerciseId, onEventCreated, initialDate = null, editingEvent = null }) => {
  const [formData, setFormData] = useState({
    name: '',
    startDate: initialDate || '',
    endDate: initialDate || '',
    type: 'milestone',
    priority: 'medium',
    poc: '',
    status: 'planned',
    description: '',
    location: ''
  });
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Populate form when editing an existing event
  useEffect(() => {
    if (editingEvent) {
      setFormData({
        name: editingEvent.name || '',
        startDate: editingEvent.start_date ? new Date(editingEvent.start_date).toISOString().slice(0, 16) : '',
        endDate: editingEvent.end_date ? new Date(editingEvent.end_date).toISOString().slice(0, 16) : '',
        type: editingEvent.type || 'milestone',
        priority: editingEvent.priority || 'medium',
        poc: editingEvent.poc || '',
        status: editingEvent.status || 'planned',
        description: editingEvent.description || '',
        location: editingEvent.location || ''
      });
    } else if (initialDate) {
      setFormData(prev => ({
        ...prev,
        startDate: initialDate,
        endDate: initialDate
      }));
    }
  }, [editingEvent, initialDate]);

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value
    }));
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    setIsSubmitting(true);
    setError('');

    try {
      const eventData = {
        ...formData,
        exercise_id: exerciseId,
        start_date: new Date(formData.startDate).toISOString(),
        end_date: new Date(formData.endDate).toISOString()
      };

      let response;
      if (editingEvent) {
        // Update existing event
        eventData.id = editingEvent.id;
        response = await fetch(`/api/events/${editingEvent.id}`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(eventData),
        });
      } else {
        // Create new event
        response = await fetch('/api/events', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(eventData),
        });
      }

      if (response.ok) {
        const result = editingEvent ? { status: 'success' } : await response.json();
        onEventCreated(result);
        handleClose();
        // Reset form
        setFormData({
          name: '',
          startDate: '',
          endDate: '',
          type: 'milestone',
          priority: 'medium',
          poc: '',
          status: 'planned',
          description: '',
          location: ''
        });
      } else {
        setError(`Failed to ${editingEvent ? 'update' : 'create'} event. Please try again.`);
      }
    } catch (error) {
      console.error(`Error ${editingEvent ? 'updating' : 'creating'} event:`, error);
      setError(`Failed to ${editingEvent ? 'update' : 'create'} event. Please try again.`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <Modal show={show} onHide={handleClose} size="lg">
      <Modal.Header closeButton>
        <Modal.Title>{editingEvent ? 'Edit Event' : 'Create New Event'}</Modal.Title>
      </Modal.Header>
      <Modal.Body>
        {error && <Alert variant="danger">{error}</Alert>}
        <Form onSubmit={handleSubmit}>
          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Event Name *</Form.Label>
                <Form.Control
                  type="text"
                  name="name"
                  value={formData.name}
                  onChange={handleInputChange}
                  required
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Type</Form.Label>
                <Form.Select
                  name="type"
                  value={formData.type}
                  onChange={handleInputChange}
                >
                  <option value="milestone">Milestone</option>
                  <option value="phase">Phase</option>
                  <option value="meeting">Meeting</option>
                  <option value="training">Training</option>
                  <option value="deployment">Deployment</option>
                  <option value="other">Other</option>
                </Form.Select>
              </Form.Group>
            </Col>
          </Row>

          <Row>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>Start Date *</Form.Label>
                <Form.Control
                  type="datetime-local"
                  name="startDate"
                  value={formData.startDate}
                  onChange={handleInputChange}
                  required
                />
              </Form.Group>
            </Col>
            <Col md={6}>
              <Form.Group className="mb-3">
                <Form.Label>End Date *</Form.Label>
                <Form.Control
                  type="datetime-local"
                  name="endDate"
                  value={formData.endDate}
                  onChange={handleInputChange}
                  required
                />
              </Form.Group>
            </Col>
          </Row>

          <Row>
            <Col md={4}>
              <Form.Group className="mb-3">
                <Form.Label>Priority</Form.Label>
                <Form.Select
                  name="priority"
                  value={formData.priority}
                  onChange={handleInputChange}
                >
                  <option value="low">Low</option>
                  <option value="medium">Medium</option>
                  <option value="high">High</option>
                </Form.Select>
              </Form.Group>
            </Col>
            <Col md={4}>
              <Form.Group className="mb-3">
                <Form.Label>Status</Form.Label>
                <Form.Select
                  name="status"
                  value={formData.status}
                  onChange={handleInputChange}
                >
                  <option value="planned">Planned</option>
                  <option value="in-progress">In Progress</option>
                  <option value="completed">Completed</option>
                  <option value="cancelled">Cancelled</option>
                </Form.Select>
              </Form.Group>
            </Col>
            <Col md={4}>
              <Form.Group className="mb-3">
                <Form.Label>POC</Form.Label>
                <Form.Control
                  type="text"
                  name="poc"
                  value={formData.poc}
                  onChange={handleInputChange}
                />
              </Form.Group>
            </Col>
          </Row>

          <Form.Group className="mb-3">
            <Form.Label>Location</Form.Label>
            <Form.Control
              type="text"
              name="location"
              value={formData.location}
              onChange={handleInputChange}
            />
          </Form.Group>

          <Form.Group className="mb-3">
            <Form.Label>Description</Form.Label>
            <Form.Control
              as="textarea"
              rows={3}
              name="description"
              value={formData.description}
              onChange={handleInputChange}
            />
          </Form.Group>
        </Form>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="secondary" onClick={handleClose}>
          Cancel
        </Button>
        <Button 
          variant="primary" 
          onClick={handleSubmit}
          disabled={isSubmitting || !formData.name || !formData.startDate || !formData.endDate}
        >
          {isSubmitting ? (editingEvent ? 'Updating...' : 'Creating...') : (editingEvent ? 'Update Event' : 'Create Event')}
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default EventModal;