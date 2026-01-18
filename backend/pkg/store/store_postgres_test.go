package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDSN() string {
	_ = godotenv.Load("../../.env")

	dsn := os.Getenv("TEST_DB_STRING")
	if dsn == "" {
		dsn = "host=localhost port=5432 user=postgres password=postgres dbname=pub_test sslmode=disable"
	}
	return dsn
}

func setupTestStore(t *testing.T) *PostgresStore {
	t.Helper()

	ctx := context.Background()
	store, err := NewPostgresStore(ctx, getTestDSN())
	if err != nil {
		t.Skipf("Skipping test: could not connect to test database: %v", err)
	}

	// Clean up tables before each test
	cleanupTables(t, store)

	return store
}

func cleanupTables(t *testing.T, s *PostgresStore) {
	t.Helper()

	// Use parameterized queries for each known table
	queries := []string{
		"DELETE FROM " + tablePrefix + "events",
		"DELETE FROM " + tablePrefix + "kv",
		"DELETE FROM " + tablePrefix + "conversation_messages",
	}

	for _, query := range queries {
		//nolint:gosec // Table names are hardcoded constants, not user input
		_, err := s.db.ExecContext(s.ctx, query)
		require.NoError(t, err)
	}
}

func TestPostgresStore_Events(t *testing.T) {
	store := setupTestStore(t)

	// Initially empty
	events, err := store.GetEvents()
	require.NoError(t, err)
	assert.Empty(t, events)

	// Add events
	require.NoError(t, store.AddEvent("event1"))
	require.NoError(t, store.AddEvent("event2"))
	require.NoError(t, store.AddEvent("event3"))

	events, err = store.GetEvents()
	require.NoError(t, err)
	assert.Len(t, events, 3)
	assert.Equal(t, []string{"event1", "event2", "event3"}, events)
}

func TestPostgresStore_EventsLimit(t *testing.T) {
	store := setupTestStore(t)

	// Add more than 500 events
	for i := 0; i < 510; i++ {
		require.NoError(t, store.AddEvent("event"))
	}

	events, err := store.GetEvents()
	require.NoError(t, err)
	assert.Len(t, events, 500)
}

func TestPostgresStore_Weight(t *testing.T) {
	store := setupTestStore(t)

	// Get weight when not set
	_, err := store.GetWeight()
	require.Error(t, err)

	// Set and get weight
	require.NoError(t, store.SetWeight(42.5))
	weight, err := store.GetWeight()
	require.NoError(t, err)
	assert.InEpsilon(t, 42.5, weight, 0.0001)

	// Update weight
	require.NoError(t, store.SetWeight(100.25))
	weight, err = store.GetWeight()
	require.NoError(t, err)
	assert.InEpsilon(t, 100.25, weight, 0.0001)
}

func TestPostgresStore_WeightAt(t *testing.T) {
	store := setupTestStore(t)

	// Get weight at when not set
	_, err := store.GetWeightAt()
	require.Error(t, err)

	// Set and get weight at
	now := time.Now().Truncate(time.Second)
	require.NoError(t, store.SetWeightAt(now))
	weightAt, err := store.GetWeightAt()
	require.NoError(t, err)
	assert.Equal(t, now.UTC(), weightAt.UTC())
}

func TestPostgresStore_ActiveKeg(t *testing.T) {
	store := setupTestStore(t)

	// Get active keg when not set
	_, err := store.GetActiveKeg()
	require.Error(t, err)

	// Set and get active keg
	require.NoError(t, store.SetActiveKeg(50))
	activeKeg, err := store.GetActiveKeg()
	require.NoError(t, err)
	assert.Equal(t, 50, activeKeg)

	// Update active keg
	require.NoError(t, store.SetActiveKeg(30))
	activeKeg, err = store.GetActiveKeg()
	require.NoError(t, err)
	assert.Equal(t, 30, activeKeg)
}

func TestPostgresStore_ActiveKegAt(t *testing.T) {
	store := setupTestStore(t)

	// Get active keg at when not set
	_, err := store.GetActiveKegAt()
	require.Error(t, err)

	// Set and get active keg at
	now := time.Now().Truncate(time.Second)
	require.NoError(t, store.SetActiveKegAt(now))
	activeKegAt, err := store.GetActiveKegAt()
	require.NoError(t, err)
	assert.Equal(t, now.UTC(), activeKegAt.UTC())
}

