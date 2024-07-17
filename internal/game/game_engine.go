package game

const BOARD_SIZE = 9

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type GameEngine interface {
	CheckWin(state *Game, player *Player) bool
	IsMoveValid(state *Game, playerId string, newPosition *Position) bool
	IsWallPlacementValid(state *Game, wall *Wall) bool
}

type GameEngineImpl struct {
}

func NewGameEngine() *GameEngineImpl {
	return &GameEngineImpl{}
}

func (engine *GameEngineImpl) CheckWin(state *Game, player *Player) bool {
	return player.Position.Y == player.Goal
}

func (engine *GameEngineImpl) IsMoveValid(state *Game, playerId string, newPosition *Position) bool {
	player := engine.getPlayer(state, playerId)
	opponent := engine.getOpponent(state, playerId)

	if engine.positionsEqual(player.Position, opponent.Position) {
		return false
	}

	if !engine.isWithinBounds(newPosition) {
		return false
	}

	if engine.isJumpOverOpponent(state, player, opponent, newPosition) {
		return true
	}

	if !engine.isAdjacent(player.Position, newPosition) {
		return false
	}

	if engine.crossesWall(state, player.Position, newPosition) {
		return false
	}

	return true
}

func (engine *GameEngineImpl) IsWallPlacementValid(state *Game, wall *Wall) bool {
	if engine.positionsEqual(wall.Pos1, wall.Pos2) {
		return false
	}

	if !engine.isWallWithinBounds(wall) {
		return false
	}

	if engine.wallsOverlap(state, wall) {
		return false
	}

	// verify that both players have a valid path to their goals
	state.Walls = append(state.Walls, wall)
	defer func() {
		state.Walls = state.Walls[:len(state.Walls)-1]
	}()

	if !engine.hasPathToGoal(state, state.Player1) || !engine.hasPathToGoal(state, state.Player2) {
		return false
	}

	return true
}

func (engine *GameEngineImpl) getPlayer(state *Game, playerId string) *Player {
	if state.Player1.UserId == playerId {
		return state.Player1
	}

	return state.Player2
}

func (engine *GameEngineImpl) getOpponent(state *Game, playerId string) *Player {
	if state.Player1.UserId == playerId {
		return state.Player2
	}

	return state.Player1
}

func (engine *GameEngineImpl) getPlayerDirection(pos1, pos2 *Position) Direction {
	if pos1.X-pos2.X == 0 {
		return Vertical
	}

	return Horizontal
}

func (engine *GameEngineImpl) isWithinBounds(pos *Position) bool {
	return pos.X >= 0 && pos.X < BOARD_SIZE && pos.Y >= 0 && pos.Y < BOARD_SIZE
}

func (engine *GameEngineImpl) isAdjacent(pos1, pos2 *Position) bool {
	return (abs(pos1.X-pos2.X) == 1 && pos1.Y == pos2.Y) || (abs(pos1.Y-pos2.Y) == 1 && pos1.X == pos2.X)
}

func (engine *GameEngineImpl) crossesWall(state *Game, pos1, pos2 *Position) bool {
	playerDirection := engine.getPlayerDirection(pos1, pos2)

	for _, wall := range state.Walls {
		if wall.Direction == playerDirection {
			continue
		}

		if playerDirection == Horizontal &&
			engine.crossesVerticalWall(wall, pos1, pos2) &&
			engine.crossesVerticalWallSpan(wall, pos1) {
			return true
		}

		if playerDirection == Vertical &&
			engine.crossesHorizontalWall(wall, pos1, pos2) &&
			engine.crossesHorizontalWallSpan(wall, pos1) {
			return true
		}
	}

	return false
}

/*
jump over opponent is alloed if:
  - opponent is adjacent to the player
  - target position is two cells away from the player's position in the same direction
  - no walls between the player and the opponent or between the opponent and the target position
*/
func (engine *GameEngineImpl) isJumpOverOpponent(state *Game, player, opponent *Player, newPosition *Position) bool {
	// check if player, opponent and new position are adjacent, and if there is connection between player and opponent
	if !engine.isAdjacent(player.Position, opponent.Position) ||
		engine.crossesWall(state, player.Position, opponent.Position) {
		return false
	}

	// if there is not wall behind the opponent, then it is the only valid jump target
	behindOpponent := &Position{
		X: opponent.Position.X + (opponent.Position.X - player.Position.X),
		Y: opponent.Position.Y + (opponent.Position.Y - player.Position.Y),
	}
	if !engine.crossesWall(state, opponent.Position, behindOpponent) {
		return engine.positionsEqual(newPosition, behindOpponent)
	}

	// if there is a wall between the opponent and the position behind, then we need to check opponent adjacent positions
	adjacentPositions := engine.getAdjacentPositions(opponent.Position)
	for _, position := range adjacentPositions {
		if engine.isWithinBounds(opponent.Position) &&
			engine.positionsEqual(newPosition, position) &&
			!engine.positionsEqual(player.Position, position) &&
			!engine.crossesWall(state, opponent.Position, position) {
			return true
		}
	}

	return false
}

func (engine *GameEngineImpl) positionsEqual(position1, position2 *Position) bool {
	return position1.X == position2.X && position1.Y == position2.Y
}

