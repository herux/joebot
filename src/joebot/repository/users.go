package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/harmonicinc-com/joebot/models"
	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

type UserRepository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (repo *UserRepository) Create(ctx context.Context, user *models.UserInfo) error {
	query := `INSERT INTO user_info (username, password, is_admin, token, ip_whitelisted)
              VALUES (:username, :password, :is_admin, :token, :ip_whitelisted)`
	_, err := repo.db.NamedExecContext(ctx, query, user)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepository) GetAll(ctx context.Context) ([]*models.UserInfo, error) {
	var users []*models.UserInfo
	err := repo.db.SelectContext(ctx, &users, "SELECT id, username, password, is_admin, ip_whitelisted FROM user_info")
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (repo *UserRepository) GetUserByUsername(ctx context.Context, username string) (*models.UserInfo, error) {
	var user models.UserInfo
	err := repo.db.GetContext(ctx, &user, "SELECT * FROM user_info WHERE username = ?", username)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) GetUserByUserPassword(ctx context.Context, username, password string) (*models.UserResponse, error) {
	var user models.UserInfo
	query := `SELECT * FROM user_info WHERE username = ? AND password = ?`
	err := repo.db.GetContext(ctx, &user, query, username, password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	userResp := &models.UserResponse{
		Username:  user.Username,
		Token:     "token",
		ExpiredAt: time.Now(),
	}
	return userResp, nil
}
