import React, { useState, useEffect } from 'react';
import './GanttChart.css';
import {
  eachMonthOfInterval,
  eachWeekOfInterval,
  eachDayOfInterval,
  differenceInDays,
  format,
  startOfMonth,
  addYears,
  addMonths,
  startOfWeek,
  endOfWeek
} from 'date-fns';
import { Button, ButtonGroup, OverlayTrigger, Tooltip } from 'react-bootstrap';
import EventModal from './EventModal';

const GanttChart = ({ exercises, onExerciseClick }) => {
  const [viewMode, setViewMode] = useState('month'); // 'month', 'week', 'day'
  const [editingPOC, setEditingPOC] = useState({});
  const [pocValues, setPocValues] = useState({});
  const [showEventModal, setShowEventModal] = useState(false);
  const [selectedExerciseId, setSelectedExerciseId] = useState(null);
  const [clickedDate, setClickedDate] = useState(null);
  const [editingEvent, setEditingEvent] = useState(null);
  const [legendCollapsed, setLegendCollapsed] = useState(true);
  const [zoomLevel, setZoomLevel] = useState(1);
  const [showZoomControls, setShowZoomControls] = useState(false);
  const [zoomTimeout, setZoomTimeout] = useState(null);

  // Calculate exercise readiness percentage based on team statuses
  const calculateReadiness = (exercise) => {
    if (!exercise.divisions || exercise.divisions.length === 0) return null;
    
    let totalTeams = 0;
    let totalScore = 0;
    
    exercise.divisions.forEach(division => {
      if (division.teams && division.teams.length > 0) {
        division.teams.forEach(team => {
          totalTeams++;
          if (team.status === 'green') {
            totalScore += 1;
          } else if (team.status === 'yellow') {
            totalScore += 0.5;
          }
          // Red teams contribute 0
        });
      }
    });
    
    if (totalTeams === 0) return null;
    
    const percentage = (totalScore / totalTeams) * 100;
    return Math.round(percentage);
  };

  // Get readiness color based on percentage
  const getReadinessColor = (percentage) => {
    if (percentage >= 75) return '#28a745'; // Green
    if (percentage >= 50) return '#ffc107'; // Yellow
    return '#dc3545'; // Red
  };

  // Initialize POC values from exercises
  useEffect(() => {
    const initialPocValues = {};
    exercises.forEach(exercise => {
      initialPocValues[exercise.id] = exercise.exercise_event_poc || '';
    });
    setPocValues(initialPocValues);
  }, [exercises]);

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (zoomTimeout) {
        clearTimeout(zoomTimeout);
      }
    };
  }, [zoomTimeout]);

  // --- Date and Timeline Calculations ---
  let chartStartDate, chartEndDate, timelineHeaders;

  const today = new Date();

  switch (viewMode) {
    case 'day':
      chartStartDate = startOfMonth(today);
      chartEndDate = addMonths(chartStartDate, 3);
      timelineHeaders = eachDayOfInterval({ start: chartStartDate, end: chartEndDate });
      break;
    case 'week':
      chartStartDate = startOfWeek(startOfMonth(today));
      chartEndDate = addMonths(chartStartDate, 6);
      timelineHeaders = eachWeekOfInterval({ start: chartStartDate, end: chartEndDate });
      break;
    case 'month':
    default:
      chartStartDate = startOfMonth(today);
      chartEndDate = addYears(chartStartDate, 2);
      timelineHeaders = eachMonthOfInterval({ start: chartStartDate, end: chartEndDate });
      break;
  }

  const totalDays = differenceInDays(chartEndDate, chartStartDate);

  // Sort exercises by priority (high -> medium -> low)
  const priorityOrder = { 'high': 1, 'medium': 2, 'low': 3 };
  const sortedExercises = [...exercises].sort((a, b) => {
    const aPriority = a.priority || 'medium';
    const bPriority = b.priority || 'medium';
    return priorityOrder[aPriority] - priorityOrder[bPriority];
  });

  const getHeaderLabel = (date) => {
    switch (viewMode) {
      case 'day':
        return format(date, 'd');
      case 'week':
        const weekStart = format(date, 'MMM d');
        const weekEnd = format(endOfWeek(date), 'd');
        return `${weekStart} - ${weekEnd}`;
      case 'month':
      default:
        // Adjust label based on zoom level to prevent cramping
        if (zoomLevel < 0.7) {
          return format(date, 'MMM yy'); // Shorter format when zoomed out
        }
        return format(date, 'MMM yyyy');
    }
  };
  
  const getMonthLabel = (date) => {
    if (viewMode === 'day' && format(date, 'd') === '1') {
        return format(date, 'MMMM yyyy');
    }
    return null;
  }

  // Show zoom controls with auto-hide
  const showZoomControlsTemporary = () => {
    setShowZoomControls(true);
    
    // Clear existing timeout
    if (zoomTimeout) {
      clearTimeout(zoomTimeout);
    }
    
    // Set new timeout to hide controls after 3 seconds
    const timeout = setTimeout(() => {
      setShowZoomControls(false);
    }, 3000);
    
    setZoomTimeout(timeout);
  };

  // Zoom handling
  const handleWheel = (e) => {
    if (e.ctrlKey || e.metaKey) { // Ctrl+scroll or Cmd+scroll for zoom
      e.preventDefault();
      const delta = e.deltaY;
      const zoomFactor = 0.1;
      
      setZoomLevel(prevZoom => {
        const newZoom = delta > 0 ? prevZoom - zoomFactor : prevZoom + zoomFactor;
        return Math.max(0.5, Math.min(3, newZoom)); // Limit zoom between 0.5x and 3x
      });
      
      // Show zoom controls temporarily
      showZoomControlsTemporary();
    }
  };

  const handleZoomChange = (newZoom) => {
    setZoomLevel(newZoom);
    showZoomControlsTemporary();
  };

  const handlePOCEdit = (exerciseId) => {
    setEditingPOC({ ...editingPOC, [exerciseId]: true });
  };

  const handlePOCChange = (exerciseId, value) => {
    setPocValues({ ...pocValues, [exerciseId]: value });
  };

  const handlePOCSave = async (exerciseId) => {
    try {
      const exercise = exercises.find(ex => ex.id === exerciseId);
      const updatedExercise = {
        ...exercise,
        exercise_event_poc: pocValues[exerciseId]
      };

      const response = await fetch(`/api/exercises/${exerciseId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updatedExercise),
      });

      if (response.ok) {
        setEditingPOC({ ...editingPOC, [exerciseId]: false });
      } else {
        console.error('Failed to update exercise POC');
      }
    } catch (error) {
      console.error('Error updating exercise POC:', error);
    }
  };

  const handlePOCCancel = (exerciseId) => {
    const exercise = exercises.find(ex => ex.id === exerciseId);
    setPocValues({ ...pocValues, [exerciseId]: exercise.exercise_event_poc || '' });
    setEditingPOC({ ...editingPOC, [exerciseId]: false });
  };


  const handleTimelineClick = (exerciseId, date, event) => {
    event.stopPropagation();
    setSelectedExerciseId(exerciseId);
    setClickedDate(date.toISOString().split('T')[0]);
    setEditingEvent(null);
    setShowEventModal(true);
  };

  const handleEventClick = (event, eventData) => {
    event.stopPropagation();
    setSelectedExerciseId(eventData.exercise_id);
    setEditingEvent(eventData);
    setShowEventModal(true);
  };

  const handleEventCreated = (newEvent) => {
    // Refresh the exercises data to include the new event
    window.location.reload(); // Simple refresh for now, could be optimized
  };

  const handleCloseEventModal = () => {
    setShowEventModal(false);
    setSelectedExerciseId(null);
    setClickedDate(null);
    setEditingEvent(null);
  };

  const getTroubledDivisions = (exercise) => {
    const troubledDivisions = {
      red: [],
      yellow: []
    };

    if (exercise.divisions) {
      exercise.divisions.forEach(division => {
        if (division.teams) {
          const hasRedTeam = division.teams.some(team => team.status === 'red');
          const hasYellowTeam = division.teams.some(team => team.status === 'yellow');
          
          if (hasRedTeam) {
            troubledDivisions.red.push(division.name);
          } else if (hasYellowTeam) {
            troubledDivisions.yellow.push(division.name);
          }
        }
      });
    }

    return troubledDivisions;
  };

  const formatTroubledDivisions = (troubledDivisions) => {
    // Only show the worst status - red takes priority over yellow
    if (troubledDivisions.red.length > 0) {
      const redList = troubledDivisions.red.join(', ');
      return `In Trouble: ${redList} 🔴`;
    }
    
    if (troubledDivisions.yellow.length > 0) {
      const yellowList = troubledDivisions.yellow.join(', ');
      return `In Trouble: ${yellowList} 🟡`;
    }
    
    return null;
  };

  return (
    <div>
      <div className="gantt-controls mb-3">
        <ButtonGroup>
          <Button variant={viewMode === 'month' ? 'primary' : 'secondary'} onClick={() => setViewMode('month')}>Month</Button>
          <Button variant={viewMode === 'week' ? 'primary' : 'secondary'} onClick={() => setViewMode('week')}>Week</Button>
          <Button variant={viewMode === 'day' ? 'primary' : 'secondary'} onClick={() => setViewMode('day')}>Day</Button>
        </ButtonGroup>
      </div>

      {/* Zoom controls - positioned in upper right, visible during zoom operations */}
      {showZoomControls && (
        <div className="zoom-controls-fixed">
          <span className="zoom-label">Zoom: {Math.round(zoomLevel * 100)}%</span>
          <ButtonGroup size="sm" className="ms-2">
            <Button variant="outline-secondary" onClick={() => handleZoomChange(Math.max(0.5, zoomLevel - 0.1))}>−</Button>
            <Button variant="outline-secondary" onClick={() => handleZoomChange(1)}>Reset</Button>
            <Button variant="outline-secondary" onClick={() => handleZoomChange(Math.min(3, zoomLevel + 0.1))}>+</Button>
          </ButtonGroup>
          <small className="text-muted ms-2 d-block">Ctrl+Scroll to zoom</small>
        </div>
      )}

      <div className="gantt-chart" onWheel={handleWheel}>
        {/* Header */}
        <div className="gantt-header">
          <div className="gantt-poc-header">Exercise Event POC</div>
          <div className="gantt-row-header">Exercise</div>
          <div className="gantt-timeline" style={{ 
            gridTemplateColumns: `repeat(${timelineHeaders.length}, ${Math.max(60, 100 * zoomLevel)}px)`
          }}>
            {timelineHeaders.map(headerDate => (
              <div key={headerDate.toString()} className={`gantt-timeline-header gantt-${viewMode}`}>
                {getMonthLabel(headerDate) && <div className="gantt-month-label">{getMonthLabel(headerDate)}</div>}
                {getHeaderLabel(headerDate)}
              </div>
            ))}
          </div>
        </div>

        {/* Rows */}
        {sortedExercises.map(exercise => {
          const exerciseStartDate = new Date(exercise.start_date);
          const exerciseEndDate = new Date(exercise.end_date);

          const offset = differenceInDays(exerciseStartDate, chartStartDate);
          const duration = Math.max(1, differenceInDays(exerciseEndDate, exerciseStartDate));

          const left = (offset / totalDays) * 100;
          const width = (duration / totalDays) * 100;

          const troubledDivisions = getTroubledDivisions(exercise);
          const troubleText = formatTroubledDivisions(troubledDivisions);
          
          const renderTooltip = (props) => {
            const readiness = calculateReadiness(exercise);
            return (
              <Tooltip id={`tooltip-${exercise.id}`} {...props}>
                <strong>{exercise.name}</strong><br />
                {format(exerciseStartDate, 'MMM d, yyyy')} - {format(exerciseEndDate, 'MMM d, yyyy')}<br />
                Priority: {(exercise.priority || 'medium').charAt(0).toUpperCase() + (exercise.priority || 'medium').slice(1)}
                {readiness !== null && (
                  <>
                    <br />
                    Ability to support: <span style={{ color: getReadinessColor(readiness) }}>{readiness}%</span>
                  </>
                )}
                {troubleText && (
                  <>
                  <br />
                  {troubleText}
                </>
              )}
            </Tooltip>
          );
          };

          return (
            <div className="gantt-row" key={exercise.id}>
              <div className="gantt-poc-cell">
                {editingPOC[exercise.id] ? (
                  <div className="poc-edit-container">
                    <input
                      type="text"
                      value={pocValues[exercise.id] || ''}
                      onChange={(e) => handlePOCChange(exercise.id, e.target.value)}
                      className="poc-input"
                      autoFocus
                    />
                    <button onClick={() => handlePOCSave(exercise.id)} className="poc-save-btn">✓</button>
                    <button onClick={() => handlePOCCancel(exercise.id)} className="poc-cancel-btn">✗</button>
                  </div>
                ) : (
                  <div className="poc-display" onClick={() => handlePOCEdit(exercise.id)}>
                    {pocValues[exercise.id] || <span className="poc-placeholder">Click to add</span>}
                  </div>
                )}
              </div>
              <div className="gantt-row-header" onClick={() => onExerciseClick(exercise)}>
                {exercise.name}
              </div>
              <div className="gantt-bars" style={{ position: 'relative', minHeight: '100px' }}>
                {/* Timeline cells for clicking */}
                <div className="gantt-timeline-cells" style={{ 
                  position: 'absolute', 
                  top: 0, 
                  left: 0, 
                  width: '100%', 
                  height: '100%', 
                  display: 'grid',
                  gridTemplateColumns: `repeat(${timelineHeaders.length}, ${Math.max(60, 100 * zoomLevel)}px)`,
                  zIndex: 1
                }}>
                  {timelineHeaders.map((headerDate, index) => {
                    const cellDate = new Date(headerDate);
                    
                    return (
                      <div 
                        key={index} 
                        className="gantt-timeline-cell"
                        onClick={(e) => handleTimelineClick(exercise.id, cellDate, e)}
                        style={{ 
                          cursor: 'pointer',
                          minHeight: '100px',
                          border: 'none',
                          background: 'transparent'
                        }}
                        title="Click to add event"
                      />
                    );
                  })}
                </div>

                {/* Exercise bar */}
                {width > 0 && left + width > 0 && left < 100 && (
                  <OverlayTrigger
                    placement="top"
                    delay={{ show: 250, hide: 400 }}
                    overlay={renderTooltip}
                  >
                    <div
                      className="gantt-bar exercise-bar"
                      style={{
                        left: `${Math.max(0, left)}%`,
                        width: `${Math.min(width, 100 - Math.max(0, left))}%`,
                        top: '40px',
                        height: '25px',
                        zIndex: 2,
                        minWidth: `${Math.max(exercise.name.length * 12 + 40, 120)}px`
                      }}
                    >
                      <span className="gantt-bar-label">
                        {exercise.name}
                      </span>
                    </div>
                  </OverlayTrigger>
                )}

                {/* Event bars */}
                {exercise.events && exercise.events.map((event, eventIndex) => {
                  const eventStartDate = new Date(event.start_date);
                  const eventEndDate = new Date(event.end_date);
                  
                  const eventOffset = differenceInDays(eventStartDate, chartStartDate);
                  const eventDuration = Math.max(1, differenceInDays(eventEndDate, eventStartDate));
                  
                  const eventLeft = (eventOffset / totalDays) * 100;
                  const eventWidth = (eventDuration / totalDays) * 100;

                  // Check if event overlaps with exercise dates
                  const exerciseStartDate = new Date(exercise.start_date);
                  const exerciseEndDate = new Date(exercise.end_date);
                  const overlapsWithExercise = eventStartDate <= exerciseEndDate && eventEndDate >= exerciseStartDate;
                  
                  // Position events: if overlaps with exercise, stack above/below, otherwise use center
                  let topPosition;
                  if (overlapsWithExercise) {
                    // Stack above or below exercise bar (exercise bar is at 40px)
                    topPosition = eventIndex % 2 === 0 ? `${15 + eventIndex * 12}px` : `${70 + (eventIndex - 1) * 12}px`;
                  } else {
                    // Center position when no overlap, with proper spacing
                    topPosition = `${15 + eventIndex * 25}px`;
                  }

                  const eventTooltip = (props) => (
                    <Tooltip id={`event-tooltip-${event.id}`} {...props}>
                      <strong>{event.name}</strong><br />
                      {format(eventStartDate, 'MMM d, yyyy')} - {format(eventEndDate, 'MMM d, yyyy')}<br />
                      Type: {event.type} | Priority: {event.priority}
                    </Tooltip>
                  );

                  return (
                    eventWidth > 0 && eventLeft + eventWidth > 0 && eventLeft < 100 && (
                      <OverlayTrigger
                        key={event.id}
                        placement="top"
                        delay={{ show: 250, hide: 400 }}
                        overlay={eventTooltip}
                      >
                        <div
                          className="gantt-bar event-bar"
                          style={{
                            left: `${Math.max(0, eventLeft)}%`,
                            width: `${Math.max(Math.min(eventWidth, 100 - Math.max(0, eventLeft)), 0.5)}%`,
                            minWidth: `${Math.max(event.name.length * 7 + 16, 80)}px`,
                            top: topPosition,
                            height: '22px',
                            fontSize: '0.75em',
                            zIndex: 3,
                            cursor: 'pointer',
                            lineHeight: '22px'
                          }}
                          onClick={(e) => handleEventClick(e, event)}
                        >
                          {event.name}
                        </div>
                      </OverlayTrigger>
                    )
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>

      <EventModal
        show={showEventModal}
        handleClose={handleCloseEventModal}
        exerciseId={selectedExerciseId}
        onEventCreated={handleEventCreated}
        initialDate={clickedDate}
        editingEvent={editingEvent}
      />

      {/* Legend */}
      <div className={`gantt-legend ${legendCollapsed ? 'collapsed' : 'expanded'}`}>
        <div className="legend-header" onClick={() => setLegendCollapsed(!legendCollapsed)}>
          <h6>Legend</h6>
          <span className="legend-toggle">
            {legendCollapsed ? '▲' : '▼'}
          </span>
        </div>
        
        {!legendCollapsed && (
          <>
            <div className="legend-section">
          <div className="legend-title">Bars:</div>
          <div className="legend-items">
            <div className="legend-item">
              <div className="legend-exercise-bar"></div>
              <span className="legend-text">Exercise Duration</span>
            </div>
            <div className="legend-item">
              <div className="legend-event-bar"></div>
              <span className="legend-text">Event Duration</span>
            </div>
          </div>
        </div>

        <div className="legend-section">
          <div className="legend-title">Team Status (Ability to Support):</div>
          <div className="legend-items">
            <div className="legend-item">
              <span className="legend-status-circle red">🔴</span>
              <span className="legend-text">Unable to Support (0%)</span>
            </div>
            <div className="legend-item">
              <span className="legend-status-circle yellow">🟡</span>
              <span className="legend-text">Limited Support (50%)</span>
            </div>
            <div className="legend-item">
              <span className="legend-status-circle green">🟢</span>
              <span className="legend-text">Full Support (100%)</span>
            </div>
          </div>
        </div>

        <div className="legend-section">
          <div className="legend-title">Interactions:</div>
          <div className="legend-items">
            <div className="legend-item">
              <span className="legend-text">• Click timeline cells to create events</span>
            </div>
            <div className="legend-item">
              <span className="legend-text">• Click bars for details/editing</span>
            </div>
            <div className="legend-item">
              <span className="legend-text">• Hover bars for tooltips</span>
            </div>
            <div className="legend-item">
              <span className="legend-text">• Exercises ordered by priority level</span>
            </div>
          </div>
        </div>
          </>
        )}
      </div>
    </div>
  );
};

export default GanttChart;
