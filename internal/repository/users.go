package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"pr-service/internal/models"

	"github.com/lib/pq"
)

type IUsersRepository interface {
	AddUser(tx *sql.Tx, user *models.UserModel) error
	GetUsersByTeam(tx *sql.Tx, teamName string) ([]*models.UserModel, error)
	GetUserById(tx *sql.Tx, id string) (*models.UserModel, error)
	UpdateUserIsActive(id string, isActive bool) error
}

type UsersRepository struct {
	Db *sql.DB
}

func (us *UsersRepository) AddUser(tx *sql.Tx, user *models.UserModel) error {
	stmt := "INSERT INTO users (user_id, username, team_name, is_active) VALUES($1, $2, $3, $4)"

	var err error
	if tx != nil {
		_, err = tx.Exec(stmt, user.Id, user.Username, user.TeamName, user.IsActive)
	} else {
		_, err = us.Db.Exec(stmt, user.Id, user.Username, user.TeamName, user.IsActive)
	}

	//ошибка во время операции или из-за дубликата user_id
	if err != nil {
		var sqlError *pq.Error
		if errors.As(err, &sqlError) {
			if sqlError.Code == "23505" {
				return ErrDuplicatedUserId
			}
		}
		return err
	}

	return nil
}

func (us *UsersRepository) GetUsersByTeam(tx *sql.Tx, teamName string) ([]*models.UserModel, error) {
	stmt := "SELECT user_id, username, is_active, team_name FROM users WHERE team_name = $1"

	var err error
	var rows *sql.Rows
	if tx != nil {
		rows, err = tx.Query(stmt, teamName)
	} else {
		rows, err = us.Db.Query(stmt, teamName)
	}

	//ошибки может быть только в процессе запроса
	if err != nil {
		return nil, err
	}

	defer func() {
		if errClose := rows.Close(); errClose != nil {
			if err == nil {
				err = fmt.Errorf("failed to close database rows: %w", errClose)
			}
		}
	}()

	var users []*models.UserModel
	for rows.Next() {
		var user models.UserModel
		if err := rows.Scan(&user.Id, &user.Username, &user.IsActive, &user.TeamName); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (us *UsersRepository) GetUserById(tx *sql.Tx, id string) (*models.UserModel, error) {
	stmt := "SELECT user_id, username, is_active, team_name FROM users WHERE user_id = $1"

	var err error
	user := models.UserModel{}
	if tx != nil {
		err = tx.QueryRow(stmt, id).Scan(&user.Id, &user.Username, &user.IsActive, &user.TeamName)
	} else {
		err = us.Db.QueryRow(stmt, id).Scan(&user.Id, &user.Username, &user.IsActive, &user.TeamName)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}

	return &user, nil
}

func (us *UsersRepository) UpdateUserIsActive(id string, isActive bool) error {
	stmt := "UPDATE users SET is_active = $1 WHERE user_id = $2"

	if _, err := us.Db.Exec(stmt, isActive, id); err != nil {
		return err
	}

	return nil
}
