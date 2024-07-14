package game

import (
	"testing"
	"time"

	"quoridor/internal/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockGameRepository struct {
	mock.Mock
}

func (m *MockGameRepository) SaveGame(state *Game) error {
	args := m.Called(state)
	return args.Error(0)
}

func (m *MockGameRepository) GetGameById(gameID string) (*Game, error) {
	args := m.Called(gameID)
	return args.Get(0).(*Game), args.Error(1)
}

func (m *MockGameRepository) GetGamesByStatus(status GameStatus) ([]*Game, error) {
	args := m.Called(status)
	return args.Get(0).([]*Game), args.Error(1)
}

func TestGetGameById(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusPending,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 0},
			Goal:     8,
			Walls:    10,
		},
		Walls: []*Wall{},
	}
	repo.On("GetGameById", "test-game-id").Return(state, nil)

	retrievedState, err := service.GetGameById("test-game-id")

	assert.NoError(t, err)
	assert.Equal(t, state.GameId, retrievedState.GameId)
	repo.AssertCalled(t, "GetGameById", "test-game-id")
}

func TestGetGameById_givenNonExistentGameId_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	repo.On("GetGameById", "non-existent-game-id").Return((*Game)(nil), errors.ErrInternalError)

	retrievedState, err := service.GetGameById("non-existent-game-id")

	assert.ErrorIs(t, err, errors.ErrInternalError)
	assert.Nil(t, retrievedState)
	repo.AssertCalled(t, "GetGameById", "non-existent-game-id")
}

func TestGetPendingGames(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	game1 := &Game{
		GameId:     "game1",
		GameStatus: GameStatusPending,
	}
	game2 := &Game{
		GameId:     "game2",
		GameStatus: GameStatusPending,
	}

	repo.On("GetGamesByStatus", GameStatusPending).Return([]*Game{game1, game2}, nil)

	pendingGames, err := service.GetPendingGames()
	assert.NoError(t, err)
	assert.Len(t, pendingGames, 2)
	repo.AssertCalled(t, "GetGamesByStatus", GameStatusPending)
}

func TestGetPendingGames_givenNoPendingGames_shouldReturnEmptyList(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	repo.On("GetGamesByStatus", GameStatusPending).Return([]*Game{}, nil)

	pendingGames, err := service.GetPendingGames()
	assert.NoError(t, err)
	assert.Empty(t, pendingGames)
	repo.AssertCalled(t, "GetGamesByStatus", GameStatusPending)
}

func TestCreateGame(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	repo.On("SaveGame", mock.Anything).Return(nil)

	state, err := service.CreateGame("player1", "player2")
	assert.NoError(t, err)
	assert.Equal(t, "player1", state.Player1.UserId)
	assert.Equal(t, "player2", state.Player2.UserId)
	assert.Equal(t, GameStatusInProgress, state.GameStatus)
	assert.Equal(t, &Position{X: 4, Y: 0}, state.Player1.Position)
	assert.Equal(t, &Position{X: 4, Y: 8}, state.Player2.Position)
	assert.Equal(t, 10, state.Player1.Walls)
	assert.Equal(t, 10, state.Player2.Walls)
	assert.NotEmpty(t, state.CreatedAt)
	assert.NotEmpty(t, state.UpdatedAt)

	repo.AssertCalled(t, "SaveGame", mock.Anything)
}

func TestMakeMove(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)
	repo.On("SaveGame", mock.Anything).Return(nil)

	validMove := &Position{X: 4, Y: 5}
	updatedState, err := service.MakeMove("test-game-id", "player1", validMove)
	assert.NoError(t, err)
	assert.Equal(t, validMove, updatedState.Player1.Position)
	assert.Equal(t, "player2", updatedState.Turn)
	assert.Len(t, updatedState.Moves, 1)
	assert.Equal(t, validMove, updatedState.Moves[0].Position)
	assert.Equal(t, MoveTypeMove, updatedState.Moves[0].Type)
	assert.NotEmpty(t, updatedState.UpdatedAt)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertCalled(t, "SaveGame", mock.Anything)
}

func TestMakeMove_givenGameNotFound_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	repo.On("GetGameById", "test-game-id").Return((*Game)(nil), errors.ErrInternalError)

	newPos := &Position{
		X: 2, Y: 2,
	}

	_, err := service.MakeMove("test-game-id", "player2", newPos)
	assert.ErrorIs(t, err, errors.ErrInternalError)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}