func TestPostgresStore_BeersLeft(t *testing.T) {
	store := setupTestStore(t)

	// Get beers left when not set
	_, err := store.GetBeersLeft()
	require.Error(t, err)

	// Set and get beers left
	require.NoError(t, store.SetBeersLeft(100))
	beersLeft, err := store.GetBeersLeft()
	require.NoError(t, err)
	assert.Equal(t, 100, beersLeft)
}

func TestPostgresStore_BeersTotal(t *testing.T) {
	store := setupTestStore(t)

	// Get beers total when not set
	_, err := store.GetBeersTotal()
	require.Error(t, err)

	// Set and get beers total
	require.NoError(t, store.SetBeersTotal(200))
	beersTotal, err := store.GetBeersTotal()
	require.NoError(t, err)
	assert.Equal(t, 200, beersTotal)
}

func TestPostgresStore_IsLow(t *testing.T) {
	store := setupTestStore(t)

	// Get is low when not set
	_, err := store.GetIsLow()
	require.Error(t, err)

	// Set and get is low
	require.NoError(t, store.SetIsLow(true))
	isLow, err := store.GetIsLow()
	require.NoError(t, err)
	assert.True(t, isLow)

	// Update is low
	require.NoError(t, store.SetIsLow(false))
	isLow, err = store.GetIsLow()
	require.NoError(t, err)
	assert.False(t, isLow)
}

func TestPostgresStore_IsOpen(t *testing.T) {
	store := setupTestStore(t)

	// Get is open when not set
	_, err := store.GetIsOpen()
	require.Error(t, err)

	// Set and get is open
	require.NoError(t, store.SetIsOpen(true))
	isOpen, err := store.GetIsOpen()
	require.NoError(t, err)
	assert.True(t, isOpen)

	// Update is open
	require.NoError(t, store.SetIsOpen(false))
	isOpen, err = store.GetIsOpen()
	require.NoError(t, err)
	assert.False(t, isOpen)
}

func TestPostgresStore_Warehouse(t *testing.T) {
	store := setupTestStore(t)

	// Get warehouse when not set
	_, err := store.GetWarehouse()
	require.Error(t, err)

	// Set and get warehouse
	warehouse := [5]int{10, 20, 30, 40, 50}
	require.NoError(t, store.SetWarehouse(warehouse))
	result, err := store.GetWarehouse()
	require.NoError(t, err)
	assert.Equal(t, warehouse, result)

	// Update warehouse
	warehouse2 := [5]int{1, 2, 3, 4, 5}
	require.NoError(t, store.SetWarehouse(warehouse2))
	result, err = store.GetWarehouse()
	require.NoError(t, err)
	assert.Equal(t, warehouse2, result)
}

func TestPostgresStore_LastOk(t *testing.T) {
	store := setupTestStore(t)

	// Get last ok when not set
	_, err := store.GetLastOk()
	require.Error(t, err)

	// Set and get last ok
	now := time.Now().Truncate(time.Second)
	require.NoError(t, store.SetLastOk(now))
	lastOk, err := store.GetLastOk()
	require.NoError(t, err)
	assert.Equal(t, now.UTC(), lastOk.UTC())
}

func TestPostgresStore_OpenAt(t *testing.T) {
	store := setupTestStore(t)

	// Get open at when not set
	_, err := store.GetOpenAt()
	require.Error(t, err)

	// Set and get open at
	now := time.Now().Truncate(time.Second)
	require.NoError(t, store.SetOpenAt(now))
	openAt, err := store.GetOpenAt()
	require.NoError(t, err)
	assert.Equal(t, now.UTC(), openAt.UTC())
}

func TestPostgresStore_CloseAt(t *testing.T) {
	store := setupTestStore(t)

	// Get close at when not set
	_, err := store.GetCloseAt()
	require.Error(t, err)

	// Set and get close at
	now := time.Now().Truncate(time.Second)
	require.NoError(t, store.SetCloseAt(now))
	closeAt, err := store.GetCloseAt()
	require.NoError(t, err)
	assert.Equal(t, now.UTC(), closeAt.UTC())
}

func TestPostgresStore_TodayBeer(t *testing.T) {
	store := setupTestStore(t)

	// Get today beer when not set (should return empty string, no error)
	todayBeer, err := store.GetTodayBeer()
	require.NoError(t, err)
	assert.Empty(t, todayBeer)

	// Set and get today beer
	require.NoError(t, store.SetTodayBeer("Pilsner Urquell"))
	todayBeer, err = store.GetTodayBeer()
	require.NoError(t, err)
	assert.Equal(t, "Pilsner Urquell", todayBeer)

	// Reset today beer
	require.NoError(t, store.ResetTodayBeer())
	todayBeer, err = store.GetTodayBeer()
	require.NoError(t, err)
	assert.Empty(t, todayBeer)
}

