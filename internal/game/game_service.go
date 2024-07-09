package game

import (
	"errors"
	"log"
	"sync"

	"github.com/google/uuid"
)

type GameService interface {
	GetGameById(gameId string) (*GameState, error)
	GetPendingGames() []*GameState
	CreateGame(userId string) *GameState
	JoinGame(gameId, userId string) (*GameState, error)
	MakeMove(gameId, userId string, newPos *Position) (*GameState, error)
	PlaceWall(gameId, userId string, wall *Wall) (*GameState, error)
}

type GameServiceImpl struct {
	engine GameEngine
	games  map[string]*GameState
	mu     sync.Mutex
}

func NewGameService(engine GameEngine) GameService {
	return &GameServiceImpl{
		engine: engine,
		games:  make(map[string]*GameState),
	}
}

func (service *GameServiceImpl) GetGameById(gameId string) (*GameState, error) {
	log.Printf("GetGameById: gameId=%v", gameId)

	service.mu.Lock()
	defer service.mu.Unlock()

	gameState, ok := service.games[gameId]
	if !ok {
		log.Printf("GetGameById: game with id=%s not found", gameId)
		return nil, errors.New("game not found")
	}
	return gameState, nil
}

func (service *GameServiceImpl) GetPendingGames() []*GameState {
	log.Println("GetPendingGames")

	service.mu.Lock()
	defer service.mu.Unlock()

	pendingGames := []*GameState{}
	for _, state := range service.games {
		if state.GameStatus == GameStatusPending {
			pendingGames = append(pendingGames, state)
		}
	}

	log.Printf("GetPendingGames: found %d pending games", len(pendingGames))
	return pendingGames
}

func (service *GameServiceImpl) CreateGame(userId string) *GameState {
	log.Printf("CreateGame: userId=%v", userId)

	service.mu.Lock()
	defer service.mu.Unlock()

	gameId := uuid.NewString()
	state := &GameState{
		GameId:     gameId,
		GameStatus: GameStatusPending,
		Player1: &Player{
			UserId:   userId,
			Position: &Position{X: 4, Y: 0},
			Goal:     8,
			Walls:    10,
		},
		Walls: []*Wall{},
	}

	service.games[gameId] = state
	log.Printf("CreateGame: created game with id=%s for user with id=%s", gameId, userId)
	return state
}

func (service *GameServiceImpl) JoinGame(gameId, userId string) (*GameState, error) {
	log.Printf("JoinGame: gameId=%v, userId=%v", gameId, userId)

	service.mu.Lock()
	defer service.mu.Unlock()

	state, exists := service.games[gameId]
	if !exists {
		log.Printf("JoinGame: game with id=%s not found", gameId)
		return nil, errors.New("game not found")
	}

	if state.GameStatus != GameStatusPending {
		log.Printf("JoinGame: game with id=%s already started", gameId)
		return nil, errors.New("game already started")
	}

	state.Player2 = &Player{
		UserId:   userId,
		Position: &Position{X: 4, Y: 8},
		Goal:     0,
		Walls:    10,
	}

	state.GameStatus = GameStatusInProgress
	state.Turn = state.Player1.UserId
	log.Printf("JoinGame: user with id=%s joined game with id=%s", userId, gameId)
	return state, nil
}

func (service *GameServiceImpl) MakeMove(gameId, userId string, newPos *Position) (*GameState, error) {
	log.Printf("MakeMove: gameId=%v, userId=%v, new position=%+v", gameId, userId, *newPos)

	service.mu.Lock()
	defer service.mu.Unlock()

	state, ok := service.games[gameId]
	if !ok {
		log.Printf("MakeMove: game with id=%s not found", gameId)
		return nil, errors.New("game not found")
	}

	if state.GameStatus != GameStatusInProgress {
		log.Printf("MakeMove: game with id=%s is not in progress", gameId)
		return nil, errors.New("game not in progress")
	}

	if state.Turn != userId {
		log.Printf("MakeMove: it is not user with id=%s turn in game with id=%s", userId, gameId)
		return nil, errors.New("not player's turn")
	}

	if !service.engine.IsMoveValid(state, userId, newPos) {
		log.Printf("MakeMove: invalid move by user with id=%s in game with id=%s. Move=%+v", userId, gameId, *newPos)
		return nil, errors.New("invalid move")
	}

	player := service.getPlayer(state, userId)
	player.Position = newPos

	if service.engine.CheckWin(state, player) {
		state.GameStatus = GameStatusCompleted
		state.Winner = userId
		log.Printf("MakeMove: user with id=%s has won the game with id=%s", userId, gameId)
		return state, nil
	}

	state.Turn = service.getNextTurn(state)
	log.Printf("MakeMove: user with id=%s moved to position=%+v in game with id=%s", userId, *newPos, gameId)
	return state, nil
}

func (service *GameServiceImpl) PlaceWall(gameId, userId string, wall *Wall) (*GameState, error) {
	log.Printf("PlaceWall: gameId=%v, userId=%v, wall={%v %+v %+v}", gameId, userId, wall.Direction, *wall.Pos1, *wall.Pos2)

	service.mu.Lock()
	defer service.mu.Unlock()

	state, ok := service.games[gameId]
	if !ok {
		log.Printf("PlaceWall: game with id=%s not found", gameId)
		return nil, errors.New("game not found")
	}

	if state.GameStatus != GameStatusInProgress {
		log.Printf("PlaceWall: game with id=%s is not in progress", gameId)
		return nil, errors.New("game not in progress")
	}

	if state.Turn != userId {
		log.Printf("PlaceWall: it is not user with id=%s turn in game with id=%s", userId, gameId)
		return nil, errors.New("not player's turn")
	}

	if !service.engine.IsWallPlacementValid(state, wall) {
		log.Printf("PlaceWall: invalid wall placement by user with id=%s in game with id=%s. Wall={%v %+v %+v}", userId, gameId, wall.Direction, *wall.Pos1, *wall.Pos2)
		return nil, errors.New("invalid wall placement")
	}

	player := service.getPlayer(state, userId)
	player.Walls--

	state.Walls = append(state.Walls, wall)
	state.Turn = service.getNextTurn(state)

	log.Printf("PlaceWall: user with id=%s placed a wall=(%v %+v, %+v) in game with id=%s", userId, wall.Direction, *wall.Pos1, *wall.Pos2, gameId)
	return state, nil
}

func (service *GameServiceImpl) getPlayer(state *GameState, playerId string) *Player {
	if state.Player1.UserId == playerId {
		return state.Player1
	}
	return state.Player2
}

func (service *GameServiceImpl) getNextTurn(state *GameState) string {
	if state.Turn == state.Player1.UserId {
		return state.Player2.UserId
	}
	return state.Player1.UserId
}
