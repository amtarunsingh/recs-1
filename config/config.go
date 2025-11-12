package config

import (
	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/timeutil"
	env "github.com/caarlos0/env/v10"
)

const (
	CountersTtlHours                    = 48
	ProjectName                         = "User Votes Storage"
	ProjectVersion                      = "1.0.0"
	DynamoDbVersionConflictRetriesCount = 3
)

type RomancesConfig struct {
	MutualRomanceTtlSeconds    int64
	NonMutualRomanceTtlSeconds int64
	DeadRomanceTtlSeconds      int64
}

type CountersConfig struct {
	TtlSeconds int64
}

type PipelineConfig struct {
	ConnectionArn string `env:"CONNECTION_ARN" envDefault:"arn:aws:codeconnections:us-east-2:875632962180:connection/cc2e1ba7-8dd3-48bf-a687-cfc28fbd99e2"`
	Branch        string `env:"PIPELINE_BRANCH" envDefault:"master"`
}

type Config struct {
	LogLevel string `env:"LOG_LEVEL" envDefault:"INFO"`
	Aws      struct {
		Region           string `env:"AWS_REGION" envDefault:"us-east-2"`
		AccountId        string `env:"AWS_ACCOUNT_ID" envDefault:"000000000000"`
		AccessKeyId      string `env:"AWS_ACCESS_KEY_ID" envDefault:"dummy"`
		SecretAccessKey  string `env:"AWS_SECRET_ACCESS_KEY" envDefault:"dummy"`
		DynamoDbEndpoint string `env:"DYNAMO_DB_ENDPOINT"`
		SnsEndpoint      string `env:"SNS_ENDPOINT"`
	}
	Counters CountersConfig
	Romances RomancesConfig
	Pipeline PipelineConfig
}

type ServerOptions struct {
	Host string `doc:"Hostname to listen on." default:"0.0.0.0"`
	Port int    `doc:"Port to listen on." short:"p" default:"8880"`
}

func Load() Config {
	cfg := Config{
		Counters: CountersConfig{
			TtlSeconds: CountersTtlHours * timeutil.HourSeconds,
		},
		Romances: RomancesConfig{
			MutualRomanceTtlSeconds:    546 * timeutil.DaySeconds,
			NonMutualRomanceTtlSeconds: 180 * timeutil.DaySeconds,
			DeadRomanceTtlSeconds:      90 * timeutil.DaySeconds,
		},
	}
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
