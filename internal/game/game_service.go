package game

import (
	"log"
	"quoridor/internal/errors"
	"time"

	"github.com/google/uuid"
)

type GameService interface {
	GetGameById(gameId string) (*Game, error)
	GetPendingGames() ([]*Game, error)
	CreateGame(user1Id, user2Id string) (*Game, error)
	MakeMove(gameId, userId string, newPos *Position) (*Game, error)
	PlaceWall(gameId, userId string, wall *Wall) (*Game, error)
}

type GameServiceImpl struct {
	engine     GameEngine
	repository GameRepository
}

func NewGameService(engine GameEngine, repository GameRepository) GameService {
	return &GameServiceImpl{
		engine:     engine,
		repository: repository,
	}
}

func (service *GameServiceImpl) GetGameById(gameId string) (*Game, error) {
	log.Printf("Fetching game by id: gameId=%v", gameId)

	state, err := service.repository.GetGameById(gameId)
	if err != nil {
		log.Printf("Error while fetching game. err=%v", err)
		return nil, errors.ErrInternalError
	}

	if state == nil {
		log.Printf("Game with id=%s not found", gameId)
		return nil, errors.ErrGameNotFound
	}

	return state, nil
}

func (service *GameServiceImpl) GetPendingGames() ([]*Game, error) {
	log.Println("Fetching pending games")

	games, err := service.repository.GetGamesByStatus(GameStatusPending)
	if err != nil {
		log.Printf("Error while fetching pending games. err=%v", err)
		return nil, errors.ErrInternalError
	}

	log.Printf("Fetched %d pending games", len(games))
	return games, nil
}

func (service *GameServiceImpl) CreateGame(user1Id, user2Id string) (*Game, error) {
	log.Printf("Creating new game: user1Id=%v, user2Id=%v", user1Id, user2Id)

	gameId := uuid.NewString()
	now := time.Now()

	state := &Game{
		GameId:     gameId,
		GameStatus: GameStatusInProgress,
		Player1: &Player{
			UserId:   user1Id,
			Position: &Position{X: 4, Y: 0},
			Goal:     8,
			Walls:    10,
		},
		Player2: &Player{
			UserId:   user2Id,
			Position: &Position{X: 4, Y: 8},
			Goal:     0,
			Walls:    10,
		},
		Turn:      user1Id,
		Walls:     []*Wall{},
		Moves:     []*Move{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := service.repository.SaveGame(state)
	if err != nil {
		log.Printf("Error while saving the game. err=%v", err)
		return nil, errors.ErrInternalError
	}

	log.Printf("Created game with id=%s for users with ids %s and %s", gameId, user1Id, user2Id)
	return state, nil
}

func (service *GameServiceImpl) MakeMove(gameId, userId string, newPos *Position) (*Game, error) {
	log.Printf("Making move: gameId=%v, userId=%v, new position=%+v", gameId, userId, *newPos)

	state, err := service.repository.GetGameById(gameId)
	if err != nil {
		log.Printf("Error while fetching game. err=%v", err)
		return nil, errors.ErrInternalError
	}
	if state == nil {
		log.Printf("Game with id=%s not found", gameId)
		return nil, errors.ErrGameNotFound
	}

	if state.GameStatus != GameStatusInProgress {
		log.Printf("Game with id=%s is not in progress", gameId)
		return nil, errors.ErrGameNotInProgress
	}

	if state.Turn != userId {
		log.Printf("It is not turn of user with id=%s in game with id=%s", userId, gameId)
		return nil, errors.ErrNotPlayersTurn
	}

	if !service.engine.IsMoveValid(state, userId, newPos) {
		log.Printf("Invalid move by user with id=%s in game with id=%s. Move=%+v", userId, gameId, *newPos)
		return nil, errors.ErrInvalidMove
	}

	player := service.getPlayer(state, userId)
	player.Position = newPos

	move := &Move{
		UserId:    userId,
		Type:      MoveTypeMove,
		Position:  newPos,
		Timestamp: time.Now(),
	}
	state.Moves = append(state.Moves, move)
	state.UpdatedAt = time.Now()

	if service.engine.CheckWin(state, player) {
		state.GameStatus = GameStatusCompleted
		state.Winner = userId
		completedAt := time.Now()
		state.CompletedAt = &completedAt
		log.Printf("User with id=%s has won the game with id=%s", userId, gameId)
	} else {
		state.Turn = service.getNextTurn(state)
		log.Printf("User with id=%s moved to position=%+v in game with id=%s", userId, *newPos, gameId)
	}

	err = service.repository.SaveGame(state)
	if err != nil {
		log.Printf("Error while saving the game state. gameId=%v, err=%v", gameId, err)
		return nil, errors.ErrInternalError
	}

	return state, nil
}

func (service *GameServiceImpl) PlaceWall(gameId, userId string, wall *Wall) (*Game, error) {
	log.Printf("PlaceWall: gameId=%v, userId=%v, wall={%v %+v %+v}", gameId, userId, wall.Direction, *wall.Pos1, *wall.Pos2)

	state, err := service.repository.GetGameById(gameId)
	if err != nil {
		return nil, errors.ErrInternalError
	}
	if state == nil {
		log.Printf("PlaceWall: game with id=%s not found", gameId)
		return nil, errors.ErrGameNotFound
	}

	if state.GameStatus != GameStatusInProgress {
		log.Printf("PlaceWall: game with id=%s is not in progress", gameId)
		return nil, errors.ErrGameNotInProgress
	}

	if state.Turn != userId {
		log.Printf("PlaceWall: it is not user with id=%s turn in game with id=%s", userId, gameId)
		return nil, errors.ErrNotPlayersTurn
	}

	if !service.engine.IsWallPlacementValid(state, wall) {
		log.Printf("PlaceWall: invalid wall placement by user with id=%s in game with id=%s. Wall={%v %+v %+v}", userId, gameId, wall.Direction, *wall.Pos1, *wall.Pos2)
		return nil, errors.ErrInvalidWallPlacement
	}

	player := service.getPlayer(state, userId)
	player.Walls--

	move := &Move{
		UserId:    userId,
		Type:      MoveTypePlaceWall,
		Wall:      wall,
		Timestamp: time.Now(),
	}
	state.Moves = append(state.Moves, move)
	state.UpdatedAt = time.Now()

	state.Walls = append(state.Walls, wall)
	state.Turn = service.getNextTurn(state)

	err = service.repository.SaveGame(state)
	if err != nil {
		log.Printf("PlaceWall: error while saving the game state. gameId=%v, err=%v", gameId, err)
		return nil, errors.ErrInternalError
	}

	log.Printf("PlaceWall: user with id=%s placed a wall=(%v %+v, %+v) in game with id=%s", userId, wall.Direction, *wall.Pos1, *wall.Pos2, gameId)
	return state, nil
}

func (service *GameServiceImpl) getPlayer(state *Game, playerId string) *Player {
	if state.Player1.UserId == playerId {
		return state.Player1
	}
	return state.Player2
}

func (service *GameServiceImpl) getNextTurn(state *Game) string {
	if state.Turn == state.Player1.UserId {
		return state.Player2.UserId
	}
	return state.Player1.UserId
}
