import { useState, useCallback } from 'react';
import { 
  type BoardState, 
  type GameState, 
  type Position, 
  type Move, 
  INITIAL_BOARD 
} from '../types';
import { gameApi } from '../services/api';
import { playMoveSound, playCaptureSound, playCheckSound, playWinSound, playLoseSound } from '../services/sounds';

const createInitialBoard = (): BoardState => ({
  pieces: INITIAL_BOARD.map(row => [...row]),
  currentPlayer: 'red',
  isCheck: false,
  gameOver: false,
  winner: null,
});

export function useGame() {
  const [gameState, setGameState] = useState<GameState>({
    board: createInitialBoard(),
    selectedPiece: null,
    validMoves: [],
    isAiThinking: false,
  });

  const selectPiece = useCallback(async (position: Position) => {
    const { board, selectedPiece, isAiThinking } = gameState;
    
    if (isAiThinking || board.currentPlayer !== 'red' || board.gameOver) {
      return;
    }

    const piece = board.pieces[position.row][position.col];
    
    if (!piece || piece.player !== 'red') {
      setGameState(prev => ({
        ...prev,
        selectedPiece: null,
        validMoves: [],
      }));
      return;
    }

    if (selectedPiece && selectedPiece.row === position.row && selectedPiece.col === position.col) {
      setGameState(prev => ({
        ...prev,
        selectedPiece: null,
        validMoves: [],
      }));
      return;
    }

    try {
      const { moves } = await gameApi.getValidMoves(position);
      setGameState(prev => ({
        ...prev,
        selectedPiece: position,
        validMoves: moves,
      }));
    } catch (error) {
      console.error('Failed to get valid moves:', error);
    }
  }, [gameState]);

  const makeMove = useCallback(async (to: Position) => {
    const { board, selectedPiece, isAiThinking } = gameState;
    
    if (!selectedPiece || isAiThinking || board.currentPlayer !== 'red' || board.gameOver) {
      return;
    }

    const move: Move = {
      from: selectedPiece,
      to,
      player: 'red',
    };

    const targetPiece = board.pieces[to.row][to.col];
    const isCapture = targetPiece !== null;

    try {
      setGameState(prev => ({
        ...prev,
        isAiThinking: true,
        selectedPiece: null,
        validMoves: [],
      }));

      const newBoard = await gameApi.makeMove(move);
      
      setGameState(prev => ({
        ...prev,
        board: newBoard,
        isAiThinking: true,
      }));

      if (isCapture) {
        playCaptureSound();
      } else {
        playMoveSound();
      }

      if (newBoard.gameOver) {
        setTimeout(() => {
          if (newBoard.winner === 'red') {
            playWinSound();
          } else {
            playLoseSound();
          }
        }, 300);
        setGameState(prev => ({
          ...prev,
          isAiThinking: false,
        }));
        return;
      }

      if (newBoard.isCheck) {
        setTimeout(playCheckSound, 300);
      }

      if (!newBoard.gameOver) {
        const aiResponse = await gameApi.getAiMove();
        
        const aiToPiece = newBoard.pieces[aiResponse.move.to.row][aiResponse.move.to.col];
        const aiIsCapture = aiToPiece !== null;

        const aiBoard = await gameApi.makeMove(aiResponse.move);
        
        setGameState(prev => ({
          ...prev,
          board: aiBoard,
          isAiThinking: false,
        }));

        setTimeout(() => {
          if (aiIsCapture) {
            playCaptureSound();
          } else {
            playMoveSound();
          }
        }, 300);

        if (aiBoard.gameOver) {
          setTimeout(() => {
            if (aiBoard.winner === 'red') {
              playWinSound();
            } else {
              playLoseSound();
            }
          }, 600);
        } else if (aiBoard.isCheck) {
          setTimeout(playCheckSound, 600);
        }
      } else {
        setGameState(prev => ({
          ...prev,
          board: newBoard,
          isAiThinking: false,
        }));
      }
    } catch (error) {
      console.error('Failed to make move:', error);
      setGameState(prev => ({
        ...prev,
        isAiThinking: false,
      }));
    }
  }, [gameState]);

  const resetGame = useCallback(async () => {
    try {
      const board = await gameApi.resetGame();
      setGameState({
        board,
        selectedPiece: null,
        validMoves: [],
        isAiThinking: false,
      });
      playMoveSound();
    } catch (error) {
      console.error('Failed to reset game:', error);
      setGameState({
        board: createInitialBoard(),
        selectedPiece: null,
        validMoves: [],
        isAiThinking: false,
      });
    }
  }, []);

  return {
    gameState,
    selectPiece,
    makeMove,
    resetGame,
  };
}
