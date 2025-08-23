package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Event は Events テーブルのレコードを表します
type Events struct {
	ID               int64          `json:"id" gorm:"primaryKey"`
	HostUserID       string         `json:"host_user_id" gorm:"type:uuid"`
	Title            string         `json:"title"`
	Note             sql.NullString `json:"note"`
	ParticipantCount int64          `json:"participant_count"`
	Status           int64          `json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (Events) TableName() string {
	return "Events"
}

// EventCondition は EventConditions テーブルのレコードを表します
type EventCondition struct {
	ID          int64          `json:"id" gorm:"primaryKey;autoIncrement"`
	EventID     int64          `json:"event_id"`
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	TimeType    int            `json:"time_type"`
	TimeStart   sql.NullString `json:"time_start"`
	TimeEnd     sql.NullString `json:"time_end"`
	DurationMin int            `json:"duration_min"`
	CreatedAt   time.Time      `json:"created_at"`
}

func (EventCondition) TableName() string {
	return "EventConditions"
}

// EventParticipant は EventParticipants テーブルのレコードを表します
type EventParticipant struct {
	ID       int64          `json:"id" gorm:"primaryKey"`     // int8 から int64 に変更
	EventID  int64          `json:"event_id"`                 // int8 から int64 に変更
	UserID   string         `json:"user_id" gorm:"type:uuid"` // uuid型に修正
	Status   int8           `json:"status"`
	JoinedAt sql.NullString `json:"joined_at"` // text型に変更
}

func (EventParticipant) TableName() string {
	return "EventParticipants"
}

// Availability は Availabilities テーブルのレコードを表します
type Availability struct {
	ID             int64     `json:"id" gorm:"primaryKey"`        // DB: int8
	EventID        int64     `json:"event_id"`                    // DB: int8
	UserID         string    `json:"user_id" gorm:"type:uuid"`    // DB: uuid
	AvailableDate  string    `json:"available_date"`              // DB: text (YYYY-MM-DD)
	AvailableStart string    `json:"available_start"`             // DB: text
	AvailableEnd   string    `json:"available_end"`               // DB: text
	Sourse         int8      `json:"sourse" gorm:"column:sourse"` // DB: int8 (0: google_calendar, 1: manual) - 実際のカラム名は sourse（タイポ）
	CreatedAt      time.Time `json:"created_at"`
}

func (Availability) TableName() string {
	return "Availabilities"
}

// Link は Links テーブルのレコードを表します
type Link struct {
	ID        int64     `json:"id" gorm:"primaryKey"` // int8 から int64 に変更
	EventID   int64     `json:"event_id"`             // int8 から int64 に変更
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

const (
	EventStatusDraft  = 0
	EventStatusOpen   = 1
	EventStatusClosed = 2
)

const (
	ParticipantStatusInvited  = 0
	ParticipantStatusAccepted = 1
	ParticipantStatusDeclined = 2
)

// SupabaseRepositoryImpl は GORM の DB インスタンスを保持します
type SupabaseRepositoryImpl struct {
	db *gorm.DB
}

// NewSupabaseRepository はリポジトリを初期化します
func NewSupabaseRepository() (*SupabaseRepositoryImpl, error) {
	db, err := connectDB()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	return &SupabaseRepositoryImpl{db: db}, nil
}

func connectDB() (*gorm.DB, error) {

	host := os.Getenv("SUPABASE_HOST")
	port := os.Getenv("SUPABASE_PORT")
	user := os.Getenv("SUPABASE_USER")
	password := os.Getenv("SUPABASE_PASSWORD")
	dbName := os.Getenv("SUPABASE_DB_NAME")

	log.Printf("SUPABASE_HOST: %s", host)
	log.Printf("SUPABASE_PORT: %s", port)
	log.Printf("SUPABASE_USER: %s", user)
	log.Printf("SUPABASE_DB_NAME: %s", dbName)

	if host == "" || port == "" || user == "" || password == "" || dbName == "" {
		return nil, fmt.Errorf("not all database connection parameters are set")
	}

	// GORM 用の DSN (Data Source Name) を作成
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=require TimeZone=Asia/Tokyo", host, user, password, dbName, port)

	// GORM を使って PostgreSQL に接続
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // ← プリペアドステートメントを使わない
	}), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database with gorm: %w", err)
	}

	// Ping を確認 (GORM v2 では明示的な Ping は不要な場合が多いですが、念のため)
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB from gorm.DB: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("successfully connected to the database with GORM")
	return db, nil
}

// // CreateUser は新しいユーザーをデータベースに作成します
// func (r *SupabaseRepositoryImpl) CreateUser(ctx context.Context, users *Users) error {
// 	// WithContext を使用してコンテキストをGORMの操作に引き継ぎます
// 	// Create メソッドでレコードを作成します
// 	result := r.db.WithContext(ctx).Create(users)
// 	if result.Error != nil {
// 		return fmt.Errorf("failed to create user: %w", result.Error)
// 	}
// 	log.Printf("successfully created user with ID: %d", users.ID)
// 	return nil
// }

func (r *SupabaseRepositoryImpl) CreateEvent(ctx context.Context, events *Events) error {
	result := r.db.WithContext(ctx).Create(events)
	if result.Error != nil {
		return fmt.Errorf("failed to create event: %w", result.Error)
	}
	log.Printf("successfully created event with ID: %d", events.ID)
	return nil
}

// CreateEventCondition は新しいイベント条件を作成します
func (r *SupabaseRepositoryImpl) CreateEventCondition(ctx context.Context, cond *EventCondition) error {
	result := r.db.WithContext(ctx).Omit("ID").Create(cond)
	if result.Error != nil {
		return fmt.Errorf("failed to create event condition: %w", result.Error)
	}
	log.Printf("successfully created event condition with ID: %d (event_id=%d)", cond.ID, cond.EventID)
	return nil
}

func (r *SupabaseRepositoryImpl) GetEventByID(ctx context.Context, eventID int64) (*Events, error) {
	var e Events
	if err := r.db.WithContext(ctx).First(&e, eventID).Error; err != nil {
		return nil, fmt.Errorf("failed to get event by id: %w", err)
	}
	return &e, nil
}

func (r *SupabaseRepositoryImpl) GetEventConditionByEventID(ctx context.Context, eventID int64) (*EventCondition, error) {
	var ec EventCondition
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Order("id DESC").First(&ec).Error; err != nil {
		return nil, fmt.Errorf("failed to get event condition by event_id: %w", err)
	}
	return &ec, nil
}

func (r *SupabaseRepositoryImpl) ListAvailabilitiesByEventID(ctx context.Context, eventID int64) ([]Availability, error) {
	var avs []Availability
	if err := r.db.WithContext(ctx).Where("event_id = ?", eventID).Find(&avs).Error; err != nil {
		return nil, fmt.Errorf("failed to list availabilities by event_id: %w", err)
	}
	return avs, nil
}

func (r *SupabaseRepositoryImpl) CountDistinctAvailabilityUsersByEventID(ctx context.Context, eventID int64) (int, error) {
	type Result struct{ Cnt int }
	var res Result
	if err := r.db.WithContext(ctx).Raw("SELECT COUNT(DISTINCT user_id) AS cnt FROM \"Availabilities\" WHERE event_id = ?", eventID).Scan(&res).Error; err != nil {
		return 0, fmt.Errorf("failed to count distinct users in availabilities: %w", err)
	}
	return res.Cnt, nil
}

// // GetUserByGoogleID は google_id から Users レコードを取得する
// func (r *SupabaseRepositoryImpl) GetUserByGoogleID(ctx context.Context, googleID string) (*Users, error) {
// 	var u Users
// 	if err := r.db.WithContext(ctx).Where("google_id = ?", googleID).First(&u).Error; err != nil {
// 		return nil, fmt.Errorf("failed to get user by google_id: %w", err)
// 	}
// 	return &u, nil
// }

// ReplaceUserAvailabilitiesForEvent は指定 event_id×user_id の既存データを削除し、与えられたレコードで置換する
func (r *SupabaseRepositoryImpl) ReplaceUserAvailabilitiesForEvent(ctx context.Context, eventID int64, userID string, avs []Availability) error {
	log.Printf("ReplaceUserAvailabilitiesForEvent: eventID=%d, userID=%s, records=%d", eventID, userID, len(avs))

	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}

	// 既存のレコードを削除（Google Calendar由来のデータ sourse=0 のみを削除）
	deleteResult := tx.Where("event_id = ? AND user_id = ? AND sourse = ?", eventID, userID, 0).Delete(&Availability{})
	if deleteResult.Error != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete existing availabilities: %w", deleteResult.Error)
	}
	log.Printf("削除されたレコード数: %d", deleteResult.RowsAffected)

	// 新しいレコードを挿入
	if len(avs) > 0 {
		createResult := tx.Omit("ID").Create(&avs)
		if createResult.Error != nil {
			_ = tx.Rollback()
			log.Printf("Create エラー: %v", createResult.Error)
			return fmt.Errorf("failed to create availabilities: %w", createResult.Error)
		}
		log.Printf("挿入されたレコード数: %d", createResult.RowsAffected)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	log.Printf("トランザクション完了")
	return nil
}

// GetOrCreateEventParticipant は参加者を取得または作成します
func (r *SupabaseRepositoryImpl) GetOrCreateEventParticipant(ctx context.Context, eventID int64, userID string) (*EventParticipant, error) {
	// 既存の参加者レコードを検索
	var participant EventParticipant
	err := r.db.WithContext(ctx).Where("event_id = ? AND user_id = ?", eventID, userID).First(&participant).Error

	if err == nil {
		// 既存のレコードが見つかった場合
		log.Printf("既存の参加者レコードを発見: ID=%d, Status=%d", participant.ID, participant.Status)
		return &participant, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		// 検索時のエラー（レコードが見つからない以外のエラー）
		return nil, fmt.Errorf("failed to search event participant: %w", err)
	}

	// レコードが見つからない場合、新しい参加者を作成
	log.Printf("新しい参加者レコードを作成: eventID=%d, userID=%s", eventID, userID)
	now := time.Now()
	newParticipant := EventParticipant{
		EventID:  eventID,
		UserID:   userID,
		Status:   ParticipantStatusAccepted, // /invite にアクセスした時点で受諾とみなす
		JoinedAt: sql.NullString{String: now.Format(time.RFC3339), Valid: true},
	}

	if err := r.db.WithContext(ctx).Create(&newParticipant).Error; err != nil {
		return nil, fmt.Errorf("failed to create event participant: %w", err)
	}

	log.Printf("新しい参加者レコードを作成しました: ID=%d", newParticipant.ID)
	return &newParticipant, nil
}
