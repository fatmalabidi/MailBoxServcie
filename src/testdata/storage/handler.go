 package storage

type TTCreate struct {
	Name     string
	Type     string
	HasError bool
}

func TTCreateS3Handler() []TTCreate {
	return []TTCreate{
		{
			"Valid Type",
			"s3",
			false,
		},
		{
			"Invalid Type",
			"invalid",
			true,
		},
	}
}