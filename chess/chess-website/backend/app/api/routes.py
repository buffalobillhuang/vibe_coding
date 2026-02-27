from fastapi import APIRouter, HTTPException
from pydantic import BaseModel
from typing import List, Optional
from app.engine import Board, Position, Player, Move, MoveGenerator, AI

router = APIRouter()

board = Board()
ai = AI(depth=2)


class PositionInput(BaseModel):
    row: int
    col: int


class MoveInput(BaseModel):
    from_pos: PositionInput
    to_pos: PositionInput
    player: str


@router.post("/game/start")
async def start_game():
    global board
    board = Board()
    return board.to_dict()


@router.get("/game/board")
async def get_board():
    return board.to_dict()


@router.post("/game/move")
async def make_move(move_input: MoveInput):
    global board

    from_pos = Position(move_input.from_pos.row, move_input.from_pos.col)
    to_pos = Position(move_input.to_pos.row, move_input.to_pos.col)

    piece = board.get_piece(from_pos)
    if not piece:
        raise HTTPException(status_code=400, detail="No piece at from position")

    valid_moves = MoveGenerator.get_valid_moves(board, from_pos)
    if to_pos not in valid_moves:
        raise HTTPException(status_code=400, detail="Invalid move")

    move = Move(from_pos, to_pos, piece.player)
    board = MoveGenerator.make_move(board, move)

    return board.to_dict()


@router.get("/game/ai-move")
async def get_ai_move():
    global board

    best_move = ai.get_best_move(board)
    if not best_move:
        raise HTTPException(status_code=400, detail="No valid AI move")

    return {"move": best_move.to_dict()}


@router.post("/game/reset")
async def reset_game():
    global board
    board = Board()
    return board.to_dict()


@router.post("/game/valid-moves")
async def get_valid_moves(position: PositionInput):
    pos = Position(position.row, position.col)
    piece = board.get_piece(pos)

    if not piece:
        return {"moves": []}

    moves = MoveGenerator.get_valid_moves(board, pos)
    return {"moves": [m.to_dict() for m in moves]}
