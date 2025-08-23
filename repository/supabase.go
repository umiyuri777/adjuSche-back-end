package repository

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

type SupabaseRepositoryImpl struct {
	db *pgx.Conn
}

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
	log.Printf("SUPABASE_PASSWORD: %s", password)
	log.Printf("SUPABASE_DB_NAME: %s", dbName)

	if host == "" || port == "" || user == "" || password == "" || dbName == "" {
		return nil, fmt.Errorf("not all database connection parameters are set")
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", user, password, host, port, dbName)

	db, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := db.Ping(context.Background()); err != nil {
		db.Close(context.Background())
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("successfully connected to the database")
	return db, nil
}

func (r *SupabaseRepositoryImpl) InsertUser() {
	// ユーザーテーブルにデータを挿入する処理を実装
}
