package main

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	awsec2 "github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	awsecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	awsecs "github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	awsiam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	awslogs "github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewEcsService(
	scope constructs.Construct,
	id string,
	vpc awsec2.IVpc,
	serviceName *string,
	blueTG awselasticloadbalancingv2.IApplicationTargetGroup,
	greenTG awselasticloadbalancingv2.IApplicationTargetGroup,
) (awsecs.Cluster, awsecs.FargateService, awsiam.IRole, awsiam.IRole) {

	cluster := awsecs.NewCluster(scope, jsii.String(id+"Cluster"), &awsecs.ClusterProps{Vpc: vpc})

	execRole := ensureExecutionRole(scope)

	taskRole := awsiam.NewRole(scope, jsii.String(id+"TaskRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})

	taskRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Effect: awsiam.Effect_ALLOW,
		Actions: &[]*string{
			jsii.String("ssmmessages:CreateControlChannel"),
			jsii.String("ssmmessages:CreateDataChannel"),
			jsii.String("ssmmessages:OpenControlChannel"),
			jsii.String("ssmmessages:OpenDataChannel"),
		},
		Resources: &[]*string{
			jsii.String("arn:aws:ecs:" + *awscdk.Stack_Of(scope).Region() + ":" + *awscdk.Stack_Of(scope).Account() + ":*")},
	}))

	logGroup := awslogs.NewLogGroup(scope, jsii.String(id+"LogGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/ecs/" + *serviceName),
		Retention:     awslogs.RetentionDays_ONE_MONTH,
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
	})

	task := awsecs.NewFargateTaskDefinition(scope, jsii.String(id+"Task"), &awsecs.FargateTaskDefinitionProps{
		Cpu:            jsii.Number(512),
		MemoryLimitMiB: jsii.Number(1024),
		ExecutionRole:  execRole,
		TaskRole:       taskRole,
	})

	ecrRepo := awsecr.Repository_FromRepositoryName(scope, jsii.String(id+"EcrRepo"), jsii.String("user-votes-api"))

	imageTag := os.Getenv("IMAGE_TAG")
	if imageTag == "" {
		imageTag = "latest"
	}

	task.AddContainer(jsii.String("app"), &awsecs.ContainerDefinitionOptions{
		Image:        awsecs.ContainerImage_FromEcrRepository(ecrRepo, jsii.String(imageTag)),
		Essential:    jsii.Bool(true),
		PortMappings: &[]*awsecs.PortMapping{{ContainerPort: jsii.Number(8888)}},
		EntryPoint:   &[]*string{jsii.String("sh"), jsii.String("-c")},
		Command: &[]*string{
			jsii.String("mkdir -p /www && echo OK > /www/health && httpd -f -p 8888 -h /www"),
		},
		Logging: awsecs.LogDrivers_AwsLogs(&awsecs.AwsLogDriverProps{
			StreamPrefix: jsii.String("app"),
			LogGroup:     logGroup,
		}),
	})

	svc := awsecs.NewFargateService(scope, jsii.String(id+"Service"), &awsecs.FargateServiceProps{
		Cluster:           cluster,
		ServiceName:       serviceName,
		TaskDefinition:    task,
		DesiredCount:      jsii.Number(2),
		MinHealthyPercent: jsii.Number(100),
		MaxHealthyPercent: jsii.Number(200),
		DeploymentController: &awsecs.DeploymentController{
			Type: awsecs.DeploymentControllerType_CODE_DEPLOY,
		},
		EnableExecuteCommand:   jsii.Bool(true),
		HealthCheckGracePeriod: awscdk.Duration_Seconds(jsii.Number(60)),
	})

	svc.AttachToApplicationTargetGroup(blueTG)
	svc.AttachToApplicationTargetGroup(greenTG)

	return cluster, svc, execRole, taskRole
}

func ensureExecutionRole(scope constructs.Construct) awsiam.IRole {
	if arn := os.Getenv("TASK_EXECUTION_ROLE_ARN"); arn != "" {
		return awsiam.Role_FromRoleArn(scope, jsii.String("ImportedEcsExecRole"), jsii.String(arn), &awsiam.FromRoleArnOptions{
			Mutable: jsii.Bool(false),
		})
	}
	role := awsiam.NewRole(scope, jsii.String("EcsTaskExecutionRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
	})
	role.AddManagedPolicy(awsiam.ManagedPolicy_FromAwsManagedPolicyName(
		jsii.String("service-role/AmazonECSTaskExecutionRolePolicy")))
	return role
}
