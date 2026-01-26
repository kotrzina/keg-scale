package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

const (
	tablePrefix = "pub_"
)

type PostgresStore struct {
	db  *sql.DB
	ctx context.Context
}

func NewPostgresStore(ctx context.Context, dsn string) (*PostgresStore, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	store := &PostgresStore{
		db:  db,
		ctx: ctx,
	}

	if err := store.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return store, nil
}

func (s *PostgresStore) migrate() error {
	migrations := []string{
		// Events table
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sevents (
			id SERIAL PRIMARY KEY,
			event TEXT NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`, tablePrefix),

		// Key-value store for simple values
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %skv (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`, tablePrefix),

		// Conversation messages
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %sconversation_messages (
			id SERIAL PRIMARY KEY,
			conv_id TEXT NOT NULL,
			msg_id TEXT NOT NULL,
			message TEXT NOT NULL,
			author TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`, tablePrefix),

		// Index for conversation messages
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS %sconversation_messages_conv_id_idx ON %sconversation_messages (conv_id)`,
			tablePrefix, tablePrefix),
	}

	for _, migration := range migrations {
		if _, err := s.db.ExecContext(s.ctx, migration); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// Helper methods for key-value store

func (s *PostgresStore) setValue(key, value string) error {
	query := fmt.Sprintf(`
		INSERT INTO %skv (key, value, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()
	`, tablePrefix)
	_, err := s.db.ExecContext(s.ctx, query, key, value)
	return err
}

func (s *PostgresStore) getValue(key string) (string, error) {
	var value string
	query := fmt.Sprintf("SELECT value FROM %skv WHERE key = $1", tablePrefix)
	err := s.db.QueryRowContext(s.ctx, query, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("key not found: %s", key)
	}
	return value, err
}

func (s *PostgresStore) deleteValue(key string) error {
	query := fmt.Sprintf("DELETE FROM %skv WHERE key = $1", tablePrefix)
	_, err := s.db.ExecContext(s.ctx, query, key)
	return err
}

func (s *PostgresStore) setMap(key string, m map[string]string) error {
	data, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal map: %w", err)
	}
	return s.setValue(key, string(data))
}

func (s *PostgresStore) getMap(key string) (map[string]string, error) {
	val, err := s.getValue(key)
	if err != nil {
		// the key does not exist - no ideal solution
		return map[string]string{}, nil
	}

	var m map[string]string
	if err := json.Unmarshal([]byte(val), &m); err != nil {
		return nil, fmt.Errorf("failed to unmarshal map: %w", err)
	}
	return m, nil
}

func (s *PostgresStore) setStructArray(key string, arr interface{}) error {
	data, err := json.Marshal(arr)
	if err != nil {
		return fmt.Errorf("failed to marshal array: %w", err)
	}
	return s.setValue(key, string(data))
}

func (s *PostgresStore) getStructArray(key string, result interface{}) error {
	val, err := s.getValue(key)
	if err != nil {
		return err
	}

	if err := json.Unmarshal([]byte(val), result); err != nil {
		return fmt.Errorf("failed to unmarshal array: %w", err)
	}
	return nil
}

// Storage interface implementation

func (s *PostgresStore) AddEvent(event string) error {
	// Insert event
	query := fmt.Sprintf("INSERT INTO %sevents (event) VALUES ($1)", tablePrefix)
	if _, err := s.db.ExecContext(s.ctx, query, event); err != nil {
		return err
	}

	// Keep only last 500 events
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %sevents
		WHERE id NOT IN (
			SELECT id FROM %sevents ORDER BY id DESC LIMIT 500
		)
	`, tablePrefix, tablePrefix)
	_, err := s.db.ExecContext(s.ctx, deleteQuery)
	return err
}

func (s *PostgresStore) GetEvents() ([]string, error) {
	query := fmt.Sprintf("SELECT event FROM %sevents ORDER BY id ASC", tablePrefix)
	rows, err := s.db.QueryContext(s.ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var events []string
	for rows.Next() {
		var event string
		if err := rows.Scan(&event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func (s *PostgresStore) SetWeight(weight float64) error {
	return s.setValue("weight", fmt.Sprintf("%f", weight))
}

func (s *PostgresStore) GetWeight() (float64, error) {
	val, err := s.getValue("weight")
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(val, 64)
}

func (s *PostgresStore) SetWeightAt(weightAt time.Time) error {
	return s.setValue("weight_at", weightAt.Format(time.RFC3339))
}

func (s *PostgresStore) GetWeightAt() (time.Time, error) {
	val, err := s.getValue("weight_at")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

func (s *PostgresStore) SetActiveKeg(weight int) error {
	return s.setValue("active_keg", strconv.Itoa(weight))
}

func (s *PostgresStore) GetActiveKeg() (int, error) {
	val, err := s.getValue("active_keg")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (s *PostgresStore) SetActiveKegAt(at time.Time) error {
	return s.setValue("active_keg_at", at.Format(time.RFC3339))
}

func (s *PostgresStore) GetActiveKegAt() (time.Time, error) {
	val, err := s.getValue("active_keg_at")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

func (s *PostgresStore) SetBeersLeft(beersLeft int) error {
	return s.setValue("beers_left", strconv.Itoa(beersLeft))
}

func (s *PostgresStore) GetBeersLeft() (int, error) {
	val, err := s.getValue("beers_left")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (s *PostgresStore) SetBeersTotal(beersTotal int) error {
	return s.setValue("beers_total", strconv.Itoa(beersTotal))
}

func (s *PostgresStore) GetBeersTotal() (int, error) {
	val, err := s.getValue("beers_total")
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(val)
}

func (s *PostgresStore) SetIsLow(isLow bool) error {
	return s.setValue("is_low", strconv.FormatBool(isLow))
}

func (s *PostgresStore) GetIsLow() (bool, error) {
	val, err := s.getValue("is_low")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(val)
}

func (s *PostgresStore) SetWarehouse(warehouse [5]int) error {
	val := fmt.Sprintf("%d,%d,%d,%d,%d", warehouse[0], warehouse[1], warehouse[2], warehouse[3], warehouse[4])
	return s.setValue("warehouse", val)
}

func (s *PostgresStore) GetWarehouse() ([5]int, error) {
	val, err := s.getValue("warehouse")
	if err != nil {
		return [5]int{0, 0, 0, 0, 0}, err
	}

	var warehouse [5]int
	parts := strings.Split(val, ",")

	if len(parts) != 5 {
		return [5]int{0, 0, 0, 0, 0}, fmt.Errorf("invalid warehouse format in the storage")
	}

	for i, part := range parts {
		x, err := strconv.Atoi(part)
		if err != nil {
			return [5]int{0, 0, 0, 0, 0}, fmt.Errorf("invalid warehouse format in the storage")
		}
		warehouse[i] = x
	}

	return warehouse, nil
}

func (s *PostgresStore) SetLastOk(lastOk time.Time) error {
	return s.setValue("last_ok", lastOk.Format(time.RFC3339))
}

func (s *PostgresStore) GetLastOk() (time.Time, error) {
	val, err := s.getValue("last_ok")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

func (s *PostgresStore) SetOpenAt(openAt time.Time) error {
	return s.setValue("open_at", openAt.Format(time.RFC3339))
}

func (s *PostgresStore) GetOpenAt() (time.Time, error) {
	val, err := s.getValue("open_at")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

func (s *PostgresStore) SetCloseAt(closeAt time.Time) error {
	return s.setValue("close_at", closeAt.Format(time.RFC3339))
}

func (s *PostgresStore) GetCloseAt() (time.Time, error) {
	val, err := s.getValue("close_at")
	if err != nil {
		return time.Time{}, err
	}
	return time.Parse(time.RFC3339, val)
}

func (s *PostgresStore) SetIsOpen(isOpen bool) error {
	return s.setValue("is_open", strconv.FormatBool(isOpen))
}

func (s *PostgresStore) GetIsOpen() (bool, error) {
	val, err := s.getValue("is_open")
	if err != nil {
		return false, err
	}
	return strconv.ParseBool(val)
}

func (s *PostgresStore) SetTodayBeer(todayBeer string) error {
	return s.setValue("today_beer", todayBeer)
}

func (s *PostgresStore) GetTodayBeer() (string, error) {
	val, err := s.getValue("today_beer")
	if err != nil {
		//nolint:nilerr // Return empty string on error like Redis does
		return "", nil
	}
	return val, nil
}

func (s *PostgresStore) ResetTodayBeer() error {
	return s.deleteValue("today_beer")
}

func (s *PostgresStore) AddConversationMessage(id string, msg ConservationMessage) error {
	query := fmt.Sprintf(`
		INSERT INTO %sconversation_messages (conv_id, msg_id, message, author, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, tablePrefix)
	if _, err := s.db.ExecContext(s.ctx, query, id, msg.ID, msg.Message, string(msg.Author), msg.At); err != nil {
		return fmt.Errorf("failed to add conversation message: %w", err)
	}

	// Keep only last 500 messages per conversation
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %sconversation_messages
		WHERE conv_id = $1 AND id NOT IN (
			SELECT id FROM %sconversation_messages WHERE conv_id = $1 ORDER BY id DESC LIMIT 500
		)
	`, tablePrefix, tablePrefix)
	_, err := s.db.ExecContext(s.ctx, deleteQuery, id)
	return err
}

func (s *PostgresStore) GetConversation(id string) ([]ConservationMessage, error) {
	query := fmt.Sprintf(`
		SELECT msg_id, message, author, created_at
		FROM %sconversation_messages
		WHERE conv_id = $1
		ORDER BY id ASC
	`, tablePrefix)
	rows, err := s.db.QueryContext(s.ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation messages: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var messages []ConservationMessage
	for rows.Next() {
		var msg ConservationMessage
		var author string
		if err := rows.Scan(&msg.ID, &msg.Message, &author, &msg.At); err != nil {
			return nil, fmt.Errorf("failed to scan conversation message: %w", err)
		}
		msg.Author = ConversationMessageAuthor(author)
		messages = append(messages, msg)
	}

	return messages, rows.Err()
}

func (s *PostgresStore) ResetConversation(id string) error {
	query := fmt.Sprintf("DELETE FROM %sconversation_messages WHERE conv_id = $1", tablePrefix)
	_, err := s.db.ExecContext(s.ctx, query, id)
	return err
}

func (s *PostgresStore) SetAttendanceKnownDevices(devices map[string]string) error {
	return s.setMap("attendance_known_devices", devices)
}

func (s *PostgresStore) GetAttendanceKnownDevices() (map[string]string, error) {
	return s.getMap("attendance_known_devices")
}

func (s *PostgresStore) SetAttendanceIrks(irks map[string]string) error {
	return s.setMap("attendance_irks", irks)
}

func (s *PostgresStore) GetAttendanceIrks() (map[string]string, error) {
	return s.getMap("attendance_irks")
}
