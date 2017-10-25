package coffeetable

import (
	"errors"
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

func GenerateGroups(relations []UserRelation, users []slack.User) ([][]slack.User, []UserRelation, error) {
	users = shuffleUsers(users)
	printUsers(users)
	groupSizes := generateGroupSizes(len(users))
	groups := make([][]slack.User, len(groupSizes))
	for i, s := range groupSizes {
		groups[i] = make([]slack.User, s)
		baseUser := users[0]
		wc := calculateWeightedChoices(baseUser, users[1:], relations)
		chosenNames, err := calculateRandomizedGroup(wc, s-1)
		if err != nil {
			return nil, nil, err
		}
		chosenUsers, err := convertNamesToUsers(users, chosenNames)
		if err != nil {
			return nil, nil, err
		}
		groups[i] = append(chosenUsers, baseUser)
		relations = updateRelationsWithNewGroup(relations, groups[i])
		users = deleteGroupFromUsers(users, groups[i])
	}
	return groups, relations, nil
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
func calculateRandomizedGroup(weightedChoices []randutil.Choice, size int) ([]string, error) {
	names := make([]string, size)
	for i := 0; i < size; i++ {
		choice, err := randutil.WeightedChoice(weightedChoices)
		if err != nil {
			return nil, err
		}
		weightedChoices = removeChoice(weightedChoices, choice)
		item := choice.Item
		if n, ok := item.(string); ok {
			names[i] = n
		} else {
			panic(fmt.Sprintf("Choice item is not string! It's: %v", choice))
		}
	}
	return names, nil
}
func removeChoice(choices []randutil.Choice, tbd randutil.Choice) []randutil.Choice {
	for i, c := range choices {
		if c.Item == tbd.Item {
			return append(choices[:i], choices[i+1:]...)
		}
	}
	return choices
}
func convertNamesToUsers(users []slack.User, names []string) ([]slack.User, error) {
	userMap := make(map[string]slack.User)
	for _, u := range users {
		userMap[u.Name] = u
	}
	subgroup := make([]slack.User, len(names))
	for i, n := range names {
		u, ok := userMap[n]
		if !ok {
			return nil, errors.New(fmt.Sprintf("Error! The user list does not have a user with name [%s]!", n))
		}
		subgroup[i] = u
	}
	return subgroup, nil
}
func updateRelationsWithNewGroup(relations []UserRelation, group []slack.User) []UserRelation {
	relMap := make(map[string]UserRelation)
	remainingRelations := make(map[string]UserRelation)
	for _, r := range relations {
		remainingRelations[r.User1+"|"+r.User2] = r
		relMap[r.User1+"|"+r.User2] = r
		relMap[r.User2+"|"+r.User1] = r
	}
	groupSize := len(group)
	newRels := []UserRelation{}
	for i := 0; i < groupSize-1; i++ {
		for j := i + 1; j < groupSize; j++ {
			u1 := group[i]
			u2 := group[j]
			rel, ok := relMap[u1.Name+"|"+u2.Name]
			if !ok {
				rel = UserRelation{User1: u1.Name, User2: u2.Name, Encounters: 0}
			} else {
				delete(remainingRelations, rel.User1+"|"+rel.User2)
			}
			rel.Encounters++
			newRels = append(newRels, rel)
		}
	}
	for _, rel := range remainingRelations {
		newRels = append(newRels, rel)
	}
	return newRels
}
func deleteGroupFromUsers(from []slack.User, tbd []slack.User) []slack.User {
	users := make([]slack.User, len(from))
	for i, f := range from {
		users[i] = f
	}
	for _, t := range tbd {
	inner:
		for i, f := range users {
			if f.Name == t.Name {
				users = append(users[:i], users[i+1:]...)[:len(users)-1]
				break inner
			}
		}
	}

	return users
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
