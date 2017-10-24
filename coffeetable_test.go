package coffeetable

import (
	"testing"

	"github.com/jmcvetta/randutil"
	"github.com/nlopes/slack"
)

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
	}

	for input, expected := range testTable {
		actual := generateGroupSizes(input)
		if !testEq(actual, expected) {
			t.Errorf("Expected: %v but got: %v\n", expected, actual)
		}
	}
}
func TestCalculateWeightedChoices(t *testing.T) {
	users := []slack.User{slack.User{Name: "ali"}, slack.User{Name: "veli"}}
	bs := slack.User{Name: "tarik"}
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
