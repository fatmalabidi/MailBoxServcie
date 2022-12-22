package config

// TTMakeConf represents table test structure of LoadConfig test
type TTLoadConf struct {
	Name    string
	EnvVar  string
	ConfTag string
}

// CreateTTLoadConf creates table test for LoadConfig test
func CreateTTLoadConf() []TTLoadConf {
	tt := []TTLoadConf{
		{
			Name:    "test config",
			EnvVar:  "test",
			ConfTag: "test",
		}, {
			Name:    "dev config",
			EnvVar:  "development",
			ConfTag: "dev",
		}, {
			Name:    "default config",
			EnvVar:  "production",
			ConfTag: "prod",
		},
	}
	return tt
}
