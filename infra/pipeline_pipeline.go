package main

import (
	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	awscodebuild "github.com/aws/aws-cdk-go/awscdk/v2/awscodebuild"
	awscodedeploy "github.com/aws/aws-cdk-go/awscdk/v2/awscodedeploy"
	awscodepipeline "github.com/aws/aws-cdk-go/awscdk/v2/awscodepipeline"
	awscodepipelineactions "github.com/aws/aws-cdk-go/awscdk/v2/awscodepipelineactions"
	awsecr "github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	awsiam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

func NewPipeline(
	scope constructs.Construct, id string,
	pipelineName *string,
	connectionArn, repoOwner, repoName, repoBranch *string,
	ecrRepo awsecr.Repository,
	deployGroup awscodedeploy.IEcsDeploymentGroup,
	execRole awsiam.IRole,
	taskRole awsiam.IRole,
) awscodepipeline.Pipeline {

	sourceOut := awscodepipeline.NewArtifact(jsii.String("SourceOutput"), nil)
	buildOut := awscodepipeline.NewArtifact(jsii.String("BuildOutput"), nil)

	source := awscodepipelineactions.NewCodeStarConnectionsSourceAction(
		&awscodepipelineactions.CodeStarConnectionsSourceActionProps{
			ActionName:    jsii.String("Source"),
			ConnectionArn: connectionArn,
			Owner:         repoOwner,
			Repo:          repoName,
			Branch:        repoBranch,
			Output:        sourceOut,
			TriggerOnPush: jsii.Bool(true),
		})

	testProject := awscodebuild.NewPipelineProject(scope, jsii.String("TestProject"),
		&awscodebuild.PipelineProjectProps{
			Environment: &awscodebuild.BuildEnvironment{
				BuildImage: awscodebuild.LinuxBuildImage_STANDARD_7_0(),
				Privileged: jsii.Bool(true),
			},
			BuildSpec: awscodebuild.BuildSpec_FromSourceFilename(jsii.String("buildspec.yml")),
			EnvironmentVariables: &map[string]*awscodebuild.BuildEnvironmentVariable{
				"AWS_DEFAULT_REGION":  {Value: awscdk.Stack_Of(scope).Region()},
				"PIPELINE_STAGE":      {Value: jsii.String("tests")},
				"GO_REQUIRED_VERSION": {Value: jsii.String("1.25")},
			},
		})

	project := awscodebuild.NewPipelineProject(scope, jsii.String("BuildProject"),
		&awscodebuild.PipelineProjectProps{
			Environment: &awscodebuild.BuildEnvironment{
				BuildImage: awscodebuild.LinuxBuildImage_STANDARD_7_0(),
				Privileged: jsii.Bool(true),
			},
			BuildSpec: awscodebuild.BuildSpec_FromSourceFilename(jsii.String("buildspec.yml")),
			EnvironmentVariables: &map[string]*awscodebuild.BuildEnvironmentVariable{
				"ECR_REPO_NAME":      {Value: ecrRepo.RepositoryName()},
				"EXECUTION_ROLE_ARN": {Value: execRole.RoleArn()},
				"TASK_ROLE_ARN":      {Value: taskRole.RoleArn()},
				"AWS_DEFAULT_REGION": {Value: awscdk.Stack_Of(scope).Region()},
			},
		})

	ecrRepo.GrantPullPush(project)

	test := awscodepipelineactions.NewCodeBuildAction(&awscodepipelineactions.CodeBuildActionProps{
		ActionName: jsii.String("Tests"),
		Project:    testProject,
		Input:      sourceOut,
	})

	build := awscodepipelineactions.NewCodeBuildAction(&awscodepipelineactions.CodeBuildActionProps{
		ActionName: jsii.String("BuildAndPush"),
		Project:    project,
		Input:      sourceOut,
		Outputs:    &[]awscodepipeline.Artifact{buildOut},
	})

	deploy := awscodepipelineactions.NewCodeDeployEcsDeployAction(
		&awscodepipelineactions.CodeDeployEcsDeployActionProps{
			ActionName:                 jsii.String("BlueGreenDeploy"),
			DeploymentGroup:            deployGroup,
			AppSpecTemplateFile:        awscodepipeline.NewArtifactPath(buildOut, jsii.String("appspec.yaml")),
			TaskDefinitionTemplateFile: awscodepipeline.NewArtifactPath(buildOut, jsii.String("taskdef.json")),
			ContainerImageInputs: &[]*awscodepipelineactions.CodeDeployEcsContainerImageInput{
				{
					Input:                     buildOut,
					TaskDefinitionPlaceholder: jsii.String("IMAGE"),
				},
			},
		})

	pl := awscodepipeline.NewPipeline(scope, jsii.String("AppPipeline"),
		&awscodepipeline.PipelineProps{
			PipelineName: pipelineName,
			PipelineType: awscodepipeline.PipelineType_V2,
		})

	pl.AddStage(&awscodepipeline.StageOptions{StageName: jsii.String("Source"), Actions: &[]awscodepipeline.IAction{source}})
	pl.AddStage(&awscodepipeline.StageOptions{StageName: jsii.String("Test"), Actions: &[]awscodepipeline.IAction{test}})
	pl.AddStage(&awscodepipeline.StageOptions{StageName: jsii.String("Build"), Actions: &[]awscodepipeline.IAction{build}})
	pl.AddStage(&awscodepipeline.StageOptions{StageName: jsii.String("Deploy"), Actions: &[]awscodepipeline.IAction{deploy}})

	return pl
}
