package matchmaking

import "time"

type Match struct {
	User1Id string
	User2Id string
}

type MatchRequest struct {
	UserId   string
	JoinTime time.Time
}
