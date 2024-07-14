package game

import "time"

type GameStatus string

const (
	GameStatusPending    GameStatus = "pending"
	GameStatusAborted    GameStatus = "aborted"
	GameStatusInProgress GameStatus = "in_progress"
	GameStatusCompleted  GameStatus = "completed"
)

type Direction string

const (
	Horizontal Direction = "horizontal"
	Vertical   Direction = "vertical"
)

type MoveType string

const (
	MoveTypeMove      MoveType = "move"
	MoveTypePlaceWall MoveType = "place_wall"
)

type Game struct {
	GameId      string     `bson:"_id" json:"game_id"`
	GameStatus  GameStatus `bson:"status" json:"status"`
	Winner      string     `bson:"winner,omitempty" json:"winner,omitempty"` // id of the winner
	Turn        string     `bson:"turn" json:"turn"`                         // id of the player
	Player1     *Player    `bson:"player_1" json:"player_1"`
	Player2     *Player    `bson:"player_2" json:"player_2"`
	Walls       []*Wall    `bson:"walls" json:"walls"`
	Moves       []*Move    `bson:"moves" json:"moves"`
	CreatedAt   time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at" json:"updated_at"`
	CompletedAt *time.Time `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
}

type Player struct {
	UserId   string    `bson:"user_id" json:"user_id"`
	Position *Position `bson:"position" json:"position"`
	Goal     int       `bson:"goal" json:"goal"`   // row user needs to get to to win the game
	Walls    int       `bson:"walls" json:"walls"` // number of walls available
}

type Move struct {
	UserId    string    `bson:"user_id" json:"user_id"`
	Type      MoveType  `bson:"type" json:"type"`
	Position  *Position `bson:"position,omitempty" json:"position,omitempty"`
	Wall      *Wall     `bson:"wall,omitempty" json:"wall,omitempty"`
	Timestamp time.Time `bson:"timestamp" json:"timestamp"`
}

/*
each wall spans two cells:
- a horizontal wall starts at {pos.X, pos.Y} and ends at {pos.X + 1, pos.Y}
- a vertical wall starts at {pos.X, pos.Y} and ends at {pos.X, pos.Y + 1}
*/
type Wall struct {
	Direction Direction `bson:"direction" json:"direction"`
	Pos1      *Position `bson:"position_1" json:"position_1"`
	Pos2      *Position `bson:"position_2" json:"position_2"`
}

type Position struct {
	X int `bson:"x" json:"x"`
	Y int `bson:"y" json:"y"`
}
