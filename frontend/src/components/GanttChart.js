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

const GanttChart = ({ exercises, onExerciseClick }) => {
  const [viewMode, setViewMode] = useState('month'); // 'month', 'week', 'day'
  const [editingPOC, setEditingPOC] = useState({});
  const [pocValues, setPocValues] = useState({});

  // Initialize POC values from exercises
  useEffect(() => {
    const initialPocValues = {};
    exercises.forEach(exercise => {
      initialPocValues[exercise.id] = exercise.exercise_event_poc || '';
    });
    setPocValues(initialPocValues);
  }, [exercises]);

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
        return format(date, 'MMM yyyy');
    }
  };
  
  const getMonthLabel = (date) => {
    if (viewMode === 'day' && format(date, 'd') === '1') {
        return format(date, 'MMMM yyyy');
    }
    return null;
  }

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

  return (
    <div>
      <div className="gantt-controls mb-3">
        <ButtonGroup>
          <Button variant={viewMode === 'month' ? 'primary' : 'secondary'} onClick={() => setViewMode('month')}>Month</Button>
          <Button variant={viewMode === 'week' ? 'primary' : 'secondary'} onClick={() => setViewMode('week')}>Week</Button>
          <Button variant={viewMode === 'day' ? 'primary' : 'secondary'} onClick={() => setViewMode('day')}>Day</Button>
        </ButtonGroup>
      </div>

      <div className="gantt-chart">
        {/* Header */}
        <div className="gantt-header">
          <div className="gantt-poc-header">Exercise Event POC</div>
          <div className="gantt-row-header">Exercise</div>
          <div className="gantt-timeline" style={{ gridTemplateColumns: `repeat(${timelineHeaders.length}, 1fr)` }}>
            {timelineHeaders.map(headerDate => (
              <div key={headerDate.toString()} className={`gantt-timeline-header gantt-${viewMode}`}>
                {getMonthLabel(headerDate) && <div className="gantt-month-label">{getMonthLabel(headerDate)}</div>}
                {getHeaderLabel(headerDate)}
              </div>
            ))}
          </div>
        </div>

        {/* Rows */}
        {exercises.map(exercise => {
          const exerciseStartDate = new Date(exercise.start_date);
          const exerciseEndDate = new Date(exercise.end_date);

          const offset = differenceInDays(exerciseStartDate, chartStartDate);
          const duration = Math.max(1, differenceInDays(exerciseEndDate, exerciseStartDate));

          const left = (offset / totalDays) * 100;
          const width = (duration / totalDays) * 100;

          const renderTooltip = (props) => (
            <Tooltip id={`tooltip-${exercise.id}`} {...props}>
              <strong>{exercise.name}</strong><br />
              {format(exerciseStartDate, 'MMM d, yyyy')} - {format(exerciseEndDate, 'MMM d, yyyy')}
            </Tooltip>
          );

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
              <div className="gantt-bars">
                {width > 0 && left + width > 0 && left < 100 && (
                  <OverlayTrigger
                    placement="top"
                    delay={{ show: 250, hide: 400 }}
                    overlay={renderTooltip}
                  >
                    <div
                      className="gantt-bar"
                      style={{
                        left: `${Math.max(0, left)}%`,
                        width: `${Math.min(width, 100 - Math.max(0, left))}%`,
                      }}
                    >
                      <span className="gantt-bar-label">{exercise.name}</span>
                    </div>
                  </OverlayTrigger>
                )}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};

export default GanttChart;
