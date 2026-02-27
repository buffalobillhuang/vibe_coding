import random
from typing import List, Optional, Tuple
from .board import Board, Position, Player, Move
from .moves import MoveGenerator

PIECE_VALUES = {
    "rook": 500,
    "horse": 300,
    "cannon": 300,
    "pawn": 50,
    "advisor": 100,
    "elephant": 100,
    "general": 1000,
}

POSITION_BONUS = {
    "pawn": {
        "red": [
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
            [2, 4, 6, 8, 10, 8, 6, 4, 2],
            [4, 8, 12, 16, 20, 16, 12, 8, 4],
            [6, 12, 18, 24, 30, 24, 18, 12, 6],
            [8, 16, 24, 32, 40, 32, 24, 16, 8],
            [10, 20, 30, 40, 50, 40, 30, 20, 10],
            [10, 20, 30, 40, 50, 40, 30, 20, 10],
        ],
        "black": [
            [10, 20, 30, 40, 50, 40, 30, 20, 10],
            [10, 20, 30, 40, 50, 40, 30, 20, 10],
            [8, 16, 24, 32, 40, 32, 24, 16, 8],
            [6, 12, 18, 24, 30, 24, 18, 12, 6],
            [4, 8, 12, 16, 20, 16, 12, 8, 4],
            [2, 4, 6, 8, 10, 8, 6, 4, 2],
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
            [0, 0, 0, 0, 0, 0, 0, 0, 0],
        ],
    },
    "horse": {
        "red": [
            [0, -4, 0, 0, 0, 0, 0, -4, 0],
            [0, 2, 4, 4, 0, 4, 4, 2, 0],
            [4, 2, 8, 8, 4, 8, 8, 2, 4],
            [2, 6, 6, 8, 10, 8, 6, 6, 2],
            [4, 8, 10, 14, 16, 14, 10, 8, 4],
            [4, 10, 14, 18, 20, 18, 14, 10, 4],
            [6, 14, 18, 24, 26, 24, 18, 14, 6],
            [8, 16, 14, 20, 22, 20, 14, 16, 8],
            [12, 14, 12, 18, 18, 18, 12, 14, 12],
            [4, 10, 8, 14, 14, 14, 8, 10, 4],
        ],
        "black": [
            [4, 10, 8, 14, 14, 14, 8, 10, 4],
            [12, 14, 12, 18, 18, 18, 12, 14, 12],
            [8, 16, 14, 20, 22, 20, 14, 16, 8],
            [6, 14, 18, 24, 26, 24, 18, 14, 6],
            [4, 10, 14, 18, 20, 18, 14, 10, 4],
            [4, 8, 10, 14, 16, 14, 10, 8, 4],
            [2, 6, 6, 8, 10, 8, 6, 6, 2],
            [4, 2, 8, 8, 4, 8, 8, 2, 4],
            [0, 2, 4, 4, 0, 4, 4, 2, 0],
            [0, -4, 0, 0, 0, 0, 0, -4, 0],
        ],
    },
    "cannon": {
        "red": [
            [0, 2, 4, 6, 8, 6, 4, 2, 0],
            [2, 4, 6, 8, 10, 8, 6, 4, 2],
            [4, 8, 12, 16, 20, 16, 12, 8, 4],
            [6, 12, 18, 24, 30, 24, 18, 12, 6],
            [8, 16, 24, 32, 40, 32, 24, 16, 8],
            [10, 20, 30, 40, 50, 40, 30, 20, 10],
            [8, 16, 24, 32, 40, 32, 24, 16, 8],
            [6, 12, 18, 24, 30, 24, 18, 12, 6],
            [4, 8, 12, 16, 20, 16, 12, 8, 4],
            [2, 4, 6, 8, 10, 8, 6, 4, 2],
        ],
        "black": [
            [2, 4, 6, 8, 10, 8, 6, 4, 2],
            [4, 8, 12, 16, 20, 16, 12, 8, 4],
            [6, 12, 18, 24, 30, 24, 18, 12, 6],
            [8, 16, 24, 32, 40, 32, 24, 16, 8],
            [10, 20, 30, 40, 50, 40, 30, 20, 10],
            [8, 16, 24, 32, 40, 32, 24, 16, 8],
            [6, 12, 18, 24, 30, 24, 18, 12, 6],
            [4, 8, 12, 16, 20, 16, 12, 8, 4],
            [2, 4, 6, 8, 10, 8, 6, 4, 2],
            [0, 2, 4, 6, 8, 6, 4, 2, 0],
        ],
    },
    "rook": {
        "red": [
            [6, 8, 6, 10, 12, 10, 6, 8, 6],
            [6, 10, 8, 12, 14, 12, 8, 10, 6],
            [4, 12, 6, 14, 14, 14, 6, 12, 4],
            [12, 16, 14, 20, 22, 20, 14, 16, 12],
            [12, 14, 12, 18, 18, 18, 12, 14, 12],
            [14, 18, 16, 22, 24, 22, 16, 18, 14],
            [12, 12, 12, 18, 18, 18, 12, 12, 12],
            [16, 20, 18, 24, 26, 24, 18, 20, 16],
            [16, 16, 16, 20, 20, 20, 16, 16, 16],
            [16, 20, 18, 24, 26, 24, 18, 20, 16],
        ],
        "black": [
            [16, 20, 18, 24, 26, 24, 18, 20, 16],
            [16, 16, 16, 20, 20, 20, 16, 16, 16],
            [16, 20, 18, 24, 26, 24, 18, 20, 16],
            [12, 12, 12, 18, 18, 18, 12, 12, 12],
            [14, 18, 16, 22, 24, 22, 16, 18, 14],
            [12, 14, 12, 18, 18, 18, 12, 14, 12],
            [12, 16, 14, 20, 22, 20, 14, 16, 12],
            [4, 12, 6, 14, 14, 14, 6, 12, 4],
            [6, 10, 8, 12, 14, 12, 8, 10, 6],
            [6, 8, 6, 10, 12, 10, 6, 8, 6],
        ],
    },
}


