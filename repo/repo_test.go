package repo

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	ct "github.com/mtyurt/coffeetable"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestNew(t *testing.T) {
	db := &sql.DB{}
	r := New(db)
	if r == nil {
		t.Fatalf("Repo instance is returned nil")
	}
	re, ok := r.(*repo)
	if !ok {
		t.Fatalf("New should return an instance of repo, but it is %v", reflect.TypeOf(re))
	}
	if re.db != db {
		t.Fatalf("field of repo does not match")
	}
}
func TestShouldGetUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := repo{db}
	rows := sqlmock.NewRows([]string{"id", "user1", "user2", "encounters"}).
		AddRow(1, "ali", "veli", 3).
		AddRow(2, "veli", "ahmet", 1)
	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table';").WillReturnRows(sqlmock.NewRows([]string{"table"}).AddRow("user_relation"))
	mock.ExpectQuery("SELECT [*] FROM user_relation").WillReturnRows(rows)
	relations, err := r.GetUserRelations()
	if err != nil {
		t.Fatal(err)
	}

	checkUser(t, ct.UserRelation{1, "ali", "veli", 3}, relations[0])
	checkUser(t, ct.UserRelation{2, "veli", "ahmet", 1}, relations[1])

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func checkUser(t *testing.T, expected ct.UserRelation, actual ct.UserRelation) {
	if actual.ID != expected.ID || actual.User1 != expected.User1 ||
		actual.User2 != expected.User2 || actual.Encounters != expected.Encounters {
		t.Fatalf("User does not match! expected: %v actual: %v", expected, actual)
	}
}
func TestCheckTableShouldCreateTableWhenAbsent(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := repo{db}

	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table';").WillReturnRows(sqlmock.NewRows([]string{"table"}))
	mock.ExpectExec(`CREATE TABLE user_relation .*`).WillReturnResult(sqlmock.NewResult(1, 1))
	if err = r.checkTable(); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
func TestIncreaseEncounterShouldSucceed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := repo{db}

	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table';").WillReturnRows(sqlmock.NewRows([]string{"table"}).AddRow("user_relation"))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM user_relation WHERE [(] user1=[?] AND user2=[?] [)] OR [(] user2=[?] AND user1=[?] [)]").WithArgs("ali", "veli", "ali", "veli").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO user_relation(.*)").WithArgs("ali", "veli", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	//second run
	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table';").WillReturnRows(sqlmock.NewRows([]string{"table"}).AddRow("user_relation"))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM user_relation WHERE [(] user1=[?] AND user2=[?] [)] OR [(] user2=[?] AND user1=[?] [)]").WithArgs("veli", "ali", "veli", "ali").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("UPDATE user_relation SET encounters=[?] WHERE id=[?]").WithArgs(3, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := r.UpdateEncounters(userRelation("ali", "veli", 1)); err != nil {
		t.Fatal(err)
	}

	if err = r.UpdateEncounters(userRelation("veli", "ali", 3)); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
func TestUpdateEncountersShouldRollbackWhenSqlFails(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := repo{db}

	mock.ExpectQuery("SELECT name FROM sqlite_master WHERE type='table';").WillReturnRows(sqlmock.NewRows([]string{"table"}).AddRow("user_relation"))
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM user_relation WHERE [(] user1=[?] AND user2=[?] [)] OR [(] user2=[?] AND user1=[?] [)]").WithArgs("ali", "veli", "ali", "veli").WillReturnError(errors.New("query failed"))
	mock.ExpectRollback()

	if err := r.UpdateEncounters(userRelation("ali", "veli", 1)); err == nil {
		t.Fatal("Query should fail")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
func userRelation(user1 string, user2 string, encounters int) ct.UserRelation {
	return ct.UserRelation{User1: user1, User2: user2, Encounters: encounters}
}
