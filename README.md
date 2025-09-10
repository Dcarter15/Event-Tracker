# SRD Calendar Project

A web-based exercise tracking and management system for AOC (Army Operations Center) with a Gantt chart visualization, team status tracking, and AI-powered chatbot interface.

## Features

- 📅 **Interactive Gantt Chart**: Visual timeline of exercises with day/week/month views
- 👥 **Division & Team Management**: Track multiple divisions and teams per exercise
- 🚦 **Status Tracking**: Red/Yellow/Green status indicators with date ranges
- 💬 **AI Chatbot**: Natural language interface for managing exercises
- 🗄️ **PostgreSQL Database**: Persistent data storage
- 🔄 **Real-time Updates**: Changes immediately reflected across the application

## Tech Stack

### Backend
- Go (Golang)
- Chi router for HTTP routing
- PostgreSQL database
- RESTful API design

### Frontend
- React.js
- Bootstrap for UI components
- date-fns for date manipulation
- React Bootstrap for modal components

## Prerequisites

- Go 1.19 or higher
- Node.js 16 or higher
- PostgreSQL 12 or higher
- Git

## Installation

1. **Clone the repository**
```bash
git clone https://github.com/yourusername/srd-calendar-project.git
cd srd-calendar-project
```

2. **Set up the database**
- Ensure PostgreSQL is running
- Create a database or use existing one
- Update credentials in `backend/.env`:
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=test_db
```

3. **Install backend dependencies**
```bash
cd backend
go mod download
```

4. **Install frontend dependencies**
```bash
cd ../frontend
npm install
```

## Running the Application

1. **Start the backend server**
```bash
cd backend
go run ./cmd/api/main.go
```
The backend will run on http://localhost:8081

2. **Start the frontend development server**
```bash
cd frontend
npm start
```
The frontend will run on http://localhost:3000

## Usage

### Main Calendar View
- View exercises on a Gantt chart timeline
- Switch between Month, Week, and Day views
- Click on any exercise bar to see details

### Exercise Management
- Click on an exercise to view/edit division and team details
- Each team has:
  - Status indicator (Green/Yellow/Red)
  - Point of Contact (POC)
  - Status duration dates
  - Comments section
- Click "Edit" to modify team information

### Chatbot Commands
Click the "?" button in the bottom right to open the chatbot. Available commands:

- `help` - Show available commands
- `list exercises` - Display all exercises
- `add exercise [name] from [date] to [date]` - Create new exercise
- `update exercise [ID] [field] to [value]` - Modify exercise
- `delete exercise [ID]` - Remove exercise
- `show exercises this week/month` - Time-based queries
- `show upcoming exercises` - View future exercises

## Project Structure

```
srd-calendar-project/
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go          # Application entry point
│   ├── internal/
│   │   ├── database/            # Database connection and setup
│   │   ├── handlers/            # HTTP request handlers
│   │   ├── models/              # Data models
│   │   └── repository/          # Data access layer
│   ├── go.mod
│   └── go.sum
├── frontend/
│   ├── public/
│   ├── src/
│   │   ├── components/          # React components
│   │   │   ├── GanttChart.js    # Timeline visualization
│   │   │   ├── ExerciseModal.js # Exercise details modal
│   │   │   └── Chatbot.js       # AI assistant interface
│   │   ├── App.js               # Main application component
│   │   └── index.js             # Application entry point
│   └── package.json
└── README.md
```

## Database Schema

- **exercises**: Main exercise information
- **divisions**: Divisions linked to exercises
- **teams**: Teams within divisions
- **tasked_divisions**: Many-to-many relationship for assigned divisions

## Environment Variables

Backend environment variables (in `backend/.env`):
- `DB_HOST` - PostgreSQL host (default: localhost)
- `DB_PORT` - PostgreSQL port (default: 5432)
- `DB_USER` - Database username (default: postgres)
- `DB_PASSWORD` - Database password
- `DB_NAME` - Database name

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License.

## Contact

Project Link: [https://github.com/yourusername/srd-calendar-project](https://github.com/yourusername/srd-calendar-project)