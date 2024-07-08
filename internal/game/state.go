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
	GameId     string
	GameStatus GameStatus
	Turn       string // id of the player
	Player1    *Player
	Player2    *Player
	Walls      []*Wall
}

type Player struct {
	UserId   string
	Position *Position
	Goal     int // raw user need to get to to win the game
	Walls    int // number of walls avaiable
}

/*
i'm using two positions of the cells that the wall is between to simplify checks.
each wall spans two cells: 
- a horizontal wall starts at {pos.X, pos.Y} and ends at {pos.X + 1, pos.Y}.
- a vertical wall starts at {pos.X, pos.Y} and ends at {pos.X, pos.Y + 1}.
*/
type Wall struct {
	Direction Direction
	Pos1      *Position
	Pos2      *Position
}

type Position struct {
	X int
	Y int
}
