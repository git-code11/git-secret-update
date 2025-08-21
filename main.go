package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var repoID string
var ownerID string
var accessKey string
var envKey string
var envFile string
var envValue string
var isSecret bool
var defaultKey = "ENV_CONTENT"

func init() {
	flag.StringVar(&repoID, "repo", os.Getenv("GITHUB_ID"), "Owner/Repository ID")
	flag.StringVar(&accessKey, "key", os.Getenv("GITHUB_TOKEN"), "Access Token")
	flag.StringVar(&envKey, "env-key", defaultKey, "Env Key")
	flag.StringVar(&envValue, "env-value", "", "Env Value")
	flag.StringVar(&envFile, "env-file", "", "Env File")
	flag.BoolVar(&isSecret, "secret", false, "Secret Env")
}

func promptOnEmpty(prompt string, val *string) error {
	if *val == "" {
		fmt.Printf("%s: ", prompt)
		_, err := fmt.Scanln(val)
		return err
	}
	return nil
}

func Parse() {
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

	fmt.Printf("Repository: %s EnvKey: %s\n", repoID, envKey)

	// Check EnvValue from local environment
	if envValue == "" {
		envValue = os.Getenv(envKey)
	}

	// Check EnvValue from .env
	if envValue == "" && envFile != "" {
		out, err := os.ReadFile(envFile)
		if err != nil {
			log.Fatal(err)
		}
		envValue = string(out)
	}

	// Read Env from terminal
	if envValue == "" {
		fmt.Println("Paste Env Value here: Use <CTRL-D> to close")
		out, err := io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal(err)
		}
		envValue = string(out)
	}
}

func main() {
	Parse()

	client := NewUpdateGithubEnv(accessKey, ownerID, repoID)
	var err error
	if isSecret {
		fmt.Println("Updating Environ Secret...")
		err = client.UpdateSecret(envKey, envValue)
	} else {
		fmt.Println("Updating Environ Variable...")
		err = client.UpdateVariable(envKey, envValue)
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated Value with: \n %s: %s\n", envKey, envValue)
}
