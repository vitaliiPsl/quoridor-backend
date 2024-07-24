package errors

import "errors"

var (
	ErrBadRequest           = errors.New("bad_request")
	ErrInternalError        = errors.New("internal_error")
	ErrGameNotFound         = errors.New("game_not_found")
	ErrGameNotInProgress    = errors.New("game_not_in_progress")
	ErrNotPlayersTurn       = errors.New("not_players_turn")
	ErrInvalidMove          = errors.New("invalid_move")
	ErrInvalidWallPlacement = errors.New("invalid_wall_placement")
	ErrNotAPlayer           = errors.New("not_a_player")
)
