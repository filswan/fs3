package sysconfig

import (
	"github.com/BurntSushi/toml"
	"log"
	"strings"
)

type Configuration struct {
	StandAlone bool `toml:"stand_alone"`
}

var config *Configuration

func InitSysConfig(configFile string) {
	if strings.Trim(configFile, " ") == "" {
		configFile = "./config/sysconfig.toml"
	}
	if metaData, err := toml.DecodeFile(configFile, &config); err != nil {
		log.Fatal("error:", err)
	} else {
		if !requiredFieldsAreGiven(metaData) {
			log.Fatal("required fields not given")
		}
	}
}

func GetSysConfig() Configuration {
	if config == nil {
		InitSysConfig("")
	}
	return *config
}

func requiredFieldsAreGiven(metaData toml.MetaData) bool {
	requiredFields := [][]string{
		{"stand_alone"},
	}

	for _, v := range requiredFields {
		if !metaData.IsDefined(v...) {
			log.Fatal("required fields ", v)
		}
	}

	return true
}
