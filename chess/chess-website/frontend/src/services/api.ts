import axios from 'axios';
import { type BoardState, type Move, type Position } from '../types';

const API_BASE = 'http://localhost:8001';

const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
});

export const gameApi = {
  startGame: async (): Promise<BoardState> => {
    const response = await api.post('/game/start');
    return response.data;
  },

  getBoard: async (): Promise<BoardState> => {
    const response = await api.get('/game/board');
    return response.data;
  },

  makeMove: async (move: Move): Promise<BoardState> => {
    const moveData = {
      from_pos: move.from,
      to_pos: move.to,
      player: move.player,
    };
    const response = await api.post('/game/move', moveData);
    return response.data;
  },

  getAiMove: async (): Promise<{ move: Move }> => {
    const response = await api.get('/game/ai-move');
    return response.data;
  },

  resetGame: async (): Promise<BoardState> => {
    const response = await api.post('/game/reset');
    return response.data;
  },

  getValidMoves: async (position: Position): Promise<{ moves: Position[] }> => {
    const response = await api.post('/game/valid-moves', position);
    return response.data;
  },
};
