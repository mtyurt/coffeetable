package repo

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	ct "github.com/mtyurt/coffeetable"
)

type Repo struct {
	db *sql.DB
}

func New(db *sql.DB) Repo {
	return Repo{db}
}
func (r *Repo) GetUserRelations() ([]ct.UserRelation, error) {
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

func (r *Repo) UpdateEncounters(user1 string, user2 string, encounters int) (err error) {
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
