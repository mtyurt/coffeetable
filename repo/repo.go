package repo

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	ct "github.com/mtyurt/coffeetable"
)

type repo struct {
	db *sql.DB
}
type Repo interface {
	GetUserRelations() ([]ct.UserRelation, error)
	UpdateEncounters(ct.UserRelation) error
}

func New(db *sql.DB) Repo {
	return &repo{db}
}
func (r *repo) GetUserRelations() ([]ct.UserRelation, error) {
	if err := r.checkTable(); err != nil {
		return nil, err
	}
	rows, err := r.db.Query("SELECT * FROM user_relation")
	if err != nil {
		return nil, err
	}
	relations := []ct.UserRelation{}
	for rows.Next() {
		rel := ct.UserRelation{}
		err = rows.Scan(&rel.ID, &rel.User1, &rel.User2, &rel.Encounters)
		if err != nil {
			return nil, err
		}
		relations = append(relations, rel)
	}

	rows.Close() //good habit to close

	return relations, nil
}

func (r *repo) checkTable() error {
	rows, err := r.db.Query("SELECT * FROM sqlite_master WHERE type='table';")
	if err != nil {
		return err
	}
	for rows.Next() {
		tableName := ""
		err = rows.Scan(&tableName)
		if err != nil {
			return err
		}
		if tableName == "user_relation" {
			return nil
		}
	}

	_, err = r.db.Exec(`
CREATE TABLE user_relation (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user1 VARCHAR(64) NOT NULL,
    user2 VARCHAR(64) NOT NULL,
    encounters INTEGER
)
	`)

	return err
}
func (r *repo) UpdateEncounters(rel ct.UserRelation) (err error) {
	if err := r.checkTable(); err != nil {
		return err
	}
	user1 := rel.User1
	user2 := rel.User2
	encounters := rel.Encounters

	tx, err := r.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		switch err {
		case nil:
			err = tx.Commit()
		default:
			tx.Rollback()
		}
	}()
	id := 0
	rows, err := tx.Query("SELECT id FROM user_relation WHERE ( user1=? AND user2=? ) OR ( user2=? AND user1=? )", user1, user2, user1, user2)
	if err != nil {
		return err
	}
	if rows.Next() {
		if err = rows.Scan(&id); err != nil {
			return
		}
	} else {
		if _, err = tx.Exec("INSERT INTO user_relation(user1, user2, encounters) values(?,?,?)", user1, user2, encounters); err != nil {
			return
		}
		return
	}

	if _, err = tx.Exec("UPDATE user_relation SET encounters=? WHERE id=?", encounters, id); err != nil {
		return
	}
	return
}
