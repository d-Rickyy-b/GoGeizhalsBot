package userstate

import "github.com/d-Rickyy-b/gogeizhalsbot/v2/internal/bot/models"

type UserState struct {
	State      State
	Priceagent models.PriceAgent
}

type State int

const (
	Idle             State = iota
	CreatePriceagent State = iota
	SetNotification  State = iota
)

var UserStates = map[int64]UserState{}
