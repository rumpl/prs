package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/fatih/color"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Repo represents a repo
type Repo struct {
	Owner string `json:"owner,omitempty"`
	Repo  string `json:"repo,omitempty"`
}

// Config rapresent a config
type Config struct {
	Login string `json:"login,omitempty"`
	Token string `json:"token,omitempty"`
	Repos []Repo `json:"repos,omitempty"`
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func loadConfig(path string) (Config, error) {
	jsonFile, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}
	defer jsonFile.Close()

	bytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return Config{}, err
	}

	var config Config
	if err := json.Unmarshal(bytes, &config); err != nil {
		return Config{}, err
	}

	return config, nil
}

func checkStatus(ctx context.Context, client *github.Client, login string, owner string, repo string) {
	pulls, _, err := client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		panic("Failed listing pull requests")
	}

	for _, v := range pulls {
		if *v.User.Login == login {
			status, _, err := client.Repositories.GetCombinedStatus(ctx, owner, repo, *v.Head.SHA, nil)
			if err != nil {
				panic("Failed getting the status of a pull request")
			}

			fmt.Printf("%s #%d: ", *v.Base.Repo.FullName, *v.Number)
			switch *status.State {
			case "pending":
				color.Set(color.FgBlue)
			case "success":
				color.Set(color.FgGreen)
			case "failure":
				color.Set(color.FgRed)
			}
			fmt.Println(*status.State)
			color.Set(color.Reset)
		}
	}
}

func main() {
	config, err := loadConfig("./config.json")
	if err != nil {
		panic("Cannot read config file")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	for _, repo := range config.Repos {
		checkStatus(ctx, client, config.Login, repo.Owner, repo.Repo)
	}
}
