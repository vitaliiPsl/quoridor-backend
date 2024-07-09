package game

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

type GameState struct {
	GameId     string     `json:"game_id"`
	GameStatus GameStatus `json:"status"`
	Winner     string     `json:"winner,omitempty"` // id of the winner
	Turn       string     `json:"turn"`             // id of the player
	Player1    *Player    `json:"player_1"`
	Player2    *Player    `json:"player_2"`
	Walls      []*Wall    `json:"walls"`
}

type Player struct {
	UserId   string    `json:"user_id"`
	Position *Position `json:"position"`
	Goal     int       `json:"goal"`  // raw user need to get to to win the game
	Walls    int       `json:"walls"` // number of walls avaiable
}

/*
i'm using two positions of the cells that the wall is between to simplify checks.
each wall spans two cells:
- a horizontal wall starts at {pos.X, pos.Y} and ends at {pos.X + 1, pos.Y}.
- a vertical wall starts at {pos.X, pos.Y} and ends at {pos.X, pos.Y + 1}.
*/
type Wall struct {
	Direction Direction `json:"direction"`
	Pos1      *Position `json:"position_1"`
	Pos2      *Position `json:"position_2"`
}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
