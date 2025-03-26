package main

import (
	"go-cache/internal/config"
	"go-cache/internal/git"
	"go-cache/internal/project"
	"log"
)

func main() {
	cmdArgs := config.InitRepoConfig()
	if err := cmdArgs.ConfigureFlags(); err != nil {
		log.Fatalf("[ERROR] Invalid arguments: %v", err)
	}
	if cmdArgs.History {
		git.AnalizeCommit(cmdArgs.Path)
	} else {
		project.AnalyzeProject(cmdArgs.Path)
	}
}
