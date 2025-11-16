package config

import (
	"fmt"
	"os"

	"github.com/bmbl-bumble2/recs-votes-storage/internal/shared/timeutil"
	env "github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
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
	ConnectionArn string `env:"CONNECTION_ARN"`
	Owner         string `env:"REPO_OWNER"`
	Repo          string `env:"PIPELINE_REPO"`
	Branch        string `env:"PIPELINE_BRANCH"`
}

type AWSConfig struct {
	Region           string `env:"AWS_REGION"`
	AccountId        string `env:"AWS_ACCOUNT_ID"`
	AccessKeyId      string `env:"AWS_ACCESS_KEY_ID"`
	SecretAccessKey  string `env:"AWS_SECRET_ACCESS_KEY"`
	DynamoDbEndpoint string `env:"DYNAMO_DB_ENDPOINT"`
	SnsEndpoint      string `env:"SNS_ENDPOINT"`
}

type Config struct {
	LogLevel string `env:"LOG_LEVEL"`
	Aws      AWSConfig
	Counters CountersConfig
	Romances RomancesConfig
	Pipeline PipelineConfig
}

type ServerOptions struct {
	Host string `doc:"Hostname to listen on." default:"0.0.0.0"`
	Port int    `doc:"Port to listen on." short:"p" default:"8880"`
}

func Load() Config {
	_ = godotenv.Load("./../.env.local")
	cfg := Config{
		Aws: AWSConfig{
			AccountId:       os.Getenv("AWS_ACCOUNT_ID"),
			Region:          os.Getenv("AWS_REGION"),
			AccessKeyId:     os.Getenv("AWS_ACCESS_KEY_ID"),
			SecretAccessKey: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		},
		Counters: CountersConfig{
			TtlSeconds: CountersTtlHours * timeutil.HourSeconds,
		},
		Romances: RomancesConfig{
			MutualRomanceTtlSeconds:    546 * timeutil.DaySeconds,
			NonMutualRomanceTtlSeconds: 180 * timeutil.DaySeconds,
			DeadRomanceTtlSeconds:      90 * timeutil.DaySeconds,
		},
	}
	fmt.Println("AWS_ACCOUNT_ID", cfg.Aws.AccountId)
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
	return cfg
}
