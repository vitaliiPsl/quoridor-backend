package game

import "testing"

func TestCheckWin(t *testing.T) {
	engine := NewGameEngine()
	player1 := &Player{
		UserId:   "player1",
		Position: &Position{X: 4, Y: 0},
		Goal:     8,
	}
	player2 := &Player{
		UserId:   "player2",
		Position: &Position{X: 4, Y: 8},
		Goal:     0,
	}

	state := &Game{
		Player1: player1,
		Player2: player2,
	}

	if engine.CheckWin(state, player1) {
		t.Errorf("CheckWin: expected player1 not to win")
	}

	player1.Position.Y = 8
	if !engine.CheckWin(state, player1) {
		t.Errorf("CheckWin: expected player1 to win")
	}
}

func TestIsWithinBounds(t *testing.T) {
	engine := NewGameEngine()

	tests := []struct {
		position *Position
		expected bool
	}{
		{&Position{X: 0, Y: 0}, true},
		{&Position{X: 8, Y: 8}, true},
		{&Position{X: 9, Y: 9}, false},
		{&Position{X: -1, Y: 0}, false},
	}

	for _, test := range tests {
		result := engine.isWithinBounds(test.position)
		if result != test.expected {
			t.Errorf("isWithinBounds(%v) = %v; want %v", test.position, result, test.expected)
		}
	}
}

func TestIsAdjacent(t *testing.T) {
	engine := NewGameEngine()

	tests := []struct {
		pos1     *Position
		pos2     *Position
		expected bool
	}{
		{&Position{X: 4, Y: 4}, &Position{X: 4, Y: 5}, true},
		{&Position{X: 4, Y: 4}, &Position{X: 5, Y: 4}, true},
		{&Position{X: 4, Y: 4}, &Position{X: 5, Y: 5}, false},
		{&Position{X: 4, Y: 4}, &Position{X: 4, Y: 4}, false},
	}

	for _, test := range tests {
		result := engine.isAdjacent(test.pos1, test.pos2)
		if result != test.expected {
			t.Errorf("isAdjacent(%v, %v) = %v; want %v", test.pos1, test.pos2, result, test.expected)
		}
	}
}

func TestCrossesWall(t *testing.T) {
	engine := NewGameEngine()

	state := &Game{
		Walls: []*Wall{
			{Direction: Vertical, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 3, Y: 2}},
			{Direction: Horizontal, Pos1: &Position{X: 4, Y: 4}, Pos2: &Position{X: 4, Y: 5}},
		},
	}

	tests := []struct {
		pos1     *Position
		pos2     *Position
		expected bool
	}{
		{&Position{X: 2, Y: 1}, &Position{X: 2, Y: 2}, false},
		{&Position{X: 2, Y: 5}, &Position{X: 3, Y: 5}, false},
		{&Position{X: 7, Y: 5}, &Position{X: 7, Y: 4}, false},
		{&Position{X: 2, Y: 2}, &Position{X: 3, Y: 2}, true}, // vertical wall at (2, 2) | (3, 2)
		{&Position{X: 3, Y: 3}, &Position{X: 2, Y: 3}, true}, // vertical wall at (2, 2) | (3, 2)
		{&Position{X: 5, Y: 4}, &Position{X: 5, Y: 5}, true}, // horizontal wall at (4, 4) | (4, 5)
		{&Position{X: 4, Y: 5}, &Position{X: 4, Y: 4}, true}, // horizontal wall at (4, 4) | (4, 5)
	}

	for _, test := range tests {
		result := engine.crossesWall(state, test.pos1, test.pos2)
		if result != test.expected {
			t.Errorf("crossesWall(state, %v, %v) = %v; want %v", test.pos1, test.pos2, result, test.expected)
		}
	}
}

func TestIsMoveValid(t *testing.T) {
	engine := NewGameEngine()

	state := &Game{
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 5},
		},
		Walls: []*Wall{
			{Direction: Horizontal, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 3, Y: 2}},
			{Direction: Vertical, Pos1: &Position{X: 5, Y: 5}, Pos2: &Position{X: 5, Y: 6}},
		},
	}

	tests := []struct {
		playerId string
		newPos   *Position
		expected bool
	}{
		// out of bounds
		{"player1", &Position{X: 9, Y: 4}, false},
		{"player1", &Position{X: -1, Y: 4}, false},

		// non-adjacent
		{"player1", &Position{X: 6, Y: 4}, false},
		{"player1", &Position{X: 4, Y: 6}, false},

		// cross walls
		{"player1", &Position{X: 2, Y: 2}, false},
		{"player1", &Position{X: 3, Y: 2}, false},
		{"player2", &Position{X: 5, Y: 4}, false},
		{"player2", &Position{X: 5, Y: 6}, false},

		// valid
		{"player1", &Position{X: 4, Y: 3}, true},
		{"player1", &Position{X: 3, Y: 4}, true},
		{"player2", &Position{X: 4, Y: 6}, true},
		{"player2", &Position{X: 3, Y: 5}, true},
	}

	for _, test := range tests {
		result := engine.IsMoveValid(state, test.playerId, test.newPos)
		if result != test.expected {
			t.Errorf("IsMoveValid(%v, %v, %v) = %v; want %v", state, test.playerId, test.newPos, result, test.expected)
		}
	}
}

