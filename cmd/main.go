package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"

	ct "github.com/mtyurt/coffeetable"

	"github.com/go-yaml/yaml"
	_ "github.com/mattn/go-sqlite3"
	"github.com/mtyurt/coffeetable/repo"
	"github.com/mtyurt/coffeetable/slackhelper"
	"github.com/nlopes/slack"
)

type ServerConfig struct {
	SlackToken     string `yaml:"slackToken"`
	SlackChannel   string `yaml:"slackChannel"`
	PrivateChannel bool   `yaml:"privateChannel"`
	DatabasePath   string `yaml:"databasePath"`
}

var slackApi *slack.Client

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
	panicOnErr(err)
	defer db.Close()
	slackService := slackhelper.New(conf.SlackToken, conf.SlackChannel, conf.PrivateChannel)
	members, err := slackService.GetChannelMembers()
	panicOnErr(err)
	fmt.Println("Channel member count:", len(members))
	printMembers(members)
	repo := repo.New(db)
	relations, err := repo.GetUserRelations()
	panicOnErr(err)
	groups, relations, err := ct.GenerateGroups(relations, members)
	printGroups(groups)
	panicOnErr(err)
	for _, r := range relations {
		err := repo.UpdateEncounters(r)
		panicOnErr(err)
	}
	err = slackService.PublishGroupsInSlack(groups)
	panicOnErr(err)

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
func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
func printMembers(members []ct.User) {
	for _, u := range members {
		fmt.Printf("%s %15s\n", u.ID, u.Name)
	}
}
func printGroups(groups [][]ct.User) {
	fmt.Println("Groups I have found:")
	for i, g := range groups {
		fmt.Printf("Group %d:\n", i)
		printMembers(g)
		fmt.Println()
	}
}
