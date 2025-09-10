import React, { useState } from 'react';
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
              <div className="gantt-row-header" onClick={() => onExerciseClick(exercise)}>
                {exercise.name}
              </div>
              <div className="gantt-bars">
                {left >= 0 && width > 0 && (
                  <OverlayTrigger
                    placement="top"
                    delay={{ show: 250, hide: 400 }}
                    overlay={renderTooltip}
                  >
                    <div
                      className="gantt-bar"
                      style={{
                        left: `${left}%`,
                        width: `${width}%`,
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
