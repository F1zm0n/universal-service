package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type Repository interface {
	CreateLink(u VerEntity) (*sql.Tx, error)
	GetByVerId(id uuid.UUID) (VerEntity, error)
	DeleteByVerId(id uuid.UUID) (*sql.Tx, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func (r PostgresRepository) mustInitSchema() {
	schema := `
	CREATE TABLE verification(
		email VARCHAR(100) NOT NULL UNIQUE,
		password VARCHAR(100) NOT NULL,
		ver_id uuid PRIMARY KEY
	);
	`
	_, err := r.db.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
}

func NewPostgresRepository() Repository {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		"postgres",
		"postgres",
		"password",
		"users",
		"5432",
		"disable",
	)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	repo := PostgresRepository{
		db: db,
	}
	repo.mustInitSchema()

	return &repo
}

func (r PostgresRepository) CreateLink(user VerEntity) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(context.Background(), &sql.TxOptions{})
	_, err = tx.Exec(
		"INSERT INTO verification (email,password,ver_id) VALUES ($1,$2,$3)",
		user.Email,
		user.Password,
		user.VerId,
	)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (r PostgresRepository) GetByVerId(id uuid.UUID) (VerEntity, error) {
	var user VerEntity
	err := r.db.QueryRow("SELECT ver_id,email,password FROM verification WHERE ver_id=$1", id).
		Scan(&user.VerId,
			&user.Email,
			&user.Password,
		)
	if err != nil {
		return VerEntity{}, err
	}
	return user, nil
}

func (r PostgresRepository) DeleteByVerId(id uuid.UUID) (*sql.Tx, error) {
	tx, err := r.db.BeginTx(context.Background(), nil)
	if err != nil {
		return nil, err
	}
	res, err := tx.Exec("DELETE FROM verification WHERE ver_id=$1", id)
	if err != nil {
		return nil, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if rows != 1 {
		return nil, fmt.Errorf("error deleting verification")
	}
	return tx, nil
}
