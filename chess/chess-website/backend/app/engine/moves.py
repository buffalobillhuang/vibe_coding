from typing import List
from .board import Board, Position, Player, Piece, PieceType, Move


class MoveGenerator:
    @staticmethod
    def is_valid_position(pos: Position) -> bool:
        return 0 <= pos.row < 10 and 0 <= pos.col < 9

    @staticmethod
    def is_in_palace(pos: Position, player: Player) -> bool:
        if player == Player.RED:
            return pos.row >= 7 and pos.row <= 9 and 3 <= pos.col <= 5
        else:
            return pos.row >= 0 and pos.row <= 2 and 3 <= pos.col <= 5

    @staticmethod
    def is_across_river(pos: Position, player: Player) -> bool:
        if player == Player.RED:
            return pos.row <= 4
        else:
            return pos.row >= 5

    @staticmethod
    def get_rook_moves(board: Board, pos: Position) -> List[Position]:
        moves = []
        piece = board.get_piece(pos)
        if not piece:
            return moves

        directions = [(0, 1), (0, -1), (1, 0), (-1, 0)]
        for dr, dc in directions:
            r, c = pos.row + dr, pos.col + dc
            while MoveGenerator.is_valid_position(Position(r, c)):
                target = board.get_piece(Position(r, c))
                if target is None:
                    moves.append(Position(r, c))
                else:
                    if target.player != piece.player:
                        moves.append(Position(r, c))
                    break
                r += dr
                c += dc
        return moves

    @staticmethod
    def get_horse_moves(board: Board, pos: Position) -> List[Position]:
        moves = []
        piece = board.get_piece(pos)
        if not piece:
            return moves

        jumps = [
            (-2, -1),
            (-2, 1),
            (2, -1),
            (2, 1),
            (-1, -2),
            (-1, 2),
            (1, -2),
            (1, 2),
        ]
        blocks = [
            (-1, 0),
            (-1, 0),
            (1, 0),
            (1, 0),
            (0, -1),
            (0, 1),
            (0, -1),
            (0, 1),
        ]

        for (dr, dc), (br, bc) in zip(jumps, blocks):
            r, c = pos.row + dr, pos.col + dc
            br, bc = pos.row + br, pos.col + bc

            if MoveGenerator.is_valid_position(Position(r, c)):
                if board.get_piece(Position(br, bc)) is None:
                    target = board.get_piece(Position(r, c))
                    if target is None or target.player != piece.player:
                        moves.append(Position(r, c))
        return moves

    @staticmethod
    def get_elephant_moves(board: Board, pos: Position) -> List[Position]:
        moves = []
        piece = board.get_piece(pos)
        if not piece:
            return moves

        if piece.player == Player.RED and pos.row < 5:
            return moves
        if piece.player == Player.BLACK and pos.row > 4:
            return moves

        jumps = [(-2, -2), (-2, 2), (2, -2), (2, 2)]
        blocks = [(-1, -1), (-1, 1), (1, -1), (1, 1)]

        for (dr, dc), (br, bc) in zip(jumps, blocks):
            r, c = pos.row + dr, pos.col + dc
            br, bc = pos.row + br, pos.col + bc

            if MoveGenerator.is_valid_position(Position(r, c)):
                if board.get_piece(Position(br, bc)) is None:
                    target = board.get_piece(Position(r, c))
                    if target is None or target.player != piece.player:
                        moves.append(Position(r, c))
        return moves

    @staticmethod
    def get_advisor_moves(board: Board, pos: Position) -> List[Position]:
        moves = []
        piece = board.get_piece(pos)
        if not piece:
            return moves

        jumps = [(-1, -1), (-1, 1), (1, -1), (1, 1)]
        for dr, dc in jumps:
            r, c = pos.row + dr, pos.col + dc
            if MoveGenerator.is_in_palace(Position(r, c), piece.player):
                target = board.get_piece(Position(r, c))
                if target is None or target.player != piece.player:
                    moves.append(Position(r, c))
        return moves

    @staticmethod
    def get_general_moves(board: Board, pos: Position) -> List[Position]:
        moves = []
        piece = board.get_piece(pos)
        if not piece:
            return moves

        jumps = [(-1, 0), (1, 0), (0, -1), (0, 1)]
        for dr, dc in jumps:
            r, c = pos.row + dr, pos.col + dc
            if MoveGenerator.is_in_palace(Position(r, c), piece.player):
                target = board.get_piece(Position(r, c))
                if target is None or target.player != piece.player:
                    moves.append(Position(r, c))

        return moves

    @staticmethod
    def get_cannon_moves(board: Board, pos: Position) -> List[Position]:
        moves = []
        piece = board.get_piece(pos)
        if not piece:
            return moves

        directions = [(0, 1), (0, -1), (1, 0), (-1, 0)]
        for dr, dc in directions:
            r, c = pos.row + dr, pos.col + dc
            jumped = False
            while MoveGenerator.is_valid_position(Position(r, c)):
                target = board.get_piece(Position(r, c))
                if not jumped:
                    if target is None:
                        moves.append(Position(r, c))
                    else:
                        jumped = True
                else:
                    if target is not None:
                        if target.player != piece.player:
                            moves.append(Position(r, c))
                        break
                r += dr
                c += dc
        return moves

    @staticmethod
    def get_pawn_moves(board: Board, pos: Position) -> List[Position]:
        moves = []
        piece = board.get_piece(pos)
        if not piece:
            return moves

        across = MoveGenerator.is_across_river(pos, piece.player)

        forward = -1 if piece.player == Player.RED else 1
        r, c = pos.row + forward, pos.col
        if MoveGenerator.is_valid_position(Position(r, c)):
            target = board.get_piece(Position(r, c))
            if target is None or target.player != piece.player:
                moves.append(Position(r, c))

        if across:
            for dc in [-1, 1]:
                r, c = pos.row, pos.col + dc
                if MoveGenerator.is_valid_position(Position(r, c)):
                    target = board.get_piece(Position(r, c))
                    if target is None or target.player != piece.player:
                        moves.append(Position(r, c))

        return moves

    @staticmethod
    def get_valid_moves(board: Board, pos: Position) -> List[Position]:
        piece = board.get_piece(pos)
        if not piece:
            return []

        if piece.type == PieceType.ROOK:
            return MoveGenerator.get_rook_moves(board, pos)
        elif piece.type == PieceType.HORSE:
            return MoveGenerator.get_horse_moves(board, pos)
        elif piece.type == PieceType.ELEPHANT:
            return MoveGenerator.get_elephant_moves(board, pos)
        elif piece.type == PieceType.ADVISOR:
            return MoveGenerator.get_advisor_moves(board, pos)
        elif piece.type == PieceType.GENERAL:
            return MoveGenerator.get_general_moves(board, pos)
        elif piece.type == PieceType.CANNON:
            return MoveGenerator.get_cannon_moves(board, pos)
        elif piece.type == PieceType.PAWN:
            return MoveGenerator.get_pawn_moves(board, pos)

        return []

    @staticmethod
    def find_general(board: Board, player: Player) -> Position:
        for r in range(10):
            for c in range(9):
                piece = board.get_piece(Position(r, c))
                if piece and piece.type == PieceType.GENERAL and piece.player == player:
                    return Position(r, c)
        return Position(0, 0)

    @staticmethod
    def is_check(board: Board, player: Player) -> bool:
        general_pos = MoveGenerator.find_general(board, player)
        opponent = Player.BLACK if player == Player.RED else Player.RED

        for r in range(10):
            for c in range(9):
                piece = board.get_piece(Position(r, c))
                if piece and piece.player == opponent:
                    moves = MoveGenerator.get_valid_moves(board, Position(r, c))
                    if general_pos in moves:
                        return True
        return False

    @staticmethod
    def make_move(board: Board, move: Move) -> Board:
        new_board = board.copy()
        piece = new_board.get_piece(move.from_pos)
        new_board.set_piece(move.from_pos, None)
        new_board.set_piece(move.to_pos, piece)

        new_board.current_player = (
            Player.BLACK if board.current_player == Player.RED else Player.RED
        )

        if MoveGenerator.is_check(new_board, new_board.current_player):
            new_board.is_check = True

        if MoveGenerator.is_check(new_board, Player.RED):
            if MoveGenerator.is_checkmate(new_board, Player.RED):
                new_board.game_over = True
                new_board.winner = Player.BLACK

        if MoveGenerator.is_check(new_board, Player.BLACK):
            if MoveGenerator.is_checkmate(new_board, Player.BLACK):
                new_board.game_over = True
                new_board.winner = Player.RED

        return new_board

    @staticmethod
    def is_checkmate(board: Board, player: Player) -> bool:
        for r in range(10):
            for c in range(9):
                piece = board.get_piece(Position(r, c))
                if piece and piece.player == player:
                    moves = MoveGenerator.get_valid_moves(board, Position(r, c))
                    for move in moves:
                        test_board = board.copy()
                        test_piece = test_board.get_piece(Position(r, c))
                        test_board.set_piece(Position(r, c), None)
                        test_board.set_piece(move, test_piece)
                        if not MoveGenerator.is_check(test_board, player):
                            return False
        return True
