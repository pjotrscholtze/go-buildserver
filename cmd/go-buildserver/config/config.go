package config

import (
	"io/ioutil"
	"log"

	"github.com/ghodss/yaml"
)

type Trigger struct {
	Kind     string
	Schedule string
}

type Pipeline struct {
	Path            string
	URL             string
	SSHKeyLocation  string
	Name            string
	BuildScript     string
	ForceCleanBuild bool
	Triggers        []Trigger
}

type Config struct {
	MaxHistoryInMemory  uint
	HTTPPort            int
	HTTPHost            string
	WorkspaceDirectory  string
	Pipelines           []Pipeline
	SQLDriver           string
	SQLConnectionString string
	DBMigrations        string
}

func LoadConfig(path string) Config {
	var res Config
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Printf("Error reading config file '%s':\n", path)
		log.Fatal(err)
	}

	err = yaml.Unmarshal(content, &res)
	if err != nil {
		log.Printf("Error parsing config file '%s':\n", path)
		log.Fatal(err)
	}

	return res
}
