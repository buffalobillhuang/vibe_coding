from typing import Optional
from enum import Enum


class Player(str, Enum):
    RED = "red"
    BLACK = "black"


class PieceType(str, Enum):
    ROOK = "rook"
    HORSE = "horse"
    ELEPHANT = "elephant"
    ADVISOR = "advisor"
    GENERAL = "general"
    CANNON = "cannon"
    PAWN = "pawn"


class Piece:
    def __init__(self, piece_type: PieceType, player: Player):
        self.type = piece_type
        self.player = player

    def to_dict(self):
        return {"type": self.type.value, "player": self.player.value}


class Position:
    def __init__(self, row: int, col: int):
        self.row = row
        self.col = col

    def to_dict(self):
        return {"row": self.row, "col": self.col}

    def __eq__(self, other):
        return self.row == other.row and self.col == other.col


class Move:
    def __init__(self, from_pos: Position, to_pos: Position, player: Player):
        self.from_pos = from_pos
        self.to_pos = to_pos
        self.player = player

    def to_dict(self):
        return {
            "from": self.from_pos.to_dict(),
            "to": self.to_pos.to_dict(),
            "player": self.player.value,
        }


INITIAL_BOARD = [
    [
        Piece(PieceType.ROOK, Player.BLACK),
        Piece(PieceType.HORSE, Player.BLACK),
        Piece(PieceType.ELEPHANT, Player.BLACK),
        Piece(PieceType.ADVISOR, Player.BLACK),
        Piece(PieceType.GENERAL, Player.BLACK),
        Piece(PieceType.ADVISOR, Player.BLACK),
        Piece(PieceType.ELEPHANT, Player.BLACK),
        Piece(PieceType.HORSE, Player.BLACK),
        Piece(PieceType.ROOK, Player.BLACK),
    ],
    [None] * 9,
    [
        None,
        Piece(PieceType.CANNON, Player.BLACK),
        None,
        None,
        None,
        None,
        None,
        Piece(PieceType.CANNON, Player.BLACK),
        None,
    ],
    [
        Piece(PieceType.PAWN, Player.BLACK),
        None,
        Piece(PieceType.PAWN, Player.BLACK),
        None,
        Piece(PieceType.PAWN, Player.BLACK),
        None,
        Piece(PieceType.PAWN, Player.BLACK),
        None,
        Piece(PieceType.PAWN, Player.BLACK),
    ],
    [None] * 9,
    [None] * 9,
    [
        Piece(PieceType.PAWN, Player.RED),
        None,
        Piece(PieceType.PAWN, Player.RED),
        None,
        Piece(PieceType.PAWN, Player.RED),
        None,
        Piece(PieceType.PAWN, Player.RED),
        None,
        Piece(PieceType.PAWN, Player.RED),
    ],
    [
        None,
        Piece(PieceType.CANNON, Player.RED),
        None,
        None,
        None,
        None,
        None,
        Piece(PieceType.CANNON, Player.RED),
        None,
    ],
    [None] * 9,
    [
        Piece(PieceType.ROOK, Player.RED),
        Piece(PieceType.HORSE, Player.RED),
        Piece(PieceType.ELEPHANT, Player.RED),
        Piece(PieceType.ADVISOR, Player.RED),
        Piece(PieceType.GENERAL, Player.RED),
        Piece(PieceType.ADVISOR, Player.RED),
        Piece(PieceType.ELEPHANT, Player.RED),
        Piece(PieceType.HORSE, Player.RED),
        Piece(PieceType.ROOK, Player.RED),
    ],
]


class Board:
    def __init__(self):
        self.pieces = [row[:] for row in INITIAL_BOARD]
        self.current_player = Player.RED
        self.is_check = False
        self.game_over = False
        self.winner: Optional[Player] = None

    def copy(self):
        new_board = Board()
        new_board.pieces = [row[:] for row in self.pieces]
        new_board.current_player = self.current_player
        new_board.is_check = self.is_check
        new_board.game_over = self.game_over
        new_board.winner = self.winner
        return new_board

    def get_piece(self, pos: Position) -> Optional[Piece]:
        if 0 <= pos.row < 10 and 0 <= pos.col < 9:
            return self.pieces[pos.row][pos.col]
        return None

    def set_piece(self, pos: Position, piece: Optional[Piece]):
        if 0 <= pos.row < 10 and 0 <= pos.col < 9:
            self.pieces[pos.row][pos.col] = piece

    def to_dict(self):
        return {
            "pieces": [
                [p.to_dict() if p else None for p in row] for row in self.pieces
            ],
            "currentPlayer": self.current_player.value,
            "isCheck": self.is_check,
            "gameOver": self.game_over,
            "winner": self.winner.value if self.winner else None,
        }
