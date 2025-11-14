package main

import (
	"os"

	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	awsecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/jsii-runtime-go"
	"github.com/bmbl-bumble2/recs-votes-storage/config"
)

func main() {
	defer jsii.Close()
	cfg := config.Load()

	app := awscdk.NewApp(nil)

	env := &awscdk.Environment{Account: jsii.String(cfg.Aws.AccountId), Region: jsii.String(cfg.Aws.Region)}
	envType := getOrDefault("ENV_TYPE", "")

	var synthesizer awscdk.IStackSynthesizer
	if envType == "local" {
		synthesizer = awscdk.NewLegacyStackSynthesizer()
	} else {
		synthesizer = awscdk.NewDefaultStackSynthesizer(nil)
	}

	stack := awscdk.NewStack(app, jsii.String("UserVotesStorage"), &awscdk.StackProps{
		Env:         env,
		Synthesizer: synthesizer,
	})

	data := DataStack(stack, "Data", &DataStackProps{
		StackProps: awscdk.StackProps{Env: env},
	})

	if envType != "local" {
		if cfg.Pipeline.ConnectionArn == "" {
			panic("CODECONNECTION_ARN must be set for non-local deployments")
		}
		vpc, alb, prodListener, testListener, blueTG, greenTG := NewNetwork(stack, "Net", nil)

		_, svc, execRole, taskRole := NewEcsService(
			stack, "Ecs",
			vpc,
			jsii.String(getOrDefault("SERVICE_NAME", "user-votes")),
			blueTG, greenTG,
		)

		data.Counters.GrantReadWriteData(taskRole)
		data.Romances.GrantReadWriteData(taskRole)
		data.DeleteRomancesFifoTopic.GrantPublish(taskRole)
		data.DeleteRomancesGroupFifoTopic.GrantPublish(taskRole)
		data.DeleteRomancesFifoQueue.GrantConsumeMessages(taskRole)
		data.DeleteRomancesGroupFifoQueue.GrantConsumeMessages(taskRole)

		dg := NewEcsDeployment(stack, "CD", svc, prodListener, testListener, blueTG, greenTG)

		repoName := getOrDefault("ECR_REPO_NAME", "user-votes-api")
		ecrRepo := awsecr.NewRepository(stack, jsii.String("EcrRepo"), &awsecr.RepositoryProps{
			RepositoryName:     jsii.String(repoName),
			ImageScanOnPush:    jsii.Bool(true),
			ImageTagMutability: awsecr.TagMutability_IMMUTABLE,
			LifecycleRules: &[]*awsecr.LifecycleRule{{
				MaxImageCount: jsii.Number(30),
			}},
		})

		NewPipeline(
			stack, "Pipe",
			jsii.String("user-votes-bluegreen"),
			jsii.String(cfg.Pipeline.ConnectionArn),
			jsii.String(cfg.Pipeline.Owner),
			jsii.String(cfg.Pipeline.Repo),
			jsii.String(cfg.Pipeline.Branch),
			ecrRepo,
			dg,
			execRole,
			taskRole,
		)

		awscdk.NewCfnOutput(stack, jsii.String("AlbDns"), &awscdk.CfnOutputProps{
			Value: alb.LoadBalancerDnsName(),
		})
	}

	app.Synth(nil)
}

func getOrDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
