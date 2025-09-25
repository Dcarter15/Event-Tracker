import React, { useState, useEffect } from 'react';
import { Modal, Button, Accordion, Form } from 'react-bootstrap';

const statusColorMap = {
  green: 'success',
  yellow: 'warning',
  red: 'danger',
};

const getDivisionColor = (teams) => {
  if (!teams) return 'green';
  const statuses = teams.map(t => t.status);
  if (statuses.includes('red')) return 'red';
  if (statuses.includes('yellow')) return 'yellow';
  return 'green';
};

const ExerciseModal = ({ show, handleClose, exercise, onDivisionClick, onTeamClick }) => {
  // Calculate exercise readiness percentage
  const calculateReadiness = () => {
    if (!divisions || divisions.length === 0) return null;
    
    let totalTeams = 0;
    let totalScore = 0;
    
    divisions.forEach(division => {
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

  const getReadinessColor = (percentage) => {
    if (percentage >= 75) return 'success';
    if (percentage >= 50) return 'warning';
    return 'danger';
  };
  const [divisions, setDivisions] = useState([]);
  const [editingTeam, setEditingTeam] = useState(null);
  const [editValues, setEditValues] = useState({});
  const [editingDescription, setEditingDescription] = useState(false);
  const [descriptionValue, setDescriptionValue] = useState('');
  const [editingPriority, setEditingPriority] = useState(false);
  const [priorityValue, setPriorityValue] = useState('medium');
  const [editingDivisionLearning, setEditingDivisionLearning] = useState({});
  const [learningValues, setLearningValues] = useState({});
  const [showAddDivision, setShowAddDivision] = useState(false);
  const [newDivisionName, setNewDivisionName] = useState('');
  const [showAddTeam, setShowAddTeam] = useState({});
  const [newTeamName, setNewTeamName] = useState('');
  const [tasks, setTasks] = useState([]);
  const [showAddTask, setShowAddTask] = useState(false);
  const [newTask, setNewTask] = useState({ name: '', description: '', due_date: '' });
  const [editingTask, setEditingTask] = useState(null);
  const [editTaskValues, setEditTaskValues] = useState({});
  const [checkedTasks, setCheckedTasks] = useState({});
  const [showMultiAssign, setShowMultiAssign] = useState({});
  const [selectedTeams, setSelectedTeams] = useState({});

  useEffect(() => {
    if (exercise && show) {
      // Set initial description value
      setDescriptionValue(exercise.description || '');
      // Set initial priority value
      setPriorityValue(exercise.priority || 'medium');
      
      // Fetch tasks specific to this exercise
      fetch(`/api/tasks?exercise_id=${exercise.id}`)
        .then(response => response.json())
        .then(data => {
          setTasks(data || []);
        })
        .catch(error => {
          console.error('Error fetching tasks:', error);
          setTasks([]);
        });
      
      // Fetch divisions specific to this exercise
      fetch(`/api/divisions?exercise_id=${exercise.id}`)
        .then(response => response.json())
        .then(data => {
          setDivisions(data || []);
          // Initialize learning objectives values
          const initialLearning = {};
          (data || []).forEach(div => {
            initialLearning[div.id] = div.learning_objectives || '';
          });
          setLearningValues(initialLearning);
        })
        .catch(error => {
          console.error('Error fetching divisions:', error);
          setDivisions([]);
        });
    }
  }, [exercise, show]);

  const startEditing = (team) => {
    setEditingTeam(team.id);
    setEditValues({
      status: team.status,
      poc: team.poc,
      comments: team.comments,
      status_start: team.status_start ? team.status_start.split('T')[0] : '',
      status_end: team.status_end ? team.status_end.split('T')[0] : ''
    });
  };

  const cancelEditing = () => {
    setEditingTeam(null);
    setEditValues({});
  };

  const saveTeam = (divisionId, teamId) => {
    const division = divisions.find(d => d.id === divisionId);
    const team = division.teams.find(t => t.id === teamId);
    
    const updatedTeam = {
      id: team.id,
      exercise_id: team.exercise_id || exercise.id,
      name: team.name,
      division_id: team.division_id,
      poc: editValues.poc,
      status: editValues.status,
      status_start: editValues.status_start,
      status_end: editValues.status_end,
      comments: editValues.comments
    };

    fetch(`/api/team/update`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updatedTeam),
    })
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      // Update local state
      const updatedDivisions = divisions.map(div => {
        if (div.id === divisionId) {
          const updatedTeams = div.teams.map(t => {
            if (t.id === teamId) {
              return updatedTeam;
            }
            return t;
          });
          return { ...div, teams: updatedTeams };
        }
        return div;
      });
      setDivisions(updatedDivisions);
      setEditingTeam(null);
      setEditValues({});
      console.log('Team updated successfully');
    })
    .catch(error => {
      console.error('Error updating team:', error);
      alert('Failed to update team. Please try again.');
    });
  };

  const handleEditChange = (field, value) => {
    setEditValues(prev => ({ ...prev, [field]: value }));
  };

  const saveDescription = () => {
    const updatedExercise = {
      ...exercise,
      description: descriptionValue
    };

    fetch(`/api/exercises/${exercise.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updatedExercise),
    })
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      setEditingDescription(false);
      exercise.description = descriptionValue;
    })
    .catch(error => {
      console.error('Error updating description:', error);
      alert('Failed to update description. Please try again.');
    });
  };

  const savePriority = () => {
    const updatedExercise = {
      ...exercise,
      priority: priorityValue
    };

    fetch(`/api/exercises/${exercise.id}`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updatedExercise),
    })
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      setEditingPriority(false);
      exercise.priority = priorityValue;
    })
    .catch(error => {
      console.error('Error updating priority:', error);
      alert('Failed to update priority. Please try again.');
    });
  };

  const saveLearningObjectives = (divisionId) => {
    const division = divisions.find(d => d.id === divisionId);
    const updatedDivision = {
      ...division,
      learning_objectives: learningValues[divisionId]
    };

    fetch(`/api/divisions/update`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(updatedDivision),
    })
    .then(response => {
      if (!response.ok) {
        throw new Error('Network response was not ok');
      }
      setEditingDivisionLearning({ ...editingDivisionLearning, [divisionId]: false });
      // Update local state
      const updatedDivisions = divisions.map(div => {
        if (div.id === divisionId) {
          return { ...div, learning_objectives: learningValues[divisionId] };
        }
        return div;
      });
      setDivisions(updatedDivisions);
    })
    .catch(error => {
      console.error('Error updating learning objectives:', error);
      alert('Failed to update learning objectives. Please try again.');
    });
  };

  const addDivision = async () => {
    if (!newDivisionName.trim()) {
      alert('Please enter a division name');
      return;
    }

    const newDivision = {
      exercise_id: exercise.id,
      name: newDivisionName,
      learning_objectives: ''
    };

    try {
      const response = await fetch('/api/divisions', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newDivision),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const createdDivision = await response.json();
      setDivisions([...divisions, createdDivision]);
      setNewDivisionName('');
      setShowAddDivision(false);
    } catch (error) {
      console.error('Error adding division:', error);
      alert('Failed to add division. Please try again.');
    }
  };

  const addTask = async () => {
    if (!newTask.name.trim()) {
      alert('Please enter a task name');
      return;
    }

    const taskToCreate = {
      exercise_id: exercise.id,
      name: newTask.name,
      description: newTask.description,
      due_date: newTask.due_date || null,
      status: 'pending'
    };

    try {
      const response = await fetch('/api/tasks', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(taskToCreate),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const createdTask = await response.json();
      setTasks([...tasks, createdTask]);
      setNewTask({ name: '', description: '', due_date: '' });
      setShowAddTask(false);
    } catch (error) {
      console.error('Error adding task:', error);
      alert('Failed to add task. Please try again.');
    }
  };

  const updateTask = async (taskId) => {
    const task = tasks.find(t => t.id === taskId);
    const updatedTask = {
      ...task,
      ...editTaskValues[taskId]
    };

    try {
      const response = await fetch(`/api/tasks/${taskId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updatedTask),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const result = await response.json();
      setTasks(tasks.map(t => t.id === taskId ? result : t));
      setEditingTask(null);
      setEditTaskValues({});
    } catch (error) {
      console.error('Error updating task:', error);
      alert('Failed to update task. Please try again.');
    }
  };

  const deleteTask = async (taskId) => {
    if (!window.confirm('Are you sure you want to delete this task?')) {
      return;
    }

    try {
      const response = await fetch(`/api/tasks/${taskId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      setTasks(tasks.filter(t => t.id !== taskId));
    } catch (error) {
      console.error('Error deleting task:', error);
      alert('Failed to delete task. Please try again.');
    }
  };

  const startEditingTask = (task) => {
    setEditingTask(task.id);
    setEditTaskValues({
      [task.id]: {
        name: task.name,
        description: task.description || '',
        due_date: task.due_date ? task.due_date.split('T')[0] : '',
        assigned_to: task.assigned_to || '',
        status: task.status
      }
    });
  };

  const cancelEditingTask = () => {
    setEditingTask(null);
    setEditTaskValues({});
  };

  const handleTaskEditChange = (taskId, field, value) => {
    setEditTaskValues({
      ...editTaskValues,
      [taskId]: {
        ...editTaskValues[taskId],
        [field]: value
      }
    });
  };

  const assignTaskToTeam = async (taskId, teamId) => {
    try {
      // Get current task to see existing team assignments
      const currentTask = tasks.find(t => t.id === taskId);
      const existingTeamIds = currentTask.teams ? currentTask.teams.map(t => t.id) : (currentTask.team_id ? [currentTask.team_id] : []);

      // Add new team to existing teams (avoid duplicates)
      const updatedTeamIds = [...new Set([...existingTeamIds, teamId])];

      const response = await fetch(`/api/tasks/${taskId}/assign-multiple`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ team_ids: updatedTeamIds }),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      // Refresh tasks to get updated data
      const tasksResponse = await fetch(`/api/tasks?exercise_id=${exercise.id}`);
      if (tasksResponse.ok) {
        const updatedTasks = await tasksResponse.json();
        setTasks(updatedTasks || []);
      }
    } catch (error) {
      console.error('Error assigning task to team:', error);
      alert('Failed to assign task to team. Please try again.');
    }
  };

  const handleTaskCheck = (taskId) => {
    setCheckedTasks(prev => ({
      ...prev,
      [taskId]: !prev[taskId]
    }));
  };

  const markTaskComplete = async (taskId) => {
    const task = tasks.find(t => t.id === taskId);
    if (!task) return;

    try {
      const response = await fetch(`/api/tasks/${taskId}`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...task,
          status: 'completed'
        }),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      // Refresh tasks to get updated data
      const tasksResponse = await fetch(`/api/tasks?exercise_id=${exercise.id}`);
      if (tasksResponse.ok) {
        const updatedTasks = await tasksResponse.json();
        setTasks(updatedTasks || []);
        // Clear the checkbox for this task after marking complete
        setCheckedTasks(prev => {
          const newState = { ...prev };
          delete newState[taskId];
          return newState;
        });
      }
    } catch (error) {
      console.error('Error marking task as complete:', error);
      alert('Failed to mark task as complete. Please try again.');
    }
  };

  const unassignTaskFromTeam = async (taskId, specificTeamId = null) => {
    try {
      if (specificTeamId) {
        // Remove specific team while keeping others
        const currentTask = tasks.find(t => t.id === taskId);
        const existingTeamIds = currentTask.teams ? currentTask.teams.map(t => t.id) : (currentTask.team_id ? [currentTask.team_id] : []);
        const updatedTeamIds = existingTeamIds.filter(id => id !== specificTeamId);

        const response = await fetch(`/api/tasks/${taskId}/assign-multiple`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ team_ids: updatedTeamIds }),
        });

        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
      } else {
        // Remove all teams (original behavior)
        const response = await fetch(`/api/tasks/${taskId}/assign`, {
          method: 'PUT',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ team_id: null }),
        });

        if (!response.ok) {
          throw new Error('Network response was not ok');
        }
      }

      // Refresh tasks to get updated data
      const tasksResponse = await fetch(`/api/tasks?exercise_id=${exercise.id}`);
      if (tasksResponse.ok) {
        const updatedTasks = await tasksResponse.json();
        setTasks(updatedTasks || []);
      }
    } catch (error) {
      console.error('Error unassigning task from team:', error);
      alert('Failed to unassign task from team. Please try again.');
    }
  };

  const assignTaskToMultipleTeams = async (taskId) => {
    const newTeamIds = selectedTeams[taskId] || [];

    if (newTeamIds.length === 0) {
      alert('Please select at least one team to assign the task to.');
      return;
    }

    try {
      // Get current task to see existing team assignments
      const currentTask = tasks.find(t => t.id === taskId);
      const existingTeamIds = currentTask.teams ? currentTask.teams.map(t => t.id) : (currentTask.team_id ? [currentTask.team_id] : []);

      // Add new teams to existing teams (avoid duplicates)
      const allTeamIds = [...new Set([...existingTeamIds, ...newTeamIds])];

      const response = await fetch(`/api/tasks/${taskId}/assign-multiple`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ team_ids: allTeamIds }),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      // Refresh tasks to get updated data
      const tasksResponse = await fetch(`/api/tasks?exercise_id=${exercise.id}`);
      if (tasksResponse.ok) {
        const updatedTasks = await tasksResponse.json();
        setTasks(updatedTasks || []);
      }

      // Close multi-assign modal and clear selections
      setShowMultiAssign({ ...showMultiAssign, [taskId]: false });
      setSelectedTeams({ ...selectedTeams, [taskId]: [] });
    } catch (error) {
      console.error('Error assigning task to multiple teams:', error);
      alert('Failed to assign task to teams. Please try again.');
    }
  };

  const handleTeamSelection = (taskId, teamId, checked) => {
    setSelectedTeams(prev => {
      const currentSelection = prev[taskId] || [];
      if (checked) {
        return { ...prev, [taskId]: [...currentSelection, teamId] };
      } else {
        return { ...prev, [taskId]: currentSelection.filter(id => id !== teamId) };
      }
    });
  };

  const getAllTeams = () => {
    const allTeams = [];
    divisions.forEach(division => {
      division.teams?.forEach(team => {
        allTeams.push({
          ...team,
          divisionName: division.name
        });
      });
    });
    return allTeams;
  };

  const addTeam = async (divisionId) => {
    if (!newTeamName.trim()) {
      alert('Please enter a team name');
      return;
    }

    const newTeam = {
      exercise_id: exercise.id,
      division_id: divisionId,
      name: newTeamName,
      poc: '',
      status: 'green',
      comments: ''
    };

    try {
      const response = await fetch('/api/teams', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newTeam),
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      const createdTeam = await response.json();

      // Update divisions state to include the new team
      const updatedDivisions = divisions.map(div => {
        if (div.id === divisionId) {
          return {
            ...div,
            teams: [...(div.teams || []), createdTeam]
          };
        }
        return div;
      });

      setDivisions(updatedDivisions);
      setNewTeamName('');
      setShowAddTeam({ ...showAddTeam, [divisionId]: false });
    } catch (error) {
      console.error('Error adding team:', error);
      alert('Failed to add team. Please try again.');
    }
  };

  const deleteDivision = async (divisionId) => {
    if (!window.confirm('Are you sure you want to delete this division and all its teams?')) {
      return;
    }

    try {
      const response = await fetch(`/api/divisions/${divisionId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      // Remove division from local state
      setDivisions(divisions.filter(div => div.id !== divisionId));
    } catch (error) {
      console.error('Error deleting division:', error);
      alert('Failed to delete division. Please try again.');
    }
  };

  const deleteTeam = async (teamId) => {
    if (!window.confirm('Are you sure you want to delete this team?')) {
      return;
    }

    try {
      const response = await fetch(`/api/teams/${teamId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        throw new Error('Network response was not ok');
      }

      // Remove team from local state
      const updatedDivisions = divisions.map(div => ({
        ...div,
        teams: div.teams ? div.teams.filter(team => team.id !== teamId) : []
      }));
      setDivisions(updatedDivisions);
    } catch (error) {
      console.error('Error deleting team:', error);
      alert('Failed to delete team. Please try again.');
    }
  };

  if (!exercise) return null;

  return (
    <Modal show={show} onHide={handleClose} size="lg">
      <Modal.Header closeButton>
        <Modal.Title>
          {exercise.name}
          {(() => {
            const readiness = calculateReadiness();
            if (readiness !== null) {
              return (
                <span 
                  className={`badge bg-${getReadinessColor(readiness)} ms-3`}
                  style={{ fontSize: '0.6em', verticalAlign: 'middle' }}
                >
                  Ability to support: {readiness}%
                </span>
              );
            }
            return null;
          })()}
        </Modal.Title>
      </Modal.Header>
      <Modal.Body>
        <div className="mb-3">
          <div className="d-flex justify-content-between align-items-center mb-2">
            <h5>Exercise Description</h5>
            {!editingDescription ? (
              <Button variant="outline-primary" size="sm" onClick={() => setEditingDescription(true)}>
                Edit
              </Button>
            ) : (
              <div>
                <Button variant="success" size="sm" className="me-2" onClick={saveDescription}>
                  Save
                </Button>
                <Button variant="secondary" size="sm" onClick={() => {
                  setEditingDescription(false);
                  setDescriptionValue(exercise.description || '');
                }}>
                  Cancel
                </Button>
              </div>
            )}
          </div>
          {!editingDescription ? (
            <p>{exercise.description || 'No description provided'}</p>
          ) : (
            <Form.Control
              as="textarea"
              rows={3}
              value={descriptionValue}
              onChange={(e) => setDescriptionValue(e.target.value)}
              placeholder="Enter exercise description"
            />
          )}
        </div>

        <div className="mb-3">
          <div className="d-flex justify-content-between align-items-center mb-2">
            <h5>Exercise Priority</h5>
            {!editingPriority ? (
              <Button variant="outline-primary" size="sm" onClick={() => setEditingPriority(true)}>
                Edit
              </Button>
            ) : (
              <div>
                <Button variant="success" size="sm" className="me-2" onClick={savePriority}>
                  Save
                </Button>
                <Button variant="secondary" size="sm" onClick={() => {
                  setEditingPriority(false);
                  setPriorityValue(exercise.priority || 'medium');
                }}>
                  Cancel
                </Button>
              </div>
            )}
          </div>
          {!editingPriority ? (
            <p>
              <span className="badge bg-secondary me-2">
                {(exercise.priority || 'medium').charAt(0).toUpperCase() + (exercise.priority || 'medium').slice(1)} Priority
              </span>
            </p>
          ) : (
            <Form.Group>
              <Form.Label>Priority Level</Form.Label>
              <Form.Select
                value={priorityValue}
                onChange={(e) => setPriorityValue(e.target.value)}
              >
                <option value="low">Low</option>
                <option value="medium">Medium</option>
                <option value="high">High</option>
              </Form.Select>
            </Form.Group>
          )}
        </div>

        {/* Tasks Section */}
        <div className="mb-3">
          <div className="d-flex justify-content-between align-items-center mb-2">
            <h5>Tasks</h5>
            <Button variant="outline-success" size="sm" onClick={() => setShowAddTask(true)}>
              + Add Task
            </Button>
          </div>
          
          {/* Add Task Form */}
          {showAddTask && (
            <div className="mb-3 p-3 border rounded bg-light">
              <h6>Add New Task</h6>
              <div className="mb-2">
                <Form.Control
                  type="text"
                  placeholder="Task name"
                  value={newTask.name}
                  onChange={(e) => setNewTask({ ...newTask, name: e.target.value })}
                />
              </div>
              <div className="mb-2">
                <Form.Control
                  as="textarea"
                  rows={2}
                  placeholder="Description (optional)"
                  value={newTask.description}
                  onChange={(e) => setNewTask({ ...newTask, description: e.target.value })}
                />
              </div>
              <div className="row mb-2">
                <div className="col">
                  <Form.Control
                    type="date"
                    placeholder="Due date"
                    value={newTask.due_date}
                    onChange={(e) => setNewTask({ ...newTask, due_date: e.target.value })}
                  />
                </div>
              </div>
              <div className="d-flex gap-2">
                <Button variant="success" size="sm" onClick={addTask}>
                  Add Task
                </Button>
                <Button variant="secondary" size="sm" onClick={() => {
                  setShowAddTask(false);
                  setNewTask({ name: '', description: '', due_date: '' });
                }}>
                  Cancel
                </Button>
              </div>
            </div>
          )}
          
          {/* All Tasks List */}
          {tasks.length === 0 ? (
            <p className="text-muted">No tasks yet. Click "Add Task" to create one.</p>
          ) : (
            <div className="list-group">
              {tasks.map(task => {
                  const isEditing = editingTask === task.id;
                  const taskValues = editTaskValues[task.id] || task;
                  
                  return (
                    <div key={task.id} className="list-group-item">
                      {!isEditing ? (
                        <div>
                          <div className="d-flex justify-content-between align-items-start">
                            <div className="flex-grow-1">
                              <h6 className="mb-1">
                                {task.name}
                                {task.status === 'in-progress' && (
                                  <span className="badge bg-warning ms-2">In Progress</span>
                                )}
                                {(!task.teams || task.teams.length === 0) && !task.team_id ? (
                                  <>
                                    {task.status === 'pending' && (
                                      <span className="badge bg-secondary ms-2">Pending</span>
                                    )}
                                    <span className="badge bg-info ms-2">Unassigned</span>
                                  </>
                                ) : (
                                  <>
                                    {task.teams && task.teams.length > 0 ? (
                                      task.teams.map((team, idx) => (
                                        <span key={team.id} className="badge bg-primary ms-2">
                                          {team.name} ({team.divisionName || divisions.find(d => d.teams?.some(t => t.id === team.id))?.name || 'Unknown'})
                                        </span>
                                      ))
                                    ) : task.team_id ? (
                                      <span className="badge bg-primary ms-2">
                                        Assigned to {task.team_name} ({task.division_name})
                                      </span>
                                    ) : null}
                                  </>
                                )}
                              </h6>
                              {task.description && (
                                <p className="mb-1 text-muted small">{task.description}</p>
                              )}
                              <div className="small text-muted">
                                {task.due_date && (
                                  <span className="me-3">
                                    <strong>Due:</strong> {new Date(task.due_date).toLocaleDateString()}
                                  </span>
                                )}
                              </div>
                            </div>
                            <div className="d-flex gap-1">
                              <Button variant="outline-primary" size="sm" onClick={() => startEditingTask(task)}>
                                Edit
                              </Button>
                              <Button variant="outline-danger" size="sm" onClick={() => deleteTask(task.id)}>
                                Delete
                              </Button>
                            </div>
                          </div>
                        </div>
                      ) : (
                        <div>
                          <div className="mb-2">
                            <Form.Control
                              type="text"
                              value={taskValues.name}
                              onChange={(e) => handleTaskEditChange(task.id, 'name', e.target.value)}
                            />
                          </div>
                          <div className="mb-2">
                            <Form.Control
                              as="textarea"
                              rows={2}
                              value={taskValues.description}
                              onChange={(e) => handleTaskEditChange(task.id, 'description', e.target.value)}
                              placeholder="Description (optional)"
                            />
                          </div>
                          <div className="row mb-2">
                            <div className="col">
                              <Form.Select
                                value={taskValues.status}
                                onChange={(e) => handleTaskEditChange(task.id, 'status', e.target.value)}
                              >
                                <option value="pending">Pending</option>
                                <option value="in-progress">In Progress</option>
                                <option value="completed">Completed</option>
                              </Form.Select>
                            </div>
                            <div className="col">
                              <Form.Control
                                type="date"
                                value={taskValues.due_date}
                                onChange={(e) => handleTaskEditChange(task.id, 'due_date', e.target.value)}
                              />
                            </div>
                          </div>
                          <div className="d-flex gap-2">
                            <Button variant="success" size="sm" onClick={() => updateTask(task.id)}>
                              Save
                            </Button>
                            <Button variant="secondary" size="sm" onClick={cancelEditingTask}>
                              Cancel
                            </Button>
                          </div>
                        </div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
        </div>
        
        {(exercise.srd_poc || exercise.cpd_poc) && (
          <div className="mb-3">
            <h5>Points of Contact</h5>
            {exercise.srd_poc && <p><strong>SRD POC:</strong> {exercise.srd_poc}</p>}
            {exercise.cpd_poc && <p><strong>CPD POC:</strong> {exercise.cpd_poc}</p>}
          </div>
        )}
        
        {exercise.aoc_involvement && (
          <div className="mb-3">
            <h5>AOC Involvement</h5>
            <p>{exercise.aoc_involvement}</p>
          </div>
        )}
        
        {exercise.tasked_divisions && exercise.tasked_divisions.length > 0 && (
          <div className="mb-3">
            <h5>Tasked Divisions</h5>
            <ul>
              {exercise.tasked_divisions.map((div, index) => (
                <li key={index}>{div}</li>
              ))}
            </ul>
          </div>
        )}
        
        <div className="d-flex justify-content-between align-items-center mb-3">
          <h5>Participating Divisions</h5>
          <Button 
            variant="success" 
            size="sm"
            onClick={() => setShowAddDivision(true)}
          >
            + Add Division
          </Button>
        </div>
        
        {/* Add Division Form */}
        {showAddDivision && (
          <div className="mb-3 p-3 border rounded bg-light">
            <h6>Add New Division</h6>
            <div className="d-flex gap-2">
              <Form.Control
                type="text"
                placeholder="Enter division name"
                value={newDivisionName}
                onChange={(e) => setNewDivisionName(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && addDivision()}
              />
              <Button variant="success" onClick={addDivision}>
                Add
              </Button>
              <Button variant="secondary" onClick={() => {
                setShowAddDivision(false);
                setNewDivisionName('');
              }}>
                Cancel
              </Button>
            </div>
          </div>
        )}
        
        <Accordion alwaysOpen>
          {divisions.map((division, index) => {
            const divisionColor = getDivisionColor(division.teams);
            return (
              <Accordion.Item eventKey={String(index)} key={division.id}>
                <Accordion.Header>
                  <div className="d-flex justify-content-between align-items-center w-100">
                    <div>
                      <span className={`me-2 text-${statusColorMap[divisionColor]}`}>●</span>
                      <span>{division.name}</span>
                    </div>
                    <Button
                      variant="outline-danger"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation();
                        deleteDivision(division.id);
                      }}
                      title="Delete division and all its teams"
                    >
                      Delete
                    </Button>
                  </div>
                </Accordion.Header>
                <Accordion.Body>
                  {/* Learning Objectives Section */}
                  <div className="mb-4 p-3 bg-light rounded">
                    <div className="d-flex justify-content-between align-items-center mb-2">
                      <h6>Division Learning Objectives</h6>
                      {!editingDivisionLearning[division.id] ? (
                        <Button 
                          variant="outline-primary" 
                          size="sm"
                          onClick={() => setEditingDivisionLearning({ ...editingDivisionLearning, [division.id]: true })}
                        >
                          Edit
                        </Button>
                      ) : (
                        <div>
                          <Button 
                            variant="success" 
                            size="sm"
                            className="me-2"
                            onClick={() => saveLearningObjectives(division.id)}
                          >
                            Save
                          </Button>
                          <Button 
                            variant="secondary" 
                            size="sm"
                            onClick={() => {
                              setEditingDivisionLearning({ ...editingDivisionLearning, [division.id]: false });
                              setLearningValues({ ...learningValues, [division.id]: division.learning_objectives || '' });
                            }}
                          >
                            Cancel
                          </Button>
                        </div>
                      )}
                    </div>
                    {!editingDivisionLearning[division.id] ? (
                      <p className="mb-0">{division.learning_objectives || 'No learning objectives defined'}</p>
                    ) : (
                      <Form.Control
                        as="textarea"
                        rows={3}
                        value={learningValues[division.id] || ''}
                        onChange={(e) => setLearningValues({ ...learningValues, [division.id]: e.target.value })}
                        placeholder="Enter learning objectives for this division"
                      />
                    )}
                  </div>
                  {/* Teams Section */}
                  <div className="d-flex justify-content-between align-items-center mb-3">
                    <h6>Teams</h6>
                    <Button 
                      variant="outline-success" 
                      size="sm"
                      onClick={() => setShowAddTeam({ ...showAddTeam, [division.id]: true })}
                    >
                      + Add Team
                    </Button>
                  </div>
                  
                  {/* Add Team Form */}
                  {showAddTeam[division.id] && (
                    <div className="mb-3 p-2 border rounded">
                      <div className="d-flex gap-2">
                        <Form.Control
                          type="text"
                          placeholder="Enter team name"
                          value={newTeamName}
                          onChange={(e) => setNewTeamName(e.target.value)}
                          onKeyPress={(e) => e.key === 'Enter' && addTeam(division.id)}
                        />
                        <Button variant="success" size="sm" onClick={() => addTeam(division.id)}>
                          Add
                        </Button>
                        <Button variant="secondary" size="sm" onClick={() => {
                          setShowAddTeam({ ...showAddTeam, [division.id]: false });
                          setNewTeamName('');
                        }}>
                          Cancel
                        </Button>
                      </div>
                    </div>
                  )}
                  
                  {division.teams?.map(team => {
                    const isEditing = editingTeam === team.id;
                    return (
                      <div key={team.id} className="mb-4 p-3 border rounded">
                        <div className="d-flex justify-content-between align-items-center mb-3">
                          <h5>
                            <span
                              className="text-primary"
                              style={{ cursor: 'pointer', textDecoration: 'underline' }}
                              onClick={(e) => {
                                e.stopPropagation();
                                if (onTeamClick) {
                                  onTeamClick(team, division.name);
                                  handleClose();
                                }
                              }}
                              title="Click to see exercises for this team"
                            >
                              {team.name}
                            </span>
                          </h5>
                          {!isEditing ? (
                            <div className="d-flex gap-2">
                              <Button
                                variant="outline-primary"
                                size="sm"
                                onClick={() => startEditing(team)}
                              >
                                Edit
                              </Button>
                              <Button
                                variant="outline-danger"
                                size="sm"
                                onClick={() => deleteTeam(team.id)}
                                title="Delete team"
                              >
                                Delete
                              </Button>
                            </div>
                          ) : (
                            <div>
                              <Button
                                variant="success"
                                size="sm"
                                className="me-2"
                                onClick={() => saveTeam(division.id, team.id)}
                              >
                                Save
                              </Button>
                              <Button
                                variant="secondary"
                                size="sm"
                                onClick={cancelEditing}
                              >
                                Cancel
                              </Button>
                            </div>
                          )}
                        </div>

                        {!isEditing ? (
                          // Read-only view
                          <div>
                            <div className="mb-2">
                              <strong>Status:</strong>{' '}
                              <span className={`badge bg-${statusColorMap[team.status]}`}>
                                {team.status.toUpperCase()}
                              </span>
                            </div>
                            {(team.status_start || team.status_end) && (
                              <div className="mb-2">
                                <strong>Status Duration:</strong>{' '}
                                {team.status_start ? new Date(team.status_start).toLocaleDateString() : 'N/A'} to{' '}
                                {team.status_end ? new Date(team.status_end).toLocaleDateString() : 'N/A'}
                              </div>
                            )}
                            <div className="mb-2">
                              <strong>Point of Contact:</strong> {team.poc || 'Not assigned'}
                            </div>
                            <div className="mb-2">
                              <strong>Comments:</strong> {team.comments || 'No comments'}
                            </div>
                          </div>
                        ) : (
                          // Edit view
                          <Form>
                            <Form.Group className="mb-3">
                              <Form.Label>Status</Form.Label>
                              <Form.Select
                                value={editValues.status}
                                onChange={(e) => handleEditChange('status', e.target.value)}
                              >
                                <option value="green">Green</option>
                                <option value="yellow">Yellow</option>
                                <option value="red">Red</option>
                              </Form.Select>
                            </Form.Group>
                            
                            <Form.Group className="mb-3">
                              <Form.Label>Status Duration</Form.Label>
                              <div className="d-flex gap-2">
                                <Form.Control
                                  type="date"
                                  value={editValues.status_start}
                                  onChange={(e) => handleEditChange('status_start', e.target.value)}
                                />
                                <span className="align-self-center">to</span>
                                <Form.Control
                                  type="date"
                                  value={editValues.status_end}
                                  onChange={(e) => handleEditChange('status_end', e.target.value)}
                                />
                              </div>
                            </Form.Group>
                            
                            <Form.Group className="mb-3">
                              <Form.Label>Point of Contact (POC)</Form.Label>
                              <Form.Control
                                type="text"
                                value={editValues.poc}
                                onChange={(e) => handleEditChange('poc', e.target.value)}
                                placeholder="Enter POC name"
                              />
                            </Form.Group>
                            
                            <Form.Group className="mb-3">
                              <Form.Label>Comments</Form.Label>
                              <Form.Control
                                as="textarea"
                                rows={3}
                                value={editValues.comments}
                                onChange={(e) => handleEditChange('comments', e.target.value)}
                                placeholder="Enter any comments or notes"
                              />
                            </Form.Group>
                          </Form>
                        )}

                        {/* Task Assignment Section */}
                        <div className="mt-3 pt-3 border-top">
                          <h6 className="mb-3">Task Assignment</h6>
                          
                          {/* Assigned Tasks */}
                          {(() => {
                            const teamTasks = tasks.filter(task =>
                              task.team_id === team.id ||
                              (task.teams && task.teams.some(t => t.id === team.id))
                            );
                            return teamTasks.length > 0 && (
                              <div className="mb-3">
                                <div className="small text-muted mb-2"><strong>Assigned Tasks:</strong></div>
                                {teamTasks.map(task => (
                                  <div key={task.id} className="d-flex justify-content-between align-items-center mb-2 p-2 bg-light rounded">
                                    <div className="d-flex align-items-start">
                                      <Form.Check
                                        type="checkbox"
                                        className="me-2 mt-1"
                                        checked={checkedTasks[task.id] || false}
                                        onChange={() => handleTaskCheck(task.id)}
                                        title="Check if team can handle this task"
                                      />
                                      <div>
                                        <span className="fw-bold">{task.name}</span>
                                        {task.status === 'in-progress' && (
                                          <span className="badge bg-warning ms-2">In Progress</span>
                                        )}
                                        {task.status === 'pending' && (
                                          <span className="badge bg-primary ms-2">Assigned</span>
                                        )}
                                        {checkedTasks[task.id] && (
                                          <span className="badge bg-info ms-2">Can Handle</span>
                                        )}
                                        {task.due_date && (
                                          <div className="small text-muted">
                                            Due: {new Date(task.due_date).toLocaleDateString()}
                                          </div>
                                        )}
                                      </div>
                                    </div>
                                    <div className="d-flex gap-2">
                                      <Button
                                        variant="outline-warning"
                                        size="sm"
                                        onClick={() => unassignTaskFromTeam(task.id, team.id)}
                                      >
                                        Unassign from {team.name}
                                      </Button>
                                    </div>
                                  </div>
                                ))}
                              </div>
                            );
                          })()}

                          {/* Available Tasks to Assign */}
                          {(() => {
                            // Show all tasks except those already assigned to this specific team
                            const availableTasks = tasks.filter(task => {
                              const isAssignedToThisTeam = task.team_id === team.id ||
                                (task.teams && task.teams.some(t => t.id === team.id));
                              return !isAssignedToThisTeam;
                            });

                            return availableTasks.length > 0 && (
                              <div>
                                <div className="small text-muted mb-2"><strong>Available Tasks:</strong></div>
                                {availableTasks.map(task => (
                                  <div key={task.id} className="d-flex justify-content-between align-items-center mb-2 p-2 border rounded">
                                    <div>
                                      <span className="fw-bold">{task.name}</span>
                                      {task.status === 'in-progress' && (
                                        <span className="badge bg-warning ms-2">In Progress</span>
                                      )}
                                      {task.status === 'pending' && (
                                        <span className="badge bg-secondary ms-2">Pending</span>
                                      )}
                                      {/* Show which other teams this task is assigned to */}
                                      {task.teams && task.teams.length > 0 && (
                                        <div className="small text-muted">
                                          Also assigned to: {task.teams.map(t => t.name).join(', ')}
                                        </div>
                                      )}
                                      {task.team_id && task.team_name && task.team_id !== team.id && (
                                        <div className="small text-muted">
                                          Also assigned to: {task.team_name}
                                        </div>
                                      )}
                                      {task.due_date && (
                                        <div className="small text-muted">
                                          Due: {new Date(task.due_date).toLocaleDateString()}
                                        </div>
                                      )}
                                    </div>
                                    <Button
                                      variant="outline-primary"
                                      size="sm"
                                      onClick={() => assignTaskToTeam(task.id, team.id)}
                                    >
                                      Assign
                                    </Button>
                                  </div>
                                ))}
                              </div>
                            );
                          })()}

                          {tasks.filter(task => {
                            const isAssignedToThisTeam = task.team_id === team.id ||
                              (task.teams && task.teams.some(t => t.id === team.id));
                            return !isAssignedToThisTeam;
                          }).length === 0 && (
                            <p className="small text-muted">All tasks are already assigned to this team.</p>
                          )}
                        </div>
                      </div>
                    );
                  })}
                </Accordion.Body>
              </Accordion.Item>
            );
          })}
        </Accordion>
      </Modal.Body>
      <Modal.Footer>
        <Button variant="secondary" onClick={handleClose}>
          Close
        </Button>
      </Modal.Footer>
    </Modal>
  );
};

export default ExerciseModal;