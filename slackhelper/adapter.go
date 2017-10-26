package slackhelper

import "github.com/nlopes/slack"

type slackAdapter interface {
	GetChannelMembers(channel string) ([]string, error)
	GetGroupMembers(group string) ([]string, error)
	GetUserInfo(user string) (*slack.User, error)
	PostMessage(channel string, text string, params slack.PostMessageParameters) (string, string, error)
}

type realSlackAdapter struct {
	api *slack.Client
}

func (r *realSlackAdapter) GetChannelMembers(channel string) ([]string, error) {
	info, err := r.api.GetChannelInfo(channel)
	if err != nil {
		return nil, err
	}
	return info.Members, nil
}

func (r *realSlackAdapter) GetGroupMembers(group string) ([]string, error) {
	info, err := r.api.GetGroupInfo(group)
	if err != nil {
		return nil, err
	}
	return info.Members, nil
}

func (r *realSlackAdapter) GetUserInfo(user string) (*slack.User, error) {
	return r.api.GetUserInfo(user)
}

func (r *realSlackAdapter) PostMessage(channel string, text string, params slack.PostMessageParameters) (string, string, error) {
	return r.api.PostMessage(channel, text, params)
}
