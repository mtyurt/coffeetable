package slackhelper

import (
	"fmt"
	"strings"
	"sync"

	ct "github.com/mtyurt/coffeetable"

	"github.com/nlopes/slack"
)

type SlackHelper interface {
	GetChannelMembers() ([]ct.User, error)
	PublishGroupsInSlack(gorups [][]ct.User) error
}

type slackService struct {
	token       string
	channel     string
	isPrivate   bool
	apiProvider func(token string) slackAdapter
}

func New(token string, channel string, isPrivate bool) SlackHelper {
	return &slackService{token, channel, isPrivate, func(t string) slackAdapter {
		return &realSlackAdapter{slack.New(t)}
	}}
}
func (service *slackService) GetChannelMembers() (members []ct.User, err error) {
	slackApi := service.apiProvider(service.token)
	var ids []string
	if !service.isPrivate {
		ids, err = slackApi.GetChannelMembers(service.channel)
		if err != nil {
			return nil, err
		}
	} else {
		ids, err = slackApi.GetGroupMembers(service.channel)
		if err != nil {
			return nil, err
		}
	}
	members = []ct.User{}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(ids))
	for _, id := range ids {
		//9 seconds improvement
		go func(id string) {
			userinfo, err := slackApi.GetUserInfo(id)
			if err != nil {
				panic(err)
			}
			if !userinfo.Deleted && !userinfo.IsBot {
				mu.Lock()
				members = append(members, ct.User(*userinfo))
				mu.Unlock()
			}
			wg.Done()
		}(id)
	}
	wg.Wait()
	return
}

func (service *slackService) PublishGroupsInSlack(groups [][]ct.User) error {
	slackApi := service.apiProvider(service.token)
	text := ""
	for i, group := range groups {
		ids := make([]string, len(group))
		for j, u := range group {
			ids[j] = fmt.Sprintf("<@%s>", u.ID)
		}
		text += fmt.Sprintf("*Group %d:* %v\n", i+1, strings.Join(ids, ", "))
	}
	params := slack.PostMessageParameters{
		AsUser: true,
	}
	_, _, err := slackApi.PostMessage(service.channel, fmt.Sprintf("Coffee time! Today's groups: \n%s\nZoom up!", text), params)
	return err
}
