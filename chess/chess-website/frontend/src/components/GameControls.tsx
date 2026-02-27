import React from 'react';

interface GameControlsProps {
  onReset: () => void;
  isAiThinking: boolean;
  currentPlayer: 'red' | 'black';
  gameOver: boolean;
  winner: 'red' | 'black' | null;
}

export const GameControls: React.FC<GameControlsProps> = ({
  onReset,
  isAiThinking,
  currentPlayer,
  gameOver,
  winner,
}) => {
  const getStatusText = () => {
    if (gameOver && winner) {
      return winner === 'red' ? '恭喜！你赢了！' : '很遗憾，你输了！';
    }
    if (isAiThinking) {
      return '电脑思考中...';
    }
    return currentPlayer === 'red' ? '轮到你走棋 (红方)' : '电脑走棋 (黑方)';
  };

  return (
    <div style={{ 
      display: 'flex', 
      flexDirection: 'column', 
      alignItems: 'center', 
      gap: '16px',
      marginTop: '20px',
    }}>
      <div style={{
        padding: '12px 24px',
        backgroundColor: gameOver 
          ? (winner === 'red' ? '#4CAF50' : '#f44336')
          : isAiThinking 
            ? '#FF9800' 
            : currentPlayer === 'red' ? '#d32f2f' : '#212121',
        color: '#fff',
        borderRadius: '8px',
        fontSize: '18px',
        fontWeight: 'bold',
      }}>
        {getStatusText()}
      </div>
      
      <button
        onClick={onReset}
        disabled={isAiThinking}
        style={{
          padding: '12px 32px',
          fontSize: '16px',
          fontWeight: 'bold',
          backgroundColor: '#1976D2',
          color: '#fff',
          border: 'none',
          borderRadius: '8px',
          cursor: isAiThinking ? 'not-allowed' : 'pointer',
          opacity: isAiThinking ? 0.6 : 1,
        }}
      >
        重新开始
      </button>
    </div>
  );
};
