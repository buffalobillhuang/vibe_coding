import React from 'react';
import { type Piece as PieceType, PIECE_NAMES, PIECE_NAMES_BLACK } from '../types';

interface PieceProps {
  piece: PieceType;
  onClick?: () => void;
  selected?: boolean;
  validMove?: boolean;
}

export const Piece: React.FC<PieceProps> = ({ 
  piece, 
  onClick, 
  selected, 
  validMove 
}) => {
  const isRed = piece.player === 'red';
  const displayChar = isRed 
    ? PIECE_NAMES[piece.type] 
    : PIECE_NAMES_BLACK[piece.type];

  return (
    <div
      onClick={onClick}
      style={{
        width: '100%',
        height: '100%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        cursor: onClick ? 'pointer' : 'default',
        position: 'relative',
      }}
    >
      {validMove && (
        <div
          style={{
            position: 'absolute',
            width: '30%',
            height: '30%',
            borderRadius: '50%',
            backgroundColor: 'rgba(0, 128, 0, 0.5)',
          }}
        />
      )}
      <div
        style={{
          width: '85%',
          height: '85%',
          borderRadius: '50%',
          backgroundColor: isRed ? '#d32f2f' : '#212121',
          border: selected ? '3px solid gold' : '2px solid #fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          color: isRed ? '#fff' : '#ffeb3b',
          fontSize: '24px',
          fontWeight: 'bold',
          fontFamily: 'serif',
          boxShadow: selected 
            ? '0 0 10px gold, inset 0 0 5px rgba(0,0,0,0.5)' 
            : 'inset 0 0 5px rgba(0,0,0,0.5)',
          zIndex: validMove ? 1 : 2,
        }}
      >
        {displayChar}
      </div>
    </div>
  );
};
