export type Player = 'red' | 'black';

export type PieceType = 
  | 'rook'    // 车
  | 'horse'   // 马
  | 'elephant' // 相
  | 'advisor'  // 士
  | 'general'  // 帅/将
  | 'cannon'   // 炮
  | 'pawn';    // 兵/卒

export interface Piece {
  type: PieceType;
  player: Player;
}

export interface Position {
  row: number;
  col: number;
}

export interface Move {
  from: Position;
  to: Position;
  player: Player;
}

export interface BoardState {
  pieces: (Piece | null)[][];
  currentPlayer: Player;
  isCheck: boolean;
  gameOver: boolean;
  winner: Player | null;
}

export interface GameState {
  board: BoardState;
  selectedPiece: Position | null;
  validMoves: Position[];
  isAiThinking: boolean;
}

export const INITIAL_BOARD: (Piece | null)[][] = [
  [
    { type: 'rook', player: 'black' },
    { type: 'horse', player: 'black' },
    { type: 'elephant', player: 'black' },
    { type: 'advisor', player: 'black' },
    { type: 'general', player: 'black' },
    { type: 'advisor', player: 'black' },
    { type: 'elephant', player: 'black' },
    { type: 'horse', player: 'black' },
    { type: 'rook', player: 'black' },
  ],
  [
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
  ],
  [
    null,
    { type: 'cannon', player: 'black' },
    null,
    null,
    null,
    null,
    null,
    { type: 'cannon', player: 'black' },
    null,
  ],
  [
    { type: 'pawn', player: 'black' },
    null,
    { type: 'pawn', player: 'black' },
    null,
    { type: 'pawn', player: 'black' },
    null,
    { type: 'pawn', player: 'black' },
    null,
    { type: 'pawn', player: 'black' },
  ],
  [
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
  ],
  [
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
  ],
  [
    { type: 'pawn', player: 'red' },
    null,
    { type: 'pawn', player: 'red' },
    null,
    { type: 'pawn', player: 'red' },
    null,
    { type: 'pawn', player: 'red' },
    null,
    { type: 'pawn', player: 'red' },
  ],
  [
    null,
    { type: 'cannon', player: 'red' },
    null,
    null,
    null,
    null,
    null,
    { type: 'cannon', player: 'red' },
    null,
  ],
  [
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
    null,
  ],
  [
    { type: 'rook', player: 'red' },
    { type: 'horse', player: 'red' },
    { type: 'elephant', player: 'red' },
    { type: 'advisor', player: 'red' },
    { type: 'general', player: 'red' },
    { type: 'advisor', player: 'red' },
    { type: 'elephant', player: 'red' },
    { type: 'horse', player: 'red' },
    { type: 'rook', player: 'red' },
  ],
];

export const PIECE_NAMES: Record<PieceType, string> = {
  rook: '車',
  horse: '馬',
  elephant: '相',
  advisor: '士',
  general: '帥',
  cannon: '炮',
  pawn: '兵',
};

export const PIECE_NAMES_BLACK: Record<PieceType, string> = {
  rook: '俥',
  horse: '傌',
  elephant: '象',
  advisor: '仕',
  general: '將',
  cannon: '砲',
  pawn: '卒',
};
