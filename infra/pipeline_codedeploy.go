package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	awscodedeploy "github.com/aws/aws-cdk-go/awscdk/v2/awscodedeploy"
	awsecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"

	elbv2 "github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	awsiam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewEcsDeployment(scope constructs.Construct, id string,
	svc awsecs.FargateService,
	prodListener, testListener elbv2.ApplicationListener,
	blueTG, greenTG elbv2.ApplicationTargetGroup,
) awscodedeploy.IEcsDeploymentGroup {

	metric5xx := prodListener.LoadBalancer().Metrics().HttpCodeElb(elbv2.HttpCodeElb_ELB_5XX_COUNT, nil)
	alarm := awscloudwatch.NewAlarm(scope, jsii.String("Alb5xxAlarm"), &awscloudwatch.AlarmProps{
		Metric:            metric5xx,
		EvaluationPeriods: jsii.Number(1),
		Threshold:         jsii.Number(1),
		DatapointsToAlarm: jsii.Number(1),
	})

	app := awscodedeploy.NewEcsApplication(scope, jsii.String(id+"App"), nil)

	cdRole := awsiam.NewRole(scope, jsii.String(id+"Role"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("codedeploy.amazonaws.com"), nil),
	})
	cdRole.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(
		jsii.String("AWSCodeDeployRoleForECS"),
	))

	return awscodedeploy.NewEcsDeploymentGroup(scope, jsii.String(id+"DG"),
		&awscodedeploy.EcsDeploymentGroupProps{
			Application: app,
			Service:     svc,
			BlueGreenDeploymentConfig: &awscodedeploy.EcsBlueGreenDeploymentConfig{
				Listener:         prodListener,
				TestListener:     testListener,
				BlueTargetGroup:  blueTG,
				GreenTargetGroup: greenTG,
			},
			Role:             cdRole,
			DeploymentConfig: awscodedeploy.EcsDeploymentConfig_CANARY_10PERCENT_5MINUTES(),
			Alarms:           &[]awscloudwatch.IAlarm{alarm},
			AutoRollback: &awscodedeploy.AutoRollbackConfig{
				DeploymentInAlarm: jsii.Bool(true),
				FailedDeployment:  jsii.Bool(true),
			},
		})
}