class AI:
    def __init__(self, depth: int = 2):
        self.depth = depth

    def evaluate(self, board: Board) -> float:
        score = 0.0

        for r in range(10):
            for c in range(9):
                piece = board.get_piece(Position(r, c))
                if piece:
                    value = PIECE_VALUES.get(piece.type.value, 0)

                    if piece.type.value in POSITION_BONUS:
                        bonus_map = POSITION_BONUS[piece.type.value]
                        player_key = "red" if piece.player == Player.RED else "black"
                        if player_key in bonus_map:
                            value += bonus_map[player_key][r][c]

                    if piece.player == Player.RED:
                        score += value
                    else:
                        score -= value

        return score

    def get_all_moves(self, board: Board, player: Player) -> List[Move]:
        moves = []
        for r in range(10):
            for c in range(9):
                piece = board.get_piece(Position(r, c))
                if piece and piece.player == player:
                    valid_positions = MoveGenerator.get_valid_moves(
                        board, Position(r, c)
                    )
                    for to_pos in valid_positions:
                        moves.append(Move(Position(r, c), to_pos, player))
        return moves

    def minimax(
        self, board: Board, depth: int, alpha: float, beta: float, maximizing: bool
    ) -> float:
        if depth == 0 or board.game_over:
            return self.evaluate(board)

        current_player = board.current_player

        if maximizing:
            max_eval = float("-inf")
            moves = self.get_all_moves(board, current_player)
            for move in moves:
                new_board = MoveGenerator.make_move(board, move)
                eval_score = self.minimax(new_board, depth - 1, alpha, beta, False)
                max_eval = max(max_eval, eval_score)
                alpha = max(alpha, eval_score)
                if beta <= alpha:
                    break
            return max_eval
        else:
            min_eval = float("inf")
            moves = self.get_all_moves(board, current_player)
            for move in moves:
                new_board = MoveGenerator.make_move(board, move)
                eval_score = self.minimax(new_board, depth - 1, alpha, beta, True)
                min_eval = min(min_eval, eval_score)
                beta = min(beta, eval_score)
                if beta <= alpha:
                    break
            return min_eval

    def get_best_move(self, board: Board) -> Optional[Move]:
        moves = self.get_all_moves(board, Player.BLACK)
        if not moves:
            return None

        best_move = None
        best_score = float("inf")
        alpha = float("-inf")
        beta = float("inf")

        random.shuffle(moves)

        for move in moves:
            new_board = MoveGenerator.make_move(board, move)
            score = self.minimax(new_board, self.depth - 1, alpha, beta, True)

            if score < best_score:
                best_score = score
                best_move = move

            beta = min(beta, score)
            if beta <= alpha:
                break

        return best_move
