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

const ExerciseModal = ({ show, handleClose, exercise }) => {
  const [divisions, setDivisions] = useState([]);
  const [editingTeam, setEditingTeam] = useState(null);
  const [editValues, setEditValues] = useState({});
  const [editingDescription, setEditingDescription] = useState(false);
  const [descriptionValue, setDescriptionValue] = useState('');
  const [editingDivisionLearning, setEditingDivisionLearning] = useState({});
  const [learningValues, setLearningValues] = useState({});
  const [showAddDivision, setShowAddDivision] = useState(false);
  const [newDivisionName, setNewDivisionName] = useState('');
  const [showAddTeam, setShowAddTeam] = useState({});
  const [newTeamName, setNewTeamName] = useState('');

  useEffect(() => {
    if (exercise && show) {
      // Set initial description value
      setDescriptionValue(exercise.description || '');
      
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

  if (!exercise) return null;

  return (
    <Modal show={show} onHide={handleClose} size="lg">
      <Modal.Header closeButton>
        <Modal.Title>{exercise.name}</Modal.Title>
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
                  <span className={`me-2 text-${statusColorMap[divisionColor]}`}>●</span>
                  {division.name}
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
                          <h5>{team.name}</h5>
                          {!isEditing ? (
                            <Button 
                              variant="outline-primary" 
                              size="sm"
                              onClick={() => startEditing(team)}
                            >
                              Edit
                            </Button>
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