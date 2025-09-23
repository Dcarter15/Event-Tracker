import React, { useState, useEffect } from 'react';
import './App.css';
import ExerciseModal from './components/ExerciseModal';
import GanttChart from './components/GanttChart'; // Import the new component
import Chatbot from './components/Chatbot';

function App() {
  const [exercises, setExercises] = useState([]);
  const [selectedExercise, setSelectedExercise] = useState(null);
  const [showModal, setShowModal] = useState(false);
  const [filteredView, setFilteredView] = useState(null); // { type: 'division', id: 123, name: 'Division Name' } or { type: 'team', id: 456, name: 'Team Name' }

  useEffect(() => {
    fetchExercises();
  }, [filteredView]);

  const handleExerciseClick = (exercise) => {
    setSelectedExercise(exercise);
    setShowModal(true);
  };

  const handleCloseModal = () => {
    setShowModal(false);
    setSelectedExercise(null);
    // Refresh exercises data to pick up any changes made in the modal
    fetchExercises();
  };

  const fetchExercises = () => {
    let url = '/api/exercises';
    if (filteredView) {
      if (filteredView.type === 'division') {
        url += `?division_name=${encodeURIComponent(filteredView.name)}`;
      } else if (filteredView.type === 'team') {
        url += `?team_name=${encodeURIComponent(filteredView.name)}`;
      }
    }

    fetch(url)
      .then(response => response.json())
      .then(data => {
        if (Array.isArray(data)) {
          setExercises(data);
        } else {
          console.error('Error: Exercises data is not an array', data);
          setExercises([]);
        }
      })
      .catch(error => {
        console.error('Error fetching exercises:', error);
        setExercises([]);
      });
  };

  const handleDivisionClick = (division) => {
    setFilteredView({ type: 'division', name: division.name });
  };

  const handleTeamClick = (team, divisionName) => {
    setFilteredView({ type: 'team', name: team.name, displayName: `${team.name} (${divisionName})` });
  };

  const clearFilter = () => {
    setFilteredView(null);
  };

  return (
    <div className="App">
      <header className="App-header">
        <h1>AOC Event Tracker</h1>
      </header>
      <main className="container-fluid mt-4"> {/* Use container-fluid for more width */}
        <div className="mb-3">
          <h2>Event Timeline</h2>
        </div>
        {exercises.length > 0 ? (
          <GanttChart
            exercises={exercises}
            onExerciseClick={handleExerciseClick}
            onDivisionFilter={handleDivisionClick}
            filteredView={filteredView}
            onClearFilter={clearFilter}
          />
        ) : (
          <p>Loading exercises or no exercises to display.</p>
        )}
      </main>

      <ExerciseModal
        show={showModal}
        handleClose={handleCloseModal}
        exercise={selectedExercise}
        onDivisionClick={handleDivisionClick}
        onTeamClick={handleTeamClick}
      />
      <Chatbot />
    </div>
  );
}

export default App;