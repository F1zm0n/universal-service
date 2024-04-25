package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/F1zm0n/uni-auth/repository"
)

type Postgres struct {
	conn *gorm.DB
}

func MustNewPostgresDB() *Postgres {
	db, err := connectPostgres()
	if err != nil {
		panic(err)
	}
	return &Postgres{
		conn: db,
	}
}

func connectPostgres() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		viper.GetString("postgres.host"),
		viper.GetString("postgres.user"),
		viper.GetString("postgres.password"),
		viper.GetString("postgres.dbname"),
		viper.GetString("postgres.port"),
		viper.GetString("postgres.sslmode"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	return db, err
}

func (p Postgres) MustMigrateSchema() {
	err := p.conn.AutoMigrate(repository.User{})
	if err != nil {
		panic(err)
	}
	return
}

func (p Postgres) InsertUser(ctx context.Context, user repository.User) error {
	res := p.conn.WithContext(ctx).Create(&user)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("no rows were affected")
	}
	return nil
}

func (p Postgres) GetUserByEmail(ctx context.Context, email string) (repository.User, error) {
	var user repository.User
	res := p.conn.WithContext(ctx).Where("email = ?", email).First(&user)
	if res.Error != nil {
		return repository.User{}, res.Error
	}
	return user, nil
}

func (p Postgres) GetUserById(ctx context.Context, id uuid.UUID) (repository.User, error) {
	var user repository.User
	res := p.conn.WithContext(ctx).Where("id= ?", id).First(&user)
	if res.Error != nil {
		return repository.User{}, res.Error
	}
	return user, nil
}
