package main

import (
	"encoding/json"
	"fmt"
	"github.com/charmbracelet/bubbles/list"
	"os"
	"path/filepath"
)

type branchType struct {
	T string `json:"type"`
	D string `json:"description"`
}

type board struct {
	Name string `json:"name"`
	D    string `json:"description"`
}

type config struct {
	BranchTypes []branchType `json:"branchTypes"`
	Boards      []board      `json:"boards"`
}

const configFile = ".git-commands.json"

var defaultBranchTypes = []list.Item{
	branchType{"FEAT", "A new feature"},
	branchType{"FIX", "A bug fix"},
	branchType{"IMPR", "An improvement to a feature or enhancement"},
	branchType{"OPS", "Changes to our CI configuration files and scripts"},
	branchType{"CHORE", "Updating grunt tasks etc; no production code change"},
}

var defaultJiraBoards = []list.Item{
	board{"EPD", "CPO team board"},
	board{"IB", "Interim billing board"},
}

func convertBranchTypes(branchTypes []branchType) []list.Item {
	items := []list.Item{}
	for _, branchType := range branchTypes {
		items = append(items, branchType)
	}
	if len(items) == 0 {
		return defaultBranchTypes
	}
	return items
}

func convertBoards(boards []board) []list.Item {
	items := []list.Item{}
	for _, board := range boards {
		items = append(items, board)
	}
	if len(items) == 0 {
		return nil
	}
	return items
}

func loadConfigFile(path string) ([]list.Item, []list.Item, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading config file: %w", err)
	}
	var c config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, nil, fmt.Errorf("error parsing config file: %w", err)
	}
	return convertBranchTypes(c.BranchTypes), convertBoards(c.Boards), nil
}

func loadConfig() ([]list.Item, []list.Item, error) {
	basePath, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting home dir: %w", err)
	}
	targetPath, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("error getting current dir: %w", err)
	}
	for {
		rel, _ := filepath.Rel(basePath, targetPath)
		if rel == "." {
			break
		}
		filePath := filepath.Join(targetPath, configFile)
		if _, err := os.Open(filePath); err == nil {
			fmt.Println("Found config file at", filePath)
			return loadConfigFile(filePath)
		}

		targetPath += "/.."
	}
	fmt.Println("No config file found, using default config")
	return defaultBranchTypes, defaultJiraBoards, nil
}