func (engine *GameEngineImpl) getAdjacentPositions(position *Position) []*Position {
	return []*Position{
		{X: position.X - 1, Y: position.Y},
		{X: position.X + 1, Y: position.Y},
		{X: position.X, Y: position.Y - 1},
		{X: position.X, Y: position.Y + 1},
	}
}

func (engine *GameEngineImpl) crossesHorizontalWall(wall *Wall, pos1, pos2 *Position) bool {
	return (wall.Pos1.Y == pos1.Y && wall.Pos2.Y == pos2.Y) || (wall.Pos1.Y == pos2.Y && wall.Pos2.Y == pos1.Y)
}

func (engine *GameEngineImpl) crossesHorizontalWallSpan(wall *Wall, pos1 *Position) bool {
	return wall.Pos1.X == pos1.X || wall.Pos1.X+1 == pos1.X
}

func (engine *GameEngineImpl) crossesVerticalWall(wall *Wall, pos1, pos2 *Position) bool {
	return (wall.Pos1.X == pos1.X && wall.Pos2.X == pos2.X) || (wall.Pos1.X == pos2.X && wall.Pos2.X == pos1.X)
}

func (engine *GameEngineImpl) crossesVerticalWallSpan(wall *Wall, pos1 *Position) bool {
	return wall.Pos1.Y == pos1.Y || wall.Pos1.Y+1 == pos1.Y
}

// includes validation of the spann of the wall(+1 to direction coordinate)
func (engine *GameEngineImpl) isWallWithinBounds(wall *Wall) bool {
	deltaX, deltaY := 0, 0

	if wall.Direction == Horizontal {
		deltaX = 1
	} else if wall.Direction == Vertical {
		deltaY = 1
	} else {
		return false
	}

	return engine.isWithinBounds(wall.Pos1) &&
		engine.isWithinBounds(wall.Pos2) &&
		engine.isWithinBounds(&Position{wall.Pos1.X + deltaX, wall.Pos1.Y + deltaY}) &&
		engine.isWithinBounds(&Position{wall.Pos2.X + deltaX, wall.Pos2.Y + deltaY})
}

func (engine *GameEngineImpl) wallsOverlap(state *Game, wall *Wall) bool {
	for _, existingWall := range state.Walls {
		if wall.Direction != existingWall.Direction {
			continue
		}

		if wall.Direction == Horizontal && engine.wallsOverlapHorizontally(wall, existingWall) {
			return true
		}

		if wall.Direction == Vertical && engine.wallsOverlapVertically(wall, existingWall) {
			return true
		}
	}

	return false
}

/*
Walls overlap horizontally if:
  - both have horizontal directoin
  - on the same horizontal line(y coordinates are equal)
  - horizontal lines overlap
*/
func (engine *GameEngineImpl) wallsOverlapHorizontally(wall1 *Wall, wall2 *Wall) bool {
	// horizontal direction
	if !(wall1.Direction == wall2.Direction && wall1.Direction == Horizontal) {
		return false
	}

	// on the same horizontal line
	if !(wall1.Pos1.Y == wall2.Pos1.Y && wall1.Pos2.Y == wall2.Pos2.Y || wall1.Pos2.Y == wall2.Pos1.Y && wall1.Pos1.Y == wall2.Pos2.Y) {
		return false
	}

	// horizontal lines overlap
	if !(wall1.Pos1.X == wall2.Pos1.X || wall1.Pos1.X+1 == wall2.Pos1.X || wall2.Pos1.X+1 == wall1.Pos1.X) {
		return false
	}

	return true
}

/*
Walls overlap vertically if:
  - both have vertical directoin
  - on the same vertical line(x coordinates are equal)
  - vertical lines overlap
*/
func (engine *GameEngineImpl) wallsOverlapVertically(wall1 *Wall, wall2 *Wall) bool {
	// vertical direction
	if !(wall1.Direction == wall2.Direction && wall1.Direction == Vertical) {
		return false
	}

	// on the same vertical line
	if !(wall1.Pos1.X == wall2.Pos1.X && wall1.Pos2.X == wall2.Pos2.X || wall1.Pos2.X == wall2.Pos1.X && wall1.Pos1.X == wall2.Pos2.X) {
		return false
	}

	// vertical lines overlap
	if !(wall1.Pos1.Y == wall2.Pos1.Y || wall1.Pos1.Y+1 == wall2.Pos1.Y || wall2.Pos1.Y+1 == wall1.Pos1.Y) {
		return false
	}

	return true
}

/*
checks if there is a valid path for the player to reach their goal
uses a breadth-first search algorithm (BFS)
*/
func (engine *GameEngineImpl) hasPathToGoal(state *Game, player *Player) bool {
	visited := make(map[Position]bool)
	queue := []*Position{player.Position}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current.Y == player.Goal {
			return true
		}

		for _, neighbor := range engine.getAdjacentPositions(current) {
			if !engine.isWithinBounds(neighbor) || visited[*neighbor] || engine.crossesWall(state, current, neighbor) {
				continue
			}

			visited[*neighbor] = true
			queue = append(queue, neighbor)
		}
	}

	return false
}
