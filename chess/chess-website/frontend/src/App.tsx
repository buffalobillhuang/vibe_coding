import { Board } from './components/Board';
import { GameControls } from './components/GameControls';
import { useGame } from './hooks/useGame';

function App() {
  const { gameState, selectPiece, makeMove, resetGame } = useGame();

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      backgroundColor: '#f5f5dc',
      padding: '20px',
    }}>
      <h1 style={{
        fontSize: '32px',
        color: '#8B4513',
        marginBottom: '20px',
        fontFamily: 'serif',
      }}>
        中国象棋 - 人机对弈
      </h1>
      
      <Board
        board={gameState.board}
        selectedPiece={gameState.selectedPiece}
        validMoves={gameState.validMoves}
        onPieceClick={selectPiece}
        onMove={makeMove}
      />
      
      <GameControls
        onReset={resetGame}
        isAiThinking={gameState.isAiThinking}
        currentPlayer={gameState.board.currentPlayer}
        gameOver={gameState.board.gameOver}
        winner={gameState.board.winner}
      />
      
      <div style={{
        marginTop: '20px',
        color: '#666',
        fontSize: '14px',
      }}>
        红方（你）vs 黑方（电脑）
      </div>
    </div>
  );
}

export default App;
