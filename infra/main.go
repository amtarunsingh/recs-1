package main

import (
	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	cfg := config.Load()

	synth := awscdk.NewBootstraplessSynthesizer(&awscdk.BootstraplessSynthesizerProps{})

	dataProps := &DataStackProps{
		StackProps: awscdk.StackProps{Env: &awscdk.Environment{Account: jsii.String(cfg.Aws.AccountId), Region: jsii.String(cfg.Aws.Region)}, Synthesizer: synth},
	}
	_, _ = NewDataStack(app, "DataStack", dataProps)

	app.Synth(nil)
}
