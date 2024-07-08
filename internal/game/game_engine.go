package game

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

type GameEngine interface {
	IsMoveValid(state *GameState, playerId string, newPos *Position) bool

	isWithinBounds(pos *Position) bool
	isAdjacent(pos1, pos2 *Position) bool
	crossesWall(state *GameState, pos1, pos2 *Position) bool
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
