package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/go-yaml/yaml"
	_ "github.com/mattn/go-sqlite3"
)

type ServerConfig struct {
	SlackToken   string `yaml:"slackToken"`
	SlackChannel string `yaml:"slackChannel"`
	DatabasePath string `yaml:"databasePath"`
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
func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
