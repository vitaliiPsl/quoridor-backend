package game

import "testing"

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

	state := &GameState{
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

	state := &GameState{
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
