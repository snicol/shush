package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path"

	"github.com/Jeffail/gabs"
	"github.com/go-git/go-git/v5"
)

type JSONGit struct {
	Path       string
	Filename   string
	RemoteName string
	Indent     string
}

type JSONSecret struct {
	Value   string `json:"value"`
	Version int    `json:"v"`
}

func NewJSONGit(path, filename, remoteName, indent string) *JSONGit {
	return &JSONGit{
		Path:       path,
		Filename:   filename,
		RemoteName: remoteName,
		Indent:     indent,
	}
}

func (jg *JSONGit) Get(ctx context.Context, ks []string) (out []Result, err error) {
	r, err := git.PlainOpen(jg.Path)
	if err != nil {
		return nil, err
	}

	w, err := r.Worktree()
	if err != nil {
		return nil, err
	}

	err = w.PullContext(ctx, &git.PullOptions{
		RemoteName: jg.RemoteName,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		return nil, err
	}

	parsed, err := jg.parseFile()
	if err != nil {
		return nil, err
	}

	for _, k := range ks {
		value, err := extractResult(k, parsed)
		if err != nil {
			return nil, err
		}

		out = append(out, Result{
			Value:   value.Value,
			Version: value.Version,
		})
	}

	return
}

func (jg *JSONGit) Set(ctx context.Context, k, v string) error {
	parsed, err := jg.parseFile()
	if err != nil {
		return err
	}

	value, err := extractResult(k, parsed)
	if err != nil {
		return err
	}

	value.Value = v
	value.Version++

	_, err = parsed.SetP(value, k)
	if err != nil {
		return err
	}

	parsedJSON := parsed.StringIndent("", jg.Indent)

	err = ioutil.WriteFile(jg.getJSONPath(), []byte(parsedJSON), 0644)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(jg.Path)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return fmt.Errorf("failed to load working tree: %w", err)
	}

	_, err = w.Add(jg.Filename)
	if err != nil {
		return fmt.Errorf("failed to add file to path: %w", err)
	}

	_, err = w.Commit(fmt.Sprintf("set %s", k), &git.CommitOptions{})
	if err != nil {
		return fmt.Errorf("failed to commit file: %w", err)
	}

	err = r.PushContext(ctx, &git.PushOptions{
		RemoteName: jg.RemoteName,
	})
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

func (jg *JSONGit) LatestVersion(ctx context.Context, k string) (int, error) {
	vs, err := jg.Get(ctx, []string{k})
	if err != nil {
		return 0, err
	}

	if len(vs) == 0 {
		return 0, errors.New("not found")
	}

	return vs[0].Version, nil
}

func (jg *JSONGit) parseFile() (*gabs.Container, error) {
	content, err := ioutil.ReadFile(jg.getJSONPath())
	if err != nil {
		return nil, err
	}

	return gabs.ParseJSON(content)
}

func (jg *JSONGit) getJSONPath() string {
	return path.Join(jg.Path, jg.Filename)
}

func extractResult(key string, parsed *gabs.Container) (value *JSONSecret, err error) {
	if !parsed.ExistsP(key) {
		return &JSONSecret{
			Version: 0,
		}, nil
	}

	err = json.Unmarshal(parsed.Path(key).Bytes(), &value)
	return
}
