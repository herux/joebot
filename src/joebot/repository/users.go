package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"

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
	// Check if the username already exists
	var existingUser models.UserInfo
	err := repo.db.GetContext(ctx, &existingUser, "SELECT username FROM user_info WHERE username = ?", user.Username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if existingUser.Username != "" {
		return errors.New("username already exists")
	}

	query := `INSERT INTO user_info (username, password, is_admin, token, ip_whitelisted)
              VALUES (:username, :password, :is_admin, :token, :ip_whitelisted)`
	_, err = repo.db.NamedExecContext(ctx, query, user)
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
		return nil, err
	}
	// check if user is found
	if user.Username == "" {
		return nil, errors.New("user not found")
	}
	// Generate a JWT token
	token, err := generateJWT(user.Username)
	if err != nil {
		return nil, err
	}

	// Update the user's token in the database
	err = repo.UpdateUserToken(ctx, user.Username, token)
	if err != nil {
		return nil, err
	}

	userResp := &models.UserResponse{
		Username:  user.Username,
		Token:     token,
		ExpiredAt: time.Now().Add(24 * time.Hour),
	}
	return userResp, nil
}

func generateJWT(username string) (string, error) {
	secretKey := []byte("your_secret_key")

	// Create the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})

	// Sign the token with the secret key
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (repo *UserRepository) GetUserByToken(ctx context.Context, token string) (*models.UserInfo, error) {
	var user models.UserInfo
	err := repo.db.GetContext(ctx, &user, "SELECT * FROM user_info WHERE token = ?", token)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (repo *UserRepository) UpdateUserToken(ctx context.Context, username, token string) error {
	query := `UPDATE user_info SET token = ? WHERE username = ?`
	_, err := repo.db.ExecContext(ctx, query, token, username)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepository) UpdateUserIPWhitelist(ctx context.Context, username string, ipWhitelist string) error {
	query := `UPDATE user_info SET ip_whitelisted = ? WHERE username = ?`
	_, err := repo.db.ExecContext(ctx, query, ipWhitelist, username)
	if err != nil {
		return err
	}
	return nil
}

func (repo *UserRepository) GetUserIPWhitelist(ctx context.Context, username string) ([]string, error) {
	var ipWhitelist []string
	var ipWhitelistStr string
	err := repo.db.GetContext(ctx, &ipWhitelistStr, "SELECT ip_whitelisted FROM user_info WHERE username = ?", username)
	if err != nil {
		return nil, err
	}

	ipWhitelist = strings.Split(ipWhitelistStr, ",")
	if err != nil {
		return nil, err
	}

	return ipWhitelist, nil
}
