package repo

import (
	"testing"

	ct "github.com/mtyurt/coffeetable"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestShouldGetUsers(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := Repo{db}
	rows := sqlmock.NewRows([]string{"id", "user1", "user2", "encounters"}).
		AddRow(1, "ali", "veli", 3).
		AddRow(2, "veli", "ahmet", 1)

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

func TestIncreaseEncounterShouldSucceed(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	r := Repo{db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM user_relation WHERE [(] user1=[?] AND user2=[?] [)] OR [(] user2=[?] AND user1=[?] [)]").WithArgs("ali", "veli", "ali", "veli").WillReturnRows(sqlmock.NewRows([]string{"id"}))
	mock.ExpectExec("INSERT INTO user_relation(.*)").WithArgs("ali", "veli", 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()
	//second run
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM user_relation WHERE [(] user1=[?] AND user2=[?] [)] OR [(] user2=[?] AND user1=[?] [)]").WithArgs("veli", "ali", "veli", "ali").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("UPDATE user_relation SET encounters=[?] WHERE id=[?]").WithArgs(3, 1).WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	if err := r.UpdateEncounters("ali", "veli", 1); err != nil {
		t.Fatal(err)
	}

	if err = r.UpdateEncounters("veli", "ali", 3); err != nil {
		t.Fatal(err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
