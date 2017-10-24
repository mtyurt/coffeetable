package coffeetable

import (
	"fmt"
	"math/rand"

	"github.com/jmcvetta/randutil"
	"github.com/nlopes/slack"
)

type UserRelation struct {
	ID         int
	User1      string
	User2      string
	Encounters int
}

func GenerateGroups(relations []UserRelation, users []slack.User) ([][]slack.User, []UserRelation) {
	users = shuffleUsers(users)
	printUsers(users)
	groupSizes := generateGroupSizes(len(users))
	groups := make([][]slack.User, len(groupSizes))
	for i, s := range groupSizes {
		groups[i] = make([]slack.User, s)
		baseUser := users[0]
		wc := calculateWeightedChoices(baseUser, users[1:], relations)
		groups[i] = calculateRandomizedGroup(wc, 3)
		relations = updateRelationsWithNewGroup(relations, groups[i])
		users = deleteUsers(users, groups[i])
	}
	return groups, relations
}
func calculateWeightedChoices(baseUser slack.User, users []slack.User, relations []UserRelation) []randutil.Choice {
	choices := make([]randutil.Choice, len(users))
	relMap := make(map[string]int)
	for _, r := range relations {
		relMap[r.User1+"|"+r.User2] = r.Encounters
		relMap[r.User2+"|"+r.User1] = r.Encounters
	}
	maxEncounter := 0
	for i, u := range users {
		e, ok := relMap[u.Name+"|"+baseUser.Name]
		if !ok {
			e = 0
		}
		choices[i] = randutil.Choice{e, u.Name}
		if e > maxEncounter {
			maxEncounter = e
		}
	}
	maxEncounter++
	for i, c := range choices {
		choices[i] = randutil.Choice{maxEncounter - c.Weight, c.Item}
	}
	return choices
}
func calculateRandomizedGroup(weightedChoices []randutil.Choice, size int) []slack.User {
	return nil
}
func deleteUsers(from []slack.User, tbd []slack.User) []slack.User {
	return nil
}
func updateRelationsWithNewGroup(relations []UserRelation, group []slack.User) []UserRelation {
	return nil
}
func printUsers(users []slack.User) {
	names := make([]string, len(users))
	for i, m := range users {
		names[i] = m.Name
	}
	fmt.Println(names)
}
func generateGroupSizes(len int) []int {
	if len <= 3 {
		return []int{len}
	}
	remLen := len
	groupSizes := make([]int, remLen/3+1)
	i := 0
	for remLen > 0 {
		remLen = remLen - 3
		if remLen >= 0 {
			groupSizes[i] = 3
			i++
		}
	}
	switch remLen {
	case -2:
		groupSizes[i-1] = 4
	case -1:
		if i > 1 {
			groupSizes[i-1] = 4
			groupSizes[i-2] = 4
		} else {
			groupSizes[i-1] = 5
		}
	}
	return groupSizes[:i]
}

func shuffleUsers(src []slack.User) []slack.User {
	dest := make([]slack.User, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	return dest
}
