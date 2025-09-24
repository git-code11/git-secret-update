package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

type State struct {
	key    string
	value  []byte
	file   []byte
	secret bool
}

type githubStateUpdateOpts struct {
	RepoID    string
	OwnerID   string
	AccessKey string
	States    []State
}

type GithubStateUpdate struct {
	client *UpdateGithubEnv
	states []State
}

func NewGithubStateUpdate(opts *githubStateUpdateOpts) *GithubStateUpdate {
	client := NewUpdateGithubEnv(opts.AccessKey, opts.OwnerID, opts.RepoID)
	return &GithubStateUpdate{
		client: client,
		states: opts.States,
	}
}

func (g *GithubStateUpdate) Execute() error {
	for _, state := range g.states {
		var value string
		if state.value != nil {
			value = string(state.value)
		} else if state.file != nil {
			out, err := os.ReadFile(string(state.file))
			if err != nil {
				return err
			}
			value = string(out)
		}

		if state.secret {
			fmt.Printf("Updating Environ Secret (%s)...\n", state.key)
			if err := g.client.UpdateSecret(state.key, value); err != nil {
				return err
			}
		} else {
			fmt.Printf("Updating Environ Variable (%s)...\n", state.key)
			if err := g.client.UpdateVariable(state.key, value); err != nil {
				return err
			}
		}
	}
	return nil
}

var repoID string
var ownerID string
var accessKey string
var stateFile string
var states []State

func init() {
	flag.StringVar(&repoID, "repo", os.Getenv("GITHUB_ID"), "Owner/Repository ID")
	flag.StringVar(&accessKey, "key", os.Getenv("GITHUB_TOKEN"), "Access Token")
	// TODO: Add Structure of state file
	flag.StringVar(&stateFile, "file", "", "State File")
}

func promptOnEmpty(prompt string, val *string) error {
	if *val == "" {
		fmt.Printf("%s: ", prompt)
		_, err := fmt.Scanln(val)
		return err
	}
	return nil
}

func parse() {
	var err error
	defer func() {
		if err != nil {
			log.Fatalf("Failed to parse: %+v", err)
		}
	}()

	flag.Parse()

	// AccessToken
	err = promptOnEmpty("Enter AccessKey", &accessKey)
	if err != nil {
		return
	}

	// Owner/Repo ID
	err = promptOnEmpty("Enter Owner/RepoID", &repoID)
	if err != nil {
		return
	}

	ids := strings.Split(repoID, "/")
	if len(ids) != 2 {
		err = errors.New("provide OWNER/REPO")
		return
	}

	ownerID = ids[0]
	repoID = ids[1]

	fmt.Printf("Repository: %s\n", repoID)

	// Get State content
	var out []byte
	if stateFile == "" {
		fmt.Println("Paste Env Value here: Use <CTRL-D> to close")
		out, err = io.ReadAll(os.Stdin)
	} else {
		out, err = os.ReadFile(stateFile)
	}

	if err != nil {
		return
	}

	if err = json.Unmarshal(out, &states); err != nil {
		return
	}
}

func main() {
	parse()

	client := NewGithubStateUpdate(&githubStateUpdateOpts{
		RepoID:    repoID,
		OwnerID:   ownerID,
		AccessKey: accessKey,
		States:    states,
	})

	if err := client.Execute(); err != nil {
		log.Fatal(err)
	}
}
