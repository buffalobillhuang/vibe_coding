import React from 'react';
import { type BoardState, type Position } from '../types';
import { Piece } from './Piece';

interface BoardProps {
  board: BoardState;
  selectedPiece: Position | null;
  validMoves: Position[];
  onPieceClick: (position: Position) => void;
  onMove: (position: Position) => void;
}

const CELL_SIZE = 60;
const BOARD_WIDTH = CELL_SIZE * 9;
const BOARD_HEIGHT = CELL_SIZE * 10;

const isValidMove = (pos: Position, validMoves: Position[]): boolean => {
  return validMoves.some(m => m.row === pos.row && m.col === pos.col);
};

export const Board: React.FC<BoardProps> = ({
  board,
  selectedPiece,
  validMoves,
  onPieceClick,
  onMove,
}) => {
  const handleCellClick = (row: number, col: number) => {
    const position: Position = { row, col };
    const piece = board.pieces[row][col];
    const isValid = isValidMove(position, validMoves);

    if (isValid) {
      onMove(position);
    } else if (piece && piece.player === board.currentPlayer) {
      onPieceClick(position);
    }
  };

  const renderBoard = () => {
    const cells = [];
    
    for (let row = 0; row < 10; row++) {
      for (let col = 0; col < 9; col++) {
        const position = { row, col };
        const piece = board.pieces[row][col];
        const isSelected = selectedPiece?.row === row && selectedPiece?.col === col;
        const isValidMoveCell = isValidMove(position, validMoves);

        cells.push(
          <div
            key={`${row}-${col}`}
            onClick={() => handleCellClick(row, col)}
            style={{
              position: 'absolute',
              left: col * CELL_SIZE,
              top: row * CELL_SIZE,
              width: CELL_SIZE,
              height: CELL_SIZE,
              backgroundColor: (row + col) % 2 === 0 ? '#DEB887' : '#F5DEB3',
              border: '1px solid #8B4513',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              cursor: 'pointer',
            }}
          >
            {piece && (
              <Piece 
                piece={piece} 
                selected={isSelected}
                validMove={isValidMoveCell}
              />
            )}
            {isValidMoveCell && !piece && (
              <div
                style={{
                  width: '30%',
                  height: '30%',
                  borderRadius: '50%',
                  backgroundColor: 'rgba(0, 128, 0, 0.5)',
                }}
              />
            )}
          </div>
        );
      }
    }
    return cells;
  };

  const renderLines = () => {
    const lines = [];
    
    for (let i = 0; i < 10; i++) {
      lines.push(
        <line
          key={`h-${i}`}
          x1={0}
          y1={i * CELL_SIZE}
          x2={BOARD_WIDTH}
          y2={i * CELL_SIZE}
          stroke="#8B4513"
          strokeWidth="2"
        />
      );
    }
    
    for (let i = 0; i < 9; i++) {
      if (i !== 0 && i !== 8) {
        lines.push(
          <line
            key={`v-top-${i}`}
            x1={i * CELL_SIZE}
            y1={0}
            x2={i * CELL_SIZE}
            y2={CELL_SIZE * 4}
            stroke="#8B4513"
            strokeWidth="2"
          />
        );
        lines.push(
          <line
            key={`v-bottom-${i}`}
            x1={i * CELL_SIZE}
            y1={CELL_SIZE * 5}
            x2={i * CELL_SIZE}
            y2={BOARD_HEIGHT}
            stroke="#8B4513"
            strokeWidth="2"
          />
        );
      } else {
        lines.push(
          <line
            key={`v-${i}`}
            x1={i * CELL_SIZE}
            y1={0}
            x2={i * CELL_SIZE}
            y2={BOARD_HEIGHT}
            stroke="#8B4513"
            strokeWidth="2"
          />
        );
      }
    }

    lines.push(
      <line
        key="palace-left-1"
        x1={CELL_SIZE * 3}
        y1={0}
        x2={CELL_SIZE * 5}
        y2={CELL_SIZE * 2}
        stroke="#8B4513"
        strokeWidth="2"
      />
    );
    lines.push(
      <line
        key="palace-right-1"
        x1={CELL_SIZE * 5}
        y1={0}
        x2={CELL_SIZE * 3}
        y2={CELL_SIZE * 2}
        stroke="#8B4513"
        strokeWidth="2"
      />
    );
    lines.push(
      <line
        key="palace-left-2"
        x1={CELL_SIZE * 3}
        y1={BOARD_HEIGHT}
        x2={CELL_SIZE * 5}
        y2={BOARD_HEIGHT - CELL_SIZE * 2}
        stroke="#8B4513"
        strokeWidth="2"
      />
    );
    lines.push(
      <line
        key="palace-right-2"
        x1={CELL_SIZE * 5}
        y1={BOARD_HEIGHT}
        x2={CELL_SIZE * 3}
        y2={BOARD_HEIGHT - CELL_SIZE * 2}
        stroke="#8B4513"
        strokeWidth="2"
      />
    );

    return lines;
  };

  const renderLabels = () => {
    const labels = [];
    const cols = ['九', '八', '七', '六', '五', '四', '三', '二', '一'];
    const rows = ['一', '二', '三', '四', '五', '六', '七', '八', '九', '十'];
    
    for (let i = 0; i < 9; i++) {
      labels.push(
        <text
          key={`col-${i}`}
          x={i * CELL_SIZE + CELL_SIZE / 2}
          y={-5}
          textAnchor="middle"
          fill="#8B4513"
          fontSize="12"
        >
          {cols[i]}
        </text>
      );
    }
    
    for (let i = 0; i < 10; i++) {
      labels.push(
        <text
          key={`row-${i}`}
          x={-15}
          y={i * CELL_SIZE + CELL_SIZE / 2 + 4}
          textAnchor="middle"
          fill="#8B4513"
          fontSize="12"
        >
          {rows[i]}
        </text>
      );
    }
    
    return labels;
  };

  return (
    <div style={{ position: 'relative' }}>
      <svg 
        width={BOARD_WIDTH} 
        height={BOARD_HEIGHT + 30}
        style={{ fontFamily: 'serif' }}
      >
        <g transform="translate(15, 25)">
          {renderLabels()}
          {renderLines()}
        </g>
      </svg>
      <div style={{ position: 'absolute', top: 0, left: 0 }}>
        {renderBoard()}
      </div>
    </div>
  );
};
