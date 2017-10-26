package slackhelper

import (
	"fmt"
	"reflect"
	"testing"

	ct "github.com/mtyurt/coffeetable"
	"github.com/nlopes/slack"
)

func TestNew(t *testing.T) {
	tableTest := []struct {
		token   string
		channel string
		private bool
	}{
		{"token", "channel", true},
		{"123token", "chaadfnnel", false},
	}
	for i, test := range tableTest {
		actual := New(test.token, test.channel, test.private)
		s, ok := actual.(*slackService)
		if !ok {
			t.Fatalf("Test %d, Expected service type is: slackService but it was:%v", i+1, reflect.TypeOf(actual))
		}
		if s.token != test.token || s.channel != test.channel || s.isPrivate != test.private {
			t.Fatalf("Test %d, Expected service: %v but was: %v", i+1, test, s)
		}
	}

}

func TestGetChannelMembers(t *testing.T) {
	names := []string{"ali", "veli"}
	mock := &mockSlack{
		getChannelMembers: func(channel string) ([]string, error) {
			panic("shouldn't have called")
		},
		getGroupMembers: func(group string) ([]string, error) {
			return names, nil
		},
		getUserInfo: func(user string) (*slack.User, error) {
			switch user {
			case "ali":
				return &slack.User{Name: "ali"}, nil
			case "veli":
				return &slack.User{Name: "veli"}, nil
			default:
				panic(user + " is not valid")
			}
		},
	}
	slackService := &slackService{"token", "channel", true, func(token string) slackAdapter {
		return mock
	}}

	members, err := slackService.GetChannelMembers()
	if err != nil {
		t.Fatal(err)
	}
	for i, u := range members {
		if u.Name != names[i] {
			t.Errorf("%d element expected %s but was %s", i, u.Name, names[i])
		}
	}

	// test non-private

	mock.getChannelMembers = func(channel string) ([]string, error) {
		return names, nil
	}
	mock.getGroupMembers = func(group string) ([]string, error) {
		panic("shouldn't have called")
	}
	slackService.isPrivate = false
	members, err = slackService.GetChannelMembers()
	if err != nil {
		t.Fatal(err)
	}
	for i, u := range members {
		if u.Name != names[i] {
			t.Errorf("%d element expected %s but was %s", i, u.Name, names[i])
		}
	}
}
func TestPublishGroupsInSlack(t *testing.T) {
	var inputChannel, inputText string
	mock := &mockSlack{
		postMessage: func(channel string, text string, params slack.PostMessageParameters) (string, string, error) {
			inputChannel = channel
			inputText = text
			return "", "", nil
		},
	}
	slackService := &slackService{"token", "mychannel", true, func(token string) slackAdapter {
		return mock
	}}
	err := slackService.PublishGroupsInSlack([][]ct.User{
		[]ct.User{ct.User{ID: "ali"}, ct.User{ID: "veli"}},
	})

	if err != nil {
		t.Fatal(err)
	}
	if inputChannel != "mychannel" {
		t.Fatalf("Channel is expected: mychannel but was: %s", inputChannel)
	}
	expectedText := fmt.Sprintf("Coffee time! Today's groups: \n*Group 1:* <@ali>, <@veli>\n\nZoom up!")
	if inputText != expectedText {
		t.Fatalf("Text is exptected to be: %s but was: %s", expectedText, inputText)
	}
}

type mockSlack struct {
	getChannelMembers func(channel string) ([]string, error)
	getGroupMembers   func(group string) ([]string, error)
	getUserInfo       func(user string) (*slack.User, error)
	postMessage       func(channel string, text string, params slack.PostMessageParameters) (string, string, error)
}

func (m *mockSlack) GetChannelMembers(channel string) ([]string, error) {
	return m.getChannelMembers(channel)
}

func (m *mockSlack) GetGroupMembers(group string) ([]string, error) {
	return m.getGroupMembers(group)
}

func (m *mockSlack) GetUserInfo(user string) (*slack.User, error) {
	return m.getUserInfo(user)
}

func (m *mockSlack) PostMessage(channel string, text string, params slack.PostMessageParameters) (string, string, error) {
	return m.postMessage(channel, text, params)
}
