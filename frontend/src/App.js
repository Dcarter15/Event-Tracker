import React, { useState, useEffect } from 'react';
import './App.css';
import ExerciseModal from './components/ExerciseModal';
import GanttChart from './components/GanttChart'; // Import the new component
import Chatbot from './components/Chatbot';

function App() {
  const [exercises, setExercises] = useState([]);
  const [selectedExercise, setSelectedExercise] = useState(null);
  const [showModal, setShowModal] = useState(false);

  useEffect(() => {
    fetchExercises();
  }, []);

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
    fetch('/api/exercises')
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

  return (
    <div className="App">
      <header className="App-header">
        <h1>AOC Exercise Tracker</h1>
      </header>
      <main className="container-fluid mt-4"> {/* Use container-fluid for more width */}
        <h2 className="text-center">Exercise Timeline</h2>
        {exercises.length > 0 ? (
          <GanttChart exercises={exercises} onExerciseClick={handleExerciseClick} />
        ) : (
          <p>Loading exercises or no exercises to display.</p>
        )}
      </main>

      <ExerciseModal
        show={showModal}
        handleClose={handleCloseModal}
        exercise={selectedExercise}
      />
      <Chatbot />
    </div>
  );
}

export default App;