func TestMakeMove_givenGameNotInProgress_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusCompleted,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)

	newPos := &Position{
		X: 2, Y: 2,
	}

	_, err := service.MakeMove("test-game-id", "player2", newPos)
	assert.ErrorIs(t, err, errors.ErrGameNotInProgress)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}

func TestMakeMove_givenInvalidMove_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)

	invalidMove := &Position{X: 5, Y: 5}
	_, err := service.MakeMove("test-game-id", "player1", invalidMove)
	assert.ErrorIs(t, err, errors.ErrInvalidMove)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}

func TestMakeMove_givenNotPlayersTurn_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)

	validMove := &Position{X: 4, Y: 5}
	_, err := service.MakeMove("test-game-id", "player2", validMove)
	assert.ErrorIs(t, err, errors.ErrNotPlayersTurn)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}

func TestMakeMove_givenWinningMove_shouldCompleteGame(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 7},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 0},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)
	repo.On("SaveGame", mock.Anything).Return(nil)

	winningMove := &Position{X: 4, Y: 8}
	updatedState, err := service.MakeMove("test-game-id", "player1", winningMove)
	assert.NoError(t, err)
	assert.Equal(t, GameStatusCompleted, updatedState.GameStatus)
	assert.Equal(t, "player1", updatedState.Winner)
	assert.NotNil(t, updatedState.CompletedAt)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertCalled(t, "SaveGame", mock.Anything)
}

func TestPlaceWall(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)
	repo.On("SaveGame", mock.Anything).Return(nil)

	validWall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 2, Y: 2},
		Pos2:      &Position{X: 3, Y: 2},
	}

	updatedState, err := service.PlaceWall("test-game-id", "player1", validWall)
	assert.NoError(t, err)
	assert.Len(t, updatedState.Walls, 1)
	assert.Equal(t, validWall, updatedState.Walls[0])
	assert.Equal(t, 9, updatedState.Player1.Walls)
	assert.Len(t, updatedState.Moves, 1)
	assert.Equal(t, validWall, updatedState.Moves[0].Wall)
	assert.Equal(t, MoveTypePlaceWall, updatedState.Moves[0].Type)
	assert.WithinDuration(t, time.Now(), updatedState.UpdatedAt, time.Second)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertCalled(t, "SaveGame", mock.Anything)
}

func TestPlaceWall_givenGameNotFound_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	repo.On("GetGameById", "test-game-id").Return((*Game)(nil), errors.ErrInternalError)

	wall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 2, Y: 2},
		Pos2:      &Position{X: 3, Y: 2},
	}

	_, err := service.PlaceWall("test-game-id", "player2", wall)
	assert.ErrorIs(t, err, errors.ErrInternalError)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}

func TestPlaceWall_givenGameNotInProgress_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusCompleted,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)

	wall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 2, Y: 2},
		Pos2:      &Position{X: 3, Y: 2},
	}

	_, err := service.PlaceWall("test-game-id", "player2", wall)
	assert.ErrorIs(t, err, errors.ErrGameNotInProgress)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}

func TestPlaceWall_givenInvalidPlacement_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)

	invalidWall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 8, Y: 8}, // Out of bounds
		Pos2:      &Position{X: 9, Y: 8},
	}

	_, err := service.PlaceWall("test-game-id", "player1", invalidWall)
	assert.ErrorIs(t, err, errors.ErrInvalidWallPlacement)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}

func TestPlaceWall_givenNotPlayersTurn_shouldReturnError(t *testing.T) {
	repo := new(MockGameRepository)
	engine := NewGameEngine()
	service := NewGameService(engine, repo)

	state := &Game{
		GameId:     "test-game-id",
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   "player1",
			Position: &Position{X: 4, Y: 4},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   "player2",
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:  "player1",
		Walls: []*Wall{},
	}

	repo.On("GetGameById", "test-game-id").Return(state, nil)

	validWall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 2, Y: 2},
		Pos2:      &Position{X: 3, Y: 2},
	}

	_, err := service.PlaceWall("test-game-id", "player2", validWall)
	assert.ErrorIs(t, err, errors.ErrNotPlayersTurn)

	repo.AssertCalled(t, "GetGameById", "test-game-id")
	repo.AssertNotCalled(t, "SaveGame", mock.Anything)
}
