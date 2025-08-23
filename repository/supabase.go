package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jackc/pgx/v5"

)

type User struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	GoogleID  string    `json:"google_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName を追加して GORM にテーブル名を指定
func (User) TableName() string {
	return "Users" // 既存のテーブル名を指定
}

// Event は Events テーブルのレコードを表します
type Event struct {
	ID               int64          `json:"id" gorm:"primaryKey"`
	HostUserID       int64          `json:"host_user_id"`
	Title            string         `json:"title"`
	Note             sql.NullString `json:"note"`
	ParticipantCount int64          `json:"participant_count"`
	Status           int64          `json:"status"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// EventCondition は EventConditions テーブルのレコードを表します
type EventCondition struct {
	ID          int64          `json:"id" gorm:"primaryKey"`
	EventID     int64          `json:"event_id"`
	PeriodStart time.Time      `json:"period_start"`
	PeriodEnd   time.Time      `json:"period_end"`
	TimeType    int            `json:"time_type"`
	TimeStart   sql.NullString `json:"time_start"`
	TimeEnd     sql.NullString `json:"time_end"`
	DurationMin int            `json:"duration_min"`
	CreatedAt   time.Time      `json:"created_at"`
}

// EventParticipant は EventParticipants テーブルのレコードを表します
type EventParticipant struct {
	ID       int64        `json:"id" gorm:"primaryKey"` // int8 から int64 に変更
	EventID  int64        `json:"event_id"`             // int8 から int64 に変更
	UserID   int64        `json:"user_id"`              // int8 から int64 に変更
	Status   int8         `json:"status"`
	JoinedAt sql.NullTime `json:"joined_at"`
}

// Availability は Availabilities テーブルのレコードを表します
type Availability struct {
	ID             int64     `json:"id" gorm:"primaryKey"` // int8 から int64 に変更
	EventID        int64     `json:"event_id"`             // int8 から int64 に変更
	UserID         int64     `json:"user_id"`              // int8 から int64 に変更
	AvailableDate  time.Time `json:"available_date"`
	AvailableStart string    `json:"available_start"`
	AvailableEnd   string    `json:"available_end"`
	Source         string    `json:"source"` // "google_calendar" or "manual"
	CreatedAt      time.Time `json:"created_at"`
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


func connectDB() (*pgx.Conn, error) {

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

// CreateUser は新しいユーザーをデータベースに作成します
func (r *SupabaseRepositoryImpl) CreateUser(ctx context.Context, user *User) error {
	// WithContext を使用してコンテキストをGORMの操作に引き継ぎます
	// Create メソッドでレコードを作成します
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		return fmt.Errorf("failed to create user: %w", result.Error)
	}
	log.Printf("successfully created user with ID: %d", user.ID)
	return nil
}
