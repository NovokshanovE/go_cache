package config

import (
	"os"

	flag "github.com/spf13/pflag"
)

type RepoConfig struct {
	Path    string
	History bool
}

func InitRepoConfig() *RepoConfig {
	return &RepoConfig{}
}

func (rc *RepoConfig) ConfigureFlags() error {
	flag.StringVar(&rc.Path, "path", "./", "Path to the repository.")
	flag.BoolVar(&rc.History, "history", false, "Analize git commits.")

	flag.Parse()

	if err := flag.CommandLine.Parse(os.Args[1:]); err != nil {
		return err
	}

	return nil
}
