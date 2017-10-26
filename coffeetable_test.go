package coffeetable

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jmcvetta/randutil"
)

func TestGenerateGroups(t *testing.T) {
	inputUsers := []User{slackUser("tarik"), slackUser("ali"), slackUser("veli")}
	inputRelations := []UserRelation{
		UserRelation{User1: "deli", User2: "ali", Encounters: 1},
		UserRelation{User1: "deli", User2: "veli", Encounters: 1},
		UserRelation{User1: "ali", User2: "veli", Encounters: 1},
	}

	groups, relations, err := GenerateGroups(inputRelations, inputUsers)
	if err != nil {
		t.Fatal("Error is not expected:", err)
	}
	if len(groups) != 1 && len(groups[0]) != len(inputUsers) {
		t.Fatalf("Group lengths is wrong, 1 group and %d users expected but it was %d groups and %d users", len(inputUsers), len(groups), len(groups[0]))
	}
	if len(relations) != 5 {
		t.Fatalf("Relations length is wrong, 5 relations is expected but it was %d relations: ", len(relations), relations)
	}

}

func TestGenerateGroupSizes(t *testing.T) {
	testTable := map[int][]int{
		1:  []int{1},
		2:  []int{2},
		3:  []int{3},
		4:  []int{4},
		5:  []int{5},
		6:  []int{3, 3},
		7:  []int{3, 4},
		8:  []int{4, 4},
		9:  []int{3, 3, 3},
		10: []int{3, 3, 4},
		11: []int{3, 4, 4},
		12: []int{3, 3, 3, 3},
		23: []int{3, 3, 3, 3, 3, 4, 4},
	}

	for input, expected := range testTable {
		actual := generateGroupSizes(input)
		if !testEq(actual, expected) {
			t.Errorf("Expected: %v but got: %v\n", expected, actual)
		}
	}
}
func TestCalculateWeightedChoices(t *testing.T) {
	users := []User{slackUser("ali"), slackUser("veli")}
	bs := slackUser("tarik")
	testTable := []struct {
		relations []UserRelation
		expected  []randutil.Choice
	}{
		{[]UserRelation{UserRelation{0, "ali", "tarik", 5}},
			[]randutil.Choice{randutil.Choice{1, "ali"}, randutil.Choice{6, "veli"}}},
		{[]UserRelation{UserRelation{0, "ali", "tarik", 3}, UserRelation{0, "tarik", "veli", 3}},
			[]randutil.Choice{randutil.Choice{1, "ali"}, randutil.Choice{1, "veli"}}},
		{[]UserRelation{UserRelation{0, "ali", "tarik", 1}, UserRelation{0, "tarik", "veli", 3}},
			[]randutil.Choice{randutil.Choice{3, "ali"}, randutil.Choice{1, "veli"}}},
		{[]UserRelation{}, []randutil.Choice{randutil.Choice{1, "ali"}, randutil.Choice{1, "veli"}}},
	}
	for _, test := range testTable {
		actual := calculateWeightedChoices(bs, users, test.relations)
		if len(actual) != len(test.expected) {
			t.Errorf("Expected: %v Actual: %v", test.expected, actual)
		}
		for i, c := range actual {
			if test.expected[i] != c {
				t.Errorf("In %v, expected: %v actual: %v", test.expected, test.expected[i], c)
			}
		}
	}
}
func TestCalculateRandomizedGroup(t *testing.T) {
	choices := []randutil.Choice{
		randutil.Choice{1, "ali"},
		randutil.Choice{1, "veli"},
		randutil.Choice{2, "deli"},
	}
	subgroup, err := calculateRandomizedGroup(choices, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(subgroup) != 2 {
		t.Fatalf("Wrong subgroup length: %d", len(subgroup))
	}
	if subgroup[0] == subgroup[1] {
		t.Fatalf("Group elements should be different! Group: %v", subgroup)
	}

	possibleNames := []string{"ali", "veli", "deli"}
	checkName := func(actual string) bool {
		for _, expected := range possibleNames {
			if expected == actual {
				return true
			}
		}
		return false
	}
	for _, actual := range subgroup {
		if !checkName(actual) {
			t.Fatalf("Group element is invalid: %s, should be one of: %v", actual, possibleNames)
		}
	}
}
func TestCalculateRandomizedGroupFailsIfChoiceContainsNonString(t *testing.T) {
	_, err := calculateRandomizedGroup([]randutil.Choice{randutil.Choice{1, 7}}, 2)
	if err == nil {
		t.Fatal("Error expected")
	}

}
func TestConvertNamesToUsersShouldSucceed(t *testing.T) {
	users := []User{slackUser("deli"), slackUser("ali"), slackUser("veli")}
	testTable := []struct {
		inputNames []string
		expected   []User
	}{
		{[]string{"ali", "veli"}, []User{slackUser("ali"), slackUser("veli")}},
		{[]string{"veli", "deli"}, []User{slackUser("veli"), slackUser("deli")}},
	}
	for _, test := range testTable {
		actual, err := convertNamesToUsers(users, test.inputNames)
		if err != nil {
			t.Fatalf("Failed with error %v", err)
		}
		if len(actual) != len(test.expected) {
			t.Fatalf("Expected: %v but was: %v", userNames(test.expected), userNames(actual))
		}
		for i, u := range actual {
			if u != test.expected[i] {
				t.Fatalf("At index %d expected: %v but was: ", i, test.expected[i], u)
			}
		}
	}
}

func TestConvertNamesToUsersShouldFailWhenNameIsNotInUsers(t *testing.T) {
	testTable := []struct {
		input    []string
		expected error
	}{
		{[]string{"ali"}, errors.New("Error! The user list does not have a user with name [ali]!")},
		{[]string{"veli", "ali"}, errors.New("Error! The user list does not have a user with name [veli]!")},
	}
	for _, test := range testTable {
		_, err := convertNamesToUsers([]User{}, test.input)
		if err == nil {
			t.Fatalf("Function should have failed!")
		}
		if err.Error() != test.expected.Error() {
			t.Fatalf("Expected error: %v but was: %v |", test.expected.Error(), err.Error())
		}
	}

}

func TestUpdateRelationsWithNewGroup(t *testing.T) {
	users := []User{slackUser("deli"), slackUser("ali"), slackUser("veli")}
	testTable := []struct {
		input    []UserRelation
		expected []UserRelation
	}{
		{[]UserRelation{}, []UserRelation{
			UserRelation{User1: "deli", User2: "ali", Encounters: 1},
			UserRelation{User1: "deli", User2: "veli", Encounters: 1},
			UserRelation{User1: "ali", User2: "veli", Encounters: 1},
		}},
		{[]UserRelation{
			UserRelation{User1: "deli", User2: "ali", Encounters: 1},
			UserRelation{User1: "deli", User2: "veli", Encounters: 1},
			UserRelation{User1: "ali", User2: "veli", Encounters: 1},
		}, []UserRelation{
			UserRelation{User1: "deli", User2: "ali", Encounters: 2},
			UserRelation{User1: "deli", User2: "veli", Encounters: 2},
			UserRelation{User1: "ali", User2: "veli", Encounters: 2},
		}},
		{[]UserRelation{
			UserRelation{User1: "tarik", User2: "veli", Encounters: 1},
		}, []UserRelation{
			UserRelation{User1: "deli", User2: "ali", Encounters: 1},
			UserRelation{User1: "deli", User2: "veli", Encounters: 1},
			UserRelation{User1: "ali", User2: "veli", Encounters: 1},
			UserRelation{User1: "tarik", User2: "veli", Encounters: 1},
		}},
		{[]UserRelation{
			UserRelation{User1: "deli", User2: "veli", Encounters: 1},
			UserRelation{User1: "ali", User2: "veli", Encounters: 1},
		}, []UserRelation{
			UserRelation{User1: "deli", User2: "ali", Encounters: 1},
			UserRelation{User1: "deli", User2: "veli", Encounters: 2},
			UserRelation{User1: "ali", User2: "veli", Encounters: 2},
		}},
	}
	for _, test := range testTable {
		actual := updateRelationsWithNewGroup(test.input, users)
		if len(actual) != len(test.expected) {
			t.Fatalf("Expected: %v but was: %v", test.expected, actual)
		}
		for i, r := range test.expected {
			if r != actual[i] {
				t.Errorf("Index %d, expected: %v but was: %v", i, r, actual[i])
			}
		}
	}
}
func TestUpdateRelationsShouldPanicWhenInputContainsDuplicateRelations(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	updateRelationsWithNewGroup([]UserRelation{
		UserRelation{User1: "deli", User2: "ali", Encounters: 1},
		UserRelation{User1: "ali", User2: "deli", Encounters: 1},
	}, []User{})

	t.Errorf("The code did not panic")
}
func TestDeleteGroupFromUsers(t *testing.T) {
	users := []User{slackUser("deli"), slackUser("ali"), slackUser("veli")}
	testTable := []struct {
		input    []User
		expected []User
	}{
		{[]User{slackUser("ali")}, []User{slackUser("deli"), slackUser("veli")}},
		{[]User{slackUser("tarik")}, []User{slackUser("deli"), slackUser("ali"), slackUser("veli")}},
		{[]User{slackUser("tarik"), slackUser("ali"), slackUser("deli")}, []User{slackUser("veli")}},
		{[]User{slackUser("deli"), slackUser("ali"), slackUser("veli")}, []User{}},
	}
	for _, test := range testTable {
		actual := deleteGroupFromUsers(users, test.input)
		if len(actual) != len(test.expected) {
			t.Fatalf("Expected: %v but was: %v", userNames(test.expected), userNames(actual))
		}
		for i, r := range test.expected {
			if r != actual[i] {
				t.Errorf("Index %d, expected: %v but was: %v", i, r.Name, actual[i].Name)
			}
		}
	}
}
func userNames(users []User) []string {
	names := make([]string, len(users))
	for i, u := range users {
		names[i] = u.Name
	}
	return names
}
func testEq(a, b []int) bool {

	if a == nil && b == nil {
		return true
	}

	if a == nil || b == nil {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}
func slackUser(name string) User {
	return User{Name: name}
}
