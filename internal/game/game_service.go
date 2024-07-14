package game

import (
	"errors"
	"log"
	"sync"

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
	games      map[string]*Game
	mu         sync.Mutex
}

func NewGameService(engine GameEngine, repository GameRepository) GameService {
	return &GameServiceImpl{
		engine:     engine,
		repository: repository,
		games:      make(map[string]*Game),
	}
}

func (service *GameServiceImpl) GetGameById(gameId string) (*Game, error) {
	log.Printf("GetGameById: gameId=%v", gameId)

	state, err := service.repository.GetGameById(gameId)
	if err != nil {
		log.Printf("GetGameById: error while fetching game. err=%v", err)
		return nil, err
	}

	if state == nil {
		log.Printf("GetGameById: game with id=%s not found", gameId)
		return nil, errors.New("game not found")
	}

	return state, nil
}

func (service *GameServiceImpl) GetPendingGames() ([]*Game, error) {
	log.Println("GetPendingGames")

	games, err := service.repository.GetGamesByStatus(GameStatusPending)
	if err != nil {
		log.Printf("GetPendingGames: error while fetching pending games. err=%v", err)
		return nil, err
	}

	log.Printf("GetPendingGames: found %d pending games", len(games))
	return games, nil
}

func (service *GameServiceImpl) CreateGame(user1Id, user2Id string) (*Game, error) {
	log.Printf("CreateGame: user1Id=%v, user2Id=%v", user1Id, user2Id)

	gameId := uuid.NewString()
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
		Turn:  user1Id,
		Walls: []*Wall{},
	}

	err := service.repository.SaveGame(state)
	if err != nil {
		log.Printf("CreateGame: error while saving the game. err=%v", err)
		return nil, err
	}

	log.Printf("CreateGame: created game with id=%s for users with ids %s and %s", gameId, user1Id, user2Id)
	return state, nil
}

func (service *GameServiceImpl) MakeMove(gameId, userId string, newPos *Position) (*Game, error) {
	log.Printf("MakeMove: gameId=%v, userId=%v, new position=%+v", gameId, userId, *newPos)

	state, err := service.repository.GetGameById(gameId)
	if err != nil {
		return nil, err
	}
	if state == nil {
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
	} else {
		state.Turn = service.getNextTurn(state)
		log.Printf("MakeMove: user with id=%s moved to position=%+v in game with id=%s", userId, *newPos, gameId)
	}

	err = service.repository.SaveGame(state)
	if err != nil {
		log.Printf("MakeMove: error while saving the game state. gameId=%v, err=%v", gameId, err)
		return nil, err
	}

	return state, nil
}

func (service *GameServiceImpl) PlaceWall(gameId, userId string, wall *Wall) (*Game, error) {
	log.Printf("PlaceWall: gameId=%v, userId=%v, wall={%v %+v %+v}", gameId, userId, wall.Direction, *wall.Pos1, *wall.Pos2)

	state, err := service.repository.GetGameById(gameId)
	if err != nil {
		return nil, err
	}
	if state == nil {
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

	err = service.repository.SaveGame(state)
	if err != nil {
		log.Printf("PlaceWall: error while saving the game state. gameId=%v, err=%v", gameId, err)
		return nil, err
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
