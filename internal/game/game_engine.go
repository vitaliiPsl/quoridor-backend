package game

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type GameEngine interface {
	IsMoveValid(state *GameState, playerId string, newPos *Position) bool
	IsWallPlacementValid(state *GameState, wall *Wall) bool

	isWithinBounds(pos *Position) bool
	isAdjacent(pos1, pos2 *Position) bool
	crossesWall(state *GameState, pos1, pos2 *Position) bool
	overlapsWall(state *GameState, wall *Wall) bool
	hasPathToGoal(state *GameState, player *Player) bool
}

type GameEngineImpl struct {
}

func NewGameEngine() GameEngine {
	return &GameEngineImpl{}
}

func (engine *GameEngineImpl) IsMoveValid(state *GameState, playerId string, newPos *Position) bool {
	player := engine.getPlayer(state, playerId)
	if player == nil {
		return false
	}

	if !engine.isWithinBounds(newPos) {
		return false
	}

	if !engine.isAdjacent(player.Position, newPos) {
		return false
	}

	if engine.crossesWall(state, player.Position, newPos) {
		return false
	}

	return true
}

func (engine *GameEngineImpl) IsWallPlacementValid(state *GameState, wall *Wall) bool {
	if wall.Pos1.X == wall.Pos2.X && wall.Pos1.Y == wall.Pos2.Y {
		return false
	}

	if !engine.isWithinBounds(wall.Pos1) || !engine.isWithinBounds(wall.Pos2) {
		return false
	}

	if engine.overlapsWall(state, wall) {
		return false
	}

	state.Walls = append(state.Walls, wall)
	defer func() { state.Walls = state.Walls[:len(state.Walls)-1] }()

	if !engine.hasPathToGoal(state, state.Player1) || !engine.hasPathToGoal(state, state.Player2) {
		return false
	}

	return true
}

func (engine *GameEngineImpl) getPlayer(state *GameState, playerId string) *Player {
	if state.Player1.UserId == playerId {
		return state.Player1
	} else if state.Player2.UserId == playerId {
		return state.Player2
	}

	return nil
}

func (engine *GameEngineImpl) getPlayerDirection(pos1, pos2 *Position) Direction {
	if pos1.X-pos2.X == 0 {
		return Vertical
	}
	return Horizontal
}

func (engine *GameEngineImpl) isWithinBounds(pos *Position) bool {
	return pos.X >= 0 && pos.X < 9 && pos.Y >= 0 && pos.Y < 9
}

func (engine *GameEngineImpl) isAdjacent(pos1, pos2 *Position) bool {
	return (abs(pos1.X-pos2.X) == 1 && pos1.Y == pos2.Y) || (abs(pos1.Y-pos2.Y) == 1 && pos1.X == pos2.X)
}

func (engine *GameEngineImpl) crossesWall(state *GameState, pos1, pos2 *Position) bool {
	playerDirection := engine.getPlayerDirection(pos1, pos2)

	for _, wall := range state.Walls {
		if wall.Direction == playerDirection {
			continue
		}

		if playerDirection == Horizontal {
			if engine.crossesVerticalWall(wall, pos1, pos2) && engine.crossesVerticalWallSpan(wall, pos1) {
				return true
			}
		}

		if playerDirection == Vertical {
			if engine.crossesHorizontalWall(wall, pos1, pos2) && engine.crossesHorizontalWallSpan(wall, pos1) {
				return true
			}
		}
	}

	return false
}

func (engine *GameEngineImpl) overlapsWall(state *GameState, wall *Wall) bool {
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

// bfs
func (engine *GameEngineImpl) hasPathToGoal(state *GameState, player *Player) bool {
	visited := make(map[Position]bool)
	queue := []*Position{player.Position}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if engine.reachedGoal(current, player.Goal) {
			return true
		}

		for _, neighbor := range engine.getNeighbors(current) {
			if !engine.isWithinBounds(neighbor) || visited[*neighbor] || engine.crossesWall(state, current, neighbor) {
				continue
			}

			visited[*neighbor] = true
			queue = append(queue, neighbor)
		}
	}

	return false
}

func (engine *GameEngineImpl) reachedGoal(pos *Position, goal int) bool {
	return pos.Y == goal
}

func (engine *GameEngineImpl) getNeighbors(pos *Position) []*Position {
	return []*Position{
		{X: pos.X, Y: pos.Y - 1},
		{X: pos.X, Y: pos.Y + 1},
		{X: pos.X - 1, Y: pos.Y},
		{X: pos.X + 1, Y: pos.Y},
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

func (engine *GameEngineImpl) wallsOverlapHorizontally(wall1 *Wall, wall2 *Wall) bool {
	return wall1.Pos1.X == wall2.Pos1.X || wall1.Pos1.X+1 == wall2.Pos1.X || wall2.Pos1.X+1 == wall1.Pos1.X
}

func (engine *GameEngineImpl) wallsOverlapVertically(wall1 *Wall, wall2 *Wall) bool {
	return wall1.Pos1.Y == wall2.Pos1.Y || wall1.Pos1.Y+1 == wall2.Pos1.Y || wall2.Pos1.Y+1 == wall1.Pos1.Y
}
