# Chinese Chess (中国象棋)

A Chinese Chess game with human vs AI gameplay.

## Project Structure

```
chess-website/
├── backend/           # FastAPI backend
│   ├── app/
│   │   ├── api/      # API routes
│   │   └── engine/   # Game logic & AI
│   ├── requirements.txt
│   └── v Python virtual environment
env/         #└── frontend/         # React + TypeScript frontend
    └── src/
        ├── components/
        ├── hooks/
        ├── services/
        └── types/
```

## Running the Application

### Backend

```bash
cd chess-website/backend

# Activate virtual environment
source venv/bin/activate

# Run the server
uvicorn app.main:app --reload --host 0.0.0.0 --port 8001
```

The backend API will be available at http://localhost:8001

### Frontend

```bash
cd chess-website/frontend

# Install dependencies (first time only)
npm install

# Run the development server
npm run dev
```

The frontend will be available at http://localhost:5173

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/game/board` | Get current board state |
| POST | `/game/start` | Start a new game |
| POST | `/game/move` | Make a move |
| GET | `/game/ai-move` | Get AI's move |
| POST | `/game/reset` | Reset the game |
| POST | `/game/valid-moves` | Get valid moves for a position |