func TestOverlapsWall(t *testing.T) {
	engine := NewGameEngine()

	state := &Game{
		Walls: []*Wall{
			{Direction: Horizontal, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 2, Y: 3}},
			{Direction: Vertical, Pos1: &Position{X: 4, Y: 4}, Pos2: &Position{X: 5, Y: 4}},
		},
	}

	tests := []struct {
		wall     *Wall
		expected bool
	}{
		// overlapping
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 2, Y: 3}}, true},
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 3, Y: 2}, Pos2: &Position{X: 3, Y: 3}}, true},
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 1, Y: 2}, Pos2: &Position{X: 1, Y: 3}}, true},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 4, Y: 4}, Pos2: &Position{X: 5, Y: 4}}, true},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 4, Y: 5}, Pos2: &Position{X: 5, Y: 5}}, true},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 4, Y: 3}, Pos2: &Position{X: 5, Y: 3}}, true},

		// non-overlapping
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 0, Y: 2}, Pos2: &Position{X: 0, Y: 3}}, false},
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 4, Y: 2}, Pos2: &Position{X: 4, Y: 3}}, false},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 4, Y: 6}, Pos2: &Position{X: 5, Y: 6}}, false},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 4, Y: 2}, Pos2: &Position{X: 5, Y: 2}}, false},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 3, Y: 2}}, false},
	}

	for _, test := range tests {
		result := engine.overlapsWall(state, test.wall)
		if result != test.expected {
			t.Errorf("overlapsWall(state, {%v %v %v}) = %v; want %+v", test.wall.Direction, *test.wall.Pos1, *test.wall.Pos2, result, test.expected)
		}
	}
}

func TestHasPathToGoal(t *testing.T) {
	engine := NewGameEngine()

	state := &Game{
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 0},
			Goal:     8,
		},
		Walls: []*Wall{
			{Direction: Vertical, Pos1: &Position{X: 3, Y: 0}, Pos2: &Position{X: 4, Y: 0}},
			{Direction: Vertical, Pos1: &Position{X: 4, Y: 0}, Pos2: &Position{X: 5, Y: 0}},
			{Direction: Vertical, Pos1: &Position{X: 3, Y: 1}, Pos2: &Position{X: 4, Y: 1}},
			{Direction: Vertical, Pos1: &Position{X: 4, Y: 1}, Pos2: &Position{X: 5, Y: 1}},

			{Direction: Vertical, Pos1: &Position{X: 3, Y: 8}, Pos2: &Position{X: 4, Y: 8}},
			{Direction: Vertical, Pos1: &Position{X: 4, Y: 8}, Pos2: &Position{X: 5, Y: 8}},
			{Direction: Vertical, Pos1: &Position{X: 3, Y: 7}, Pos2: &Position{X: 4, Y: 7}},
			{Direction: Vertical, Pos1: &Position{X: 4, Y: 7}, Pos2: &Position{X: 5, Y: 7}},
			{Direction: Horizontal, Pos1: &Position{X: 4, Y: 6}, Pos2: &Position{X: 4, Y: 7}},
		},
	}

	tests := []struct {
		player   *Player
		expected bool
	}{
		{state.Player1, false},
		{state.Player2, true},
	}

	for _, test := range tests {
		result := engine.hasPathToGoal(state, test.player)
		if result != test.expected {
			t.Errorf("hasPathToGoal(state, %v) = %v; want %v", test.player.UserId, result, test.expected)
		}
	}
}

func TestIsWallPlacementValid(t *testing.T) {
	engine := NewGameEngine()

	state := &Game{
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 0},
			Goal:     8,
		},
		Walls: []*Wall{
			{Direction: Horizontal, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 2, Y: 3}},
			{Direction: Vertical, Pos1: &Position{X: 4, Y: 4}, Pos2: &Position{X: 5, Y: 4}},
		},
	}

	tests := []struct {
		wall     *Wall
		expected bool
	}{
		// valid
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 0, Y: 2}, Pos2: &Position{X: 0, Y: 3}}, true},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 4, Y: 6}, Pos2: &Position{X: 5, Y: 6}}, true},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 3, Y: 2}}, true},

		// out of bounds
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 8, Y: 8}, Pos2: &Position{X: 9, Y: 8}}, false},
		{&Wall{Direction: Vertical, Pos1: &Position{X: -1, Y: 0}, Pos2: &Position{X: -1, Y: 1}}, false},
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 8, Y: 8}, Pos2: &Position{X: 8, Y: 8}}, false},

		// overlapping
		{&Wall{Direction: Horizontal, Pos1: &Position{X: 2, Y: 2}, Pos2: &Position{X: 2, Y: 3}}, false},
		{&Wall{Direction: Vertical, Pos1: &Position{X: 4, Y: 4}, Pos2: &Position{X: 5, Y: 4}}, false},
	}

	for _, test := range tests {
		result := engine.IsWallPlacementValid(state, test.wall)
		if result != test.expected {
			t.Errorf("IsWallPlacementValid(state, %v) = %v; want %v", test.wall, result, test.expected)
		}
	}
}
