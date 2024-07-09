package game

import (
	"testing"
)

func TestGetGameById(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	retrievedState, err := service.GetGameById(state.GameId)

	if err != nil {
		t.Errorf("GetGameById: expected no error, got %v", err)
	}

	if retrievedState.GameId != state.GameId {
		t.Errorf("GetGameById: expected retrieved game ID to match, got %s, want %s", retrievedState.GameId, state.GameId)
	}
}

func TestGetGameById_givenNonExistentGameId_shouldReturnError(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	nonExistentGameId := "non-existent-game-id"
	retrievedState, err := service.GetGameById(nonExistentGameId)

	if err == nil {
		t.Errorf("GetGameById: expected error, got none")
	}

	if retrievedState != nil {
		t.Errorf("GetGameById: expected retrieved state to be nil")
	}
}

func TestGetPendingGames(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	service.CreateGame("player1")
	service.CreateGame("player2")

	pendingGames := service.GetPendingGames()
	if len(pendingGames) != 2 {
		t.Errorf("GetPendingGames: expected 2 pending games, got %d", len(pendingGames))
	}
}

func TestGetPendingGames_givenNoPendingGames_shouldReturnEmptyList(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	pendingGames := service.GetPendingGames()
	if len(pendingGames) != 0 {
		t.Errorf("GetPendingGames: expected 0 pending games, got %d", len(pendingGames))
	}
}

func TestGetPendingGames_givenGameJoined_shouldReturnOnePendingGame(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	service.CreateGame("player1")
	service.CreateGame("player2")

	pendingGames := service.GetPendingGames()
	_, err := service.JoinGame(pendingGames[0].GameId, "player3")
	if err != nil {
		t.Errorf("JoinGame: expected no error, got %v", err)
	}

	pendingGames = service.GetPendingGames()
	if len(pendingGames) != 1 {
		t.Errorf("GetPendingGames: expected 1 pending game, got %d", len(pendingGames))
	}
}

func TestCreateGame(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")

	if state.Player1.UserId != "player1" {
		t.Errorf("CreateGame: player1 ID not set correctly")
	}

	if state.GameStatus != GameStatusPending {
		t.Errorf("CreateGame: game status should be pending")
	}

	if state.Player1.Position.X != 4 || state.Player1.Position.Y != 0 {
		t.Errorf("CreateGame: player1 initial position not set correctly")
	}

	if state.Player1.Walls != 10 {
		t.Errorf("CreateGame: initial wall count not set correctly")
	}
}

func TestJoinGame(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	joinedState, err := service.JoinGame(state.GameId, "player2")

	if err != nil {
		t.Errorf("JoinGame: expected no error, got %v", err)
	}

	if joinedState.Player2.UserId != "player2" {
		t.Errorf("JoinGame: player2 ID not set correctly")
	}

	if joinedState.GameStatus != GameStatusInProgress {
		t.Errorf("JoinGame: game status should be in progress")
	}
}

func TestJoinGame_givenNonExistentGameId_shouldReturnError(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	_, err := service.JoinGame("non-existent-game-id", "player2")
	if err == nil {
		t.Errorf("JoinGame: expected error due to non-existent game, got none")
	}
}

func TestJoinGame_givenGameAlreadyStarted_shouldReturnError(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2") // Start the game
	_, err := service.JoinGame(state.GameId, "player3")
	if err == nil {
		t.Errorf("JoinGame: expected error due to game already started, got none")
	}
}

func TestMakeMove(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2")

	validMove := &Position{X: 4, Y: 1}
	updatedState, err := service.MakeMove(state.GameId, "player1", validMove)
	if err != nil {
		t.Errorf("MakeMove: expected no error, got %v", err)
	}

	if updatedState.Player1.Position.X != 4 || updatedState.Player1.Position.Y != 1 {
		t.Errorf("MakeMove: player1 position not updated correctly")
	}

	if updatedState.Turn != "player2" {
		t.Errorf("MakeMove: turn should switch to player2")
	}
}

func TestMakeMove_givenInvalidMove_shouldReturnError(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2")

	invalidMove := &Position{X: 5, Y: 1}
	_, err := service.MakeMove(state.GameId, "player1", invalidMove)
	if err == nil {
		t.Errorf("MakeMove: expected error due to invalid move, got none")
	}
}

func TestMakeMove_givenNotPlayersTurn_shouldReturnError(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2")

	validMove := &Position{X: 4, Y: 1}
	_, err := service.MakeMove(state.GameId, "player2", validMove)
	if err == nil {
		t.Errorf("MakeMove: expected error due to not player's turn, got none")
	}
}

func TestMakeMove_givenWinningMove_shouldCompleteGame(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2")

	state.Player1.Position = &Position{4, 7}

	winningMove := &Position{X: 4, Y: 8}
	updatedState, err := service.MakeMove(state.GameId, "player1", winningMove)
	if err != nil {
		t.Errorf("MakeMove: expected no error, got %v", err)
	}

	if updatedState.GameStatus != GameStatusCompleted {
		t.Errorf("MakeMove: game status should be completed after winning move")
	}

	if updatedState.Winner != "player1" {
		t.Errorf("MakeMove: expected winner to be player1, got %s", updatedState.Winner)
	}
}

func TestPlaceWall(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2")

	validWall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 2, Y: 2},
		Pos2:      &Position{X: 3, Y: 2},
	}

	updatedState, err := service.PlaceWall(state.GameId, "player1", validWall)
	if err != nil {
		t.Errorf("PlaceWall: expected no error, got %v", err)
	}

	if len(updatedState.Walls) != 1 {
		t.Errorf("PlaceWall: wall count should be 1")
	}

	if updatedState.Player1.Walls != 9 {
		t.Errorf("PlaceWall: player1 wall count should be decremented")
	}
}

func TestPlaceWall_givenInvalidPlacement_shouldReturnError(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2")

	invalidWall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 8, Y: 8}, // Out of bounds
		Pos2:      &Position{X: 9, Y: 8},
	}

	_, err := service.PlaceWall(state.GameId, "player1", invalidWall)
	if err == nil {
		t.Errorf("PlaceWall: expected error due to invalid wall placement, got none")
	}
}

func TestPlaceWall_givenNotPlayersTurn_shouldReturnError(t *testing.T) {
	engine := NewGameEngine()
	service := NewGameService(engine)

	state := service.CreateGame("player1")
	service.JoinGame(state.GameId, "player2")

	validWall := &Wall{
		Direction: Horizontal,
		Pos1:      &Position{X: 2, Y: 2},
			Pos2:      &Position{X: 3, Y: 2},
	}

	_, err := service.PlaceWall(state.GameId, "player2", validWall)
	if err == nil {
		t.Errorf("PlaceWall: expected error due to not player's turn, got none")
	}
}
