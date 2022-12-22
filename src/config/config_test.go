package config_test

import (
	"github.com/fatmalabidi/MailBoxServcie/src/config"
	td "github.com/fatmalabidi/MailBoxServcie/src/testdata/config"
	"os"
	"testing"
)

func Test_LoadConfig(t *testing.T) {
	for _, testCase := range td.CreateTTLoadConf() {
		t.Run(testCase.Name, func(t *testing.T) {
			err := os.Setenv("CONFIGOR_ENV", testCase.EnvVar)
			if err != nil {
				t.Errorf("expected nil, got err in loading config: %v", err)
			}
			conf, err := config.LoadConfig()
			if err != nil {
				t.Errorf("expected loading config for %v environment, got error: %v", testCase.ConfTag, err)
			}
			if testCase.ConfTag != conf.Tag {
				t.Errorf("expected loading config for %v environment, got config for : %v", testCase.ConfTag, conf.Tag)
			}
		})
	}
}
