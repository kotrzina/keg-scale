package store

import "time"

type ConversationMessageAuthor string

const (
	ConversationMessageAuthorUser ConversationMessageAuthor = "user"
	ConversationMessageAuthorBot  ConversationMessageAuthor = "bot"
)

type ConservationMessage struct {
	ID      string                    `json:"id"`
	Message string                    `json:"msg"`
	At      time.Time                 `json:"at"`
	Author  ConversationMessageAuthor `json:"author"` // user or bot
}

type Storage interface {
	AddEvent(event string) error  // add event
	GetEvents() ([]string, error) // get events

	SetWeight(weight float64) error // set weight
	GetWeight() (float64, error)    // get weight

	SetWeightAt(weightAt time.Time) error // set weight at
	GetWeightAt() (time.Time, error)      // get weight at

	SetActiveKeg(weight int) error // set active keg
	GetActiveKeg() (int, error)    // get active keg

	SetActiveKegAt(at time.Time) error  // set active keg at
	GetActiveKegAt() (time.Time, error) // get active keg at

	SetBeersLeft(beersLeft int) error // set beers left
	GetBeersLeft() (int, error)       // get beers left

	SetBeersTotal(beersTotal int) error // set beers total
	GetBeersTotal() (int, error)        // get beers total

	SetIsLow(isLow bool) error // set is low flag
	GetIsLow() (bool, error)   // get is low flag

	SetWarehouse(warehouse [5]int) error // set warehouse
	GetWarehouse() ([5]int, error)       // get warehouse

	SetLastOk(lastOk time.Time) error // set last ok
	GetLastOk() (time.Time, error)    // get last ok

	SetOpenAt(openAt time.Time) error // set open at
	GetOpenAt() (time.Time, error)    // get open at

	SetCloseAt(closeAt time.Time) error // set close at
	GetCloseAt() (time.Time, error)     // get close at

	SetIsOpen(isOpen bool) error // set is open flag
	GetIsOpen() (bool, error)    // get is open flag

	SetTodayBeer(todayBeer string) error // set today beer
	GetTodayBeer() (string, error)       // get today beer
	ResetTodayBeer() error               // reset today beer

	AddConversationMessage(id string, msg ConservationMessage) error // add conversation message
	GetConversation(id string) ([]ConservationMessage, error)        // get conversation messages from oldest to newest
	ResetConversation(id string) error                               // reset conversation - delete all messages
}
