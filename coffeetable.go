package coffeetable

import (
	"math/rand"

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
	groupSizes := generateGroupSizes(len(users))
	groups := make([][]slack.User, len(groupSizes))
	baseInd := 0
	for i, s := range groupSizes {
		groups[i] = make([]slack.User, s)
		for j, u := range users[baseInd : baseInd+s] {
			groups[i][j] = u
		}
		baseInd += s
	}
	return groups, nil
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
