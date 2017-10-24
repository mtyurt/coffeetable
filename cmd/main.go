package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"sync"

	"github.com/go-yaml/yaml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nlopes/slack"
)

type ServerConfig struct {
	SlackToken     string `yaml:"slackToken"`
	SlackChannel   string `yaml:"slackChannel"`
	PrivateChannel bool   `yaml:"privateChannel"`
	DatabasePath   string `yaml:"databasePath"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error! Usage: coffeetable <conf-file-path>")
		os.Exit(1)
	}
	conf, err := readConfig(os.Args[1])
	if err != nil {
		fmt.Println("Error while reading conf file:", err)
		os.Exit(1)
	}
	db, err := sql.Open("sqlite3", conf.DatabasePath)
	checkErr(err)
	defer db.Close()
	members, err := GetChannelMembers(conf)
	checkErr(err)
	for _, m := range members {
		fmt.Printf("%s %25s %s\n", m.ID, m.RealName, m.Name)
	}
	//	repo := repo.New(db)
}
func readConfig(filePath string) (conf *ServerConfig, err error) {
	confContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	conf = &ServerConfig{}
	err = yaml.Unmarshal(confContent, conf)
	if err != nil {
		return
	}
	return
}
func GetChannelMembers(conf *ServerConfig) (members []slack.User, err error) {
	api := slack.New(conf.SlackToken)
	var ids []string
	if !conf.PrivateChannel {
		chInfo, err := api.GetChannelInfo(conf.SlackChannel)
		if err != nil {
			return nil, err
		}
		ids = chInfo.Members
	} else {
		chInfo, err := api.GetGroupInfo(conf.SlackChannel)
		if err != nil {
			return nil, err
		}
		ids = chInfo.Members
	}
	members = []slack.User{}
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	wg.Add(len(ids))
	for _, id := range ids {
		//9 seconds improvement
		go func(id string) {
			userinfo, err := api.GetUserInfo(id)
			if err != nil {
				panic(err)
			}
			if !userinfo.Deleted && !userinfo.IsBot {
				mu.Lock()
				members = append(members, *userinfo)
				mu.Unlock()
			}
			wg.Done()
		}(id)
	}
	wg.Wait()
	return
}
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
