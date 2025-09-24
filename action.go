package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"slices"

	crypto_rand "crypto/rand"

	"github.com/google/go-github/v74/github"
	"golang.org/x/crypto/nacl/box"
)

type UpdateGithubEnv struct {
	client *github.Client
	repoPk *github.PublicKey
	repo   string
	owner  string
}

func NewUpdateGithubEnv(accessKey string, owner string, repo string) *UpdateGithubEnv {
	client := github.NewClient(nil).WithAuthToken(accessKey)
	return &UpdateGithubEnv{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

func (u *UpdateGithubEnv) UpdateSecret(key string, value string) error {
	ctx := context.Background()
	encryptedValue, err := u.EncryptUsingRepoPk(value)
	if err != nil {
		return err
	}
	v := github.EncryptedSecret{
		Name:           key,
		KeyID:          u.repoPk.GetKeyID(),
		EncryptedValue: encryptedValue,
	}
	_, err = u.client.Actions.CreateOrUpdateRepoSecret(ctx, u.owner, u.repo, &v)
	return err
}

func (u *UpdateGithubEnv) GetRepoPk() error {
	ctx := context.Background()
	pk, _, err := u.client.Actions.GetRepoPublicKey(ctx, u.owner, u.repo)
	if err == nil {
		u.repoPk = pk
	}
	return err
}

func (u *UpdateGithubEnv) EncryptUsingRepoPk(value string) (string, error) {
	var encryptedString string

	if u.repoPk == nil {
		if err := u.GetRepoPk(); err != nil {
			return encryptedString, err
		}
	}

	decodedPublicKey, err := base64.StdEncoding.DecodeString(u.repoPk.GetKey())
	if err != nil {
		return encryptedString, fmt.Errorf("base64.StdEncoding.DecodeString was unable to decode public key: %v", err)
	}

	var boxKey [32]byte
	copy(boxKey[:], decodedPublicKey)
	encryptedBytes, err := box.SealAnonymous([]byte{}, []byte(value), &boxKey, crypto_rand.Reader)
	if err != nil {
		return encryptedString, fmt.Errorf("box.SealAnonymous failed with error %w", err)
	}

	encryptedString = base64.StdEncoding.EncodeToString(encryptedBytes)
	return encryptedString, nil
}

func (u *UpdateGithubEnv) UpdateVariable(key string, value string) error {
	ctx := context.Background()
	exists, err := u.HasVariable(key)
	if err != nil {
		return err
	}
	if exists {
		_, err := u.client.Actions.UpdateRepoVariable(ctx, u.owner, u.repo, &github.ActionsVariable{
			Name:  key,
			Value: value,
		})
		return err
	}
	_, err = u.client.Actions.CreateRepoVariable(ctx, u.owner, u.repo, &github.ActionsVariable{
		Name:  key,
		Value: value,
	})
	return err
}

func (u *UpdateGithubEnv) GetVariable(key string) (string, error) {
	ctx := context.Background()
	result, _, err := u.client.Actions.GetRepoVariable(ctx, u.owner, u.repo, key)
	return result.Value, err
}

func (u *UpdateGithubEnv) HasVariable(key string) (bool, error) {
	var exists bool
	ctx := context.Background()
	result, _, err := u.client.Actions.ListRepoVariables(ctx, u.owner, u.repo, nil)
	if err != nil {
		return exists, err
	}

	exists = slices.ContainsFunc(result.Variables, func(a *github.ActionsVariable) bool {
		return a.Name == key
	})

	return exists, nil
}
