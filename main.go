package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sync"

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

func checkStatus(ctx context.Context, client *github.Client, login string, owner string, repo string, bitBar bool) {
	pulls, _, err := client.PullRequests.List(ctx, owner, repo, nil)
	if err != nil {
		panic("Failed listing pull requests")
	}
	var wg sync.WaitGroup
	wg.Add(len(pulls))

	for _, v := range pulls {
		go func(v *github.PullRequest) {
			defer wg.Done()
			if *v.User.Login == login {
				status, _, err := client.Repositories.GetCombinedStatus(ctx, owner, repo, *v.Head.SHA, nil)
				if err != nil {
					panic("Failed getting the status of a pull request")
				}

				var fgColor string
				switch *status.State {
				case "pending":
					fgColor = "blue"
					color.Set(color.FgBlue)
				case "success":
					fgColor = "green"
					color.Set(color.FgGreen)
				case "failure":
					fgColor = "red"
					color.Set(color.FgRed)
				}
				fmt.Printf("%s #%d", *v.Base.Repo.FullName, *v.Number)
				if bitBar {
					fmt.Printf("%s|color=%s", *status.State, fgColor)
				}
				fmt.Println()
				color.Set(color.Reset)
			}
		}(v)
	}
	wg.Wait()
}

func main() {
	var bitBar = flag.Bool("bit-bar", false, "")
	flag.Parse()
	config, err := loadConfig(path.Join(os.Getenv("HOME"), "./config.json"))
	if err != nil {
		panic("Cannot read config file")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	fmt.Println("PRs")
	for _, repo := range config.Repos {
		checkStatus(ctx, client, config.Login, repo.Owner, repo.Repo, *bitBar)
	}
}
