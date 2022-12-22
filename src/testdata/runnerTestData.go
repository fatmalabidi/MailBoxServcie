package testdata

import "github.com/fatmalabidi/MailBoxServcie/src/config"

type runnerTestData struct {
	Name           string
	Config         *config.Config
	ExpectedError  int
	HasError       bool
	HasRunnerError bool
}

func PrepareCreateTestData() []runnerTestData {
	validConfig, _ := config.LoadConfig()
	invalidServerTypeConf := &config.Config{
		Tag: "test",
		Server: config.Server{
			Type:     "invalid",
			Deadline: 0,
		},
		Database: config.Database{
			Messages: config.Messages{
				Type: "invalid",
			},
		},
		Cognito: validConfig.Cognito,
	}

	return []runnerTestData{
		{
			Name:           "valid configuration",
			Config:         validConfig,
			ExpectedError:  0,
			HasError:       false,
			HasRunnerError: false,
		},
		{
			Name:           "invalid server type ",
			Config:         invalidServerTypeConf,
			ExpectedError:  0,
			HasError:       true,
			HasRunnerError: true,
		},
	}
}