func TestPostgresStore_Conversation(t *testing.T) {
	store := setupTestStore(t)

	convID := "test-conversation-123"

	// Get conversation when empty
	messages, err := store.GetConversation(convID)
	require.NoError(t, err)
	assert.Empty(t, messages)

	// Add messages
	msg1 := ConservationMessage{
		ID:      "msg1",
		Message: "Hello",
		At:      time.Now().Truncate(time.Second),
		Author:  ConversationMessageAuthorUser,
	}
	msg2 := ConservationMessage{
		ID:      "msg2",
		Message: "Hi there!",
		At:      time.Now().Truncate(time.Second).Add(time.Second),
		Author:  ConversationMessageAuthorBot,
	}

	require.NoError(t, store.AddConversationMessage(convID, msg1))
	require.NoError(t, store.AddConversationMessage(convID, msg2))

	// Get conversation
	messages, err = store.GetConversation(convID)
	require.NoError(t, err)
	require.Len(t, messages, 2)

	assert.Equal(t, "msg1", messages[0].ID)
	assert.Equal(t, "Hello", messages[0].Message)
	assert.Equal(t, ConversationMessageAuthorUser, messages[0].Author)

	assert.Equal(t, "msg2", messages[1].ID)
	assert.Equal(t, "Hi there!", messages[1].Message)
	assert.Equal(t, ConversationMessageAuthorBot, messages[1].Author)

	// Reset conversation
	require.NoError(t, store.ResetConversation(convID))
	messages, err = store.GetConversation(convID)
	require.NoError(t, err)
	assert.Empty(t, messages)
}

func TestPostgresStore_ConversationIsolation(t *testing.T) {
	store := setupTestStore(t)

	convID1 := "conversation-1"
	convID2 := "conversation-2"

	msg1 := ConservationMessage{
		ID:      "msg1",
		Message: "Message in conv 1",
		At:      time.Now(),
		Author:  ConversationMessageAuthorUser,
	}
	msg2 := ConservationMessage{
		ID:      "msg2",
		Message: "Message in conv 2",
		At:      time.Now(),
		Author:  ConversationMessageAuthorBot,
	}

	require.NoError(t, store.AddConversationMessage(convID1, msg1))
	require.NoError(t, store.AddConversationMessage(convID2, msg2))

	// Check isolation
	messages1, err := store.GetConversation(convID1)
	require.NoError(t, err)
	require.Len(t, messages1, 1)
	assert.Equal(t, "Message in conv 1", messages1[0].Message)

	messages2, err := store.GetConversation(convID2)
	require.NoError(t, err)
	require.Len(t, messages2, 1)
	assert.Equal(t, "Message in conv 2", messages2[0].Message)

	// Reset one conversation doesn't affect the other
	require.NoError(t, store.ResetConversation(convID1))

	messages1, err = store.GetConversation(convID1)
	require.NoError(t, err)
	assert.Empty(t, messages1)

	messages2, err = store.GetConversation(convID2)
	require.NoError(t, err)
	require.Len(t, messages2, 1)
}

func TestPostgresStore_ConversationLimit(t *testing.T) {
	store := setupTestStore(t)

	convID := "limit-test"

	// Add more than 500 messages
	for i := 0; i < 510; i++ {
		msg := ConservationMessage{
			ID:      "msg",
			Message: "message",
			At:      time.Now(),
			Author:  ConversationMessageAuthorUser,
		}
		require.NoError(t, store.AddConversationMessage(convID, msg))
	}

	messages, err := store.GetConversation(convID)
	require.NoError(t, err)
	assert.Len(t, messages, 500)
}

func TestPostgresStore_MigrationsIdempotent(t *testing.T) {
	ctx := context.Background()
	dsn := getTestDSN()

	// Create store multiple times (simulating restarts)
	store1, err := NewPostgresStore(ctx, dsn)
	if err != nil {
		t.Skipf("Skipping test: could not connect to test database: %v", err)
	}

	// Set some data
	require.NoError(t, store1.SetWeight(42.0))

	// Create another store (simulating restart)
	store2, err := NewPostgresStore(ctx, dsn)
	require.NoError(t, err)

	// Data should still be there
	weight, err := store2.GetWeight()
	require.NoError(t, err)
	assert.InEpsilon(t, 42.0, weight, 0.0001)

	// Clean up
	cleanupTables(t, store2)
}
