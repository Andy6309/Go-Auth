package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"-"` // Don't include password in JSON
	CreatedAt time.Time `json:"created_at"`
}

// HashPassword hashes the user's password
func (u *User) HashPassword() error {
	if u.Password == "" {
		return errors.New("password cannot be empty")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword compares the provided password with the stored hash
func (u *User) CheckPassword(password string) error {
	if u.Password == "" {
		return errors.New("no password hash available for comparison")
	}
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(user *User) error {
	if user.Username == "" {
		return errors.New("username cannot be empty")
	}

	// Check if user already exists
	existingUser, _ := r.GetUserByUsername(user.Username)
	if existingUser != nil {
		return fmt.Errorf("username %s already exists", user.Username)
	}

	// Hash the password
	if err := user.HashPassword(); err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	query := `
		INSERT INTO users (username, password, created_at)
		VALUES (?, ?, ?)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		user.Username,
		user.Password,
		time.Now().UTC(),
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	return nil
}

// GetUserByUsername retrieves a user by their username
func (r *UserRepository) GetUserByUsername(username string) (*User, error) {
	if username == "" {
		return nil, errors.New("username cannot be empty")
	}

	user := &User{}
	query := `
		SELECT id, username, password, created_at
		FROM users
		WHERE username = ?
	`

	err := r.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Password,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("database error: %w", err)
	}

	return user, nil
}