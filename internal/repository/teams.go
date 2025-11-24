package repository

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type ITeamsRepository interface {
	AddTeam(tx *sql.Tx, teamName string) error
	IsExist(teamName string) (bool, error)
	GetDB() *sql.DB
}

type TeamsRepository struct {
	Db *sql.DB
}

func (tr *TeamsRepository) AddTeam(tx *sql.Tx, teamName string) error {
	stmt := "INSERT INTO teams(team_name) VALUES($1)"

	var err error
	if tx != nil {
		_, err = tx.Exec(stmt, teamName)
	} else {
		_, err = tr.Db.Exec(stmt, teamName)
	}

	//ошибка во время операции или из-за дубликата названия команды
	if err != nil {
		var sqlError *pq.Error
		if errors.As(err, &sqlError) {
			if sqlError.Code == "23505" {
				return ErrDuplicatedTeamName
			}
		}
		return err
	}

	return nil
}

func (tr *TeamsRepository) IsExist(teamName string) (bool, error) {
	stmt := `SELECT EXISTS(SELECT team_name FROM teams WHERE team_name = $1)`

	var isExist bool
	err := tr.Db.QueryRow(stmt, teamName).Scan(&isExist)
	return isExist, err
}

func (tr *TeamsRepository) GetDB() *sql.DB {
	return tr.Db
}
