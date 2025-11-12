package main

import (
	awscdk "github.com/aws/aws-cdk-go/awscdk/v2"
	awsdynamodb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	awsiam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	awssns "github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	awssqs "github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/bmbl-bumble2/recs-votes-storage/internal/context/voting/infrastructure/persistence"
)

type DataStackProps struct {
	awscdk.StackProps

	GrantRwToRole awsiam.IGrantable
}

type DataOutputs struct {
	Counters                     awsdynamodb.ITable
	Romances                     awsdynamodb.ITable
	DeleteRomancesFifoTopic      awssns.ITopic
	DeleteRomancesFifoQueue      awssqs.IQueue
	DeleteRomancesGroupFifoTopic awssns.ITopic
	DeleteRomancesGroupFifoQueue awssqs.IQueue
}

func NewDataStack(scope constructs.Construct, id string, props *DataStackProps) (awscdk.Stack, *DataOutputs) {
	var sp awscdk.StackProps
	if props != nil {
		sp = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sp)

	var counters, romances awsdynamodb.ITable
	countersTbl := awsdynamodb.NewTable(stack, jsii.String(persistence.CountersTableName), &awsdynamodb.TableProps{
		TableName:    jsii.String(persistence.CountersTableName),
		PartitionKey: &awsdynamodb.Attribute{Name: jsii.String(persistence.UserIdAttrName), Type: awsdynamodb.AttributeType_STRING},
		SortKey:      &awsdynamodb.Attribute{Name: jsii.String(persistence.HourUnixTimestampAttrName), Type: awsdynamodb.AttributeType_NUMBER},
		BillingMode:  awsdynamodb.BillingMode_PAY_PER_REQUEST,
	})
	cfnCounters := countersTbl.Node().DefaultChild().(awscdk.CfnResource)
	cfnCounters.AddOverride(jsii.String("Properties.TimeToLiveSpecification"),
		map[string]interface{}{"Enabled": true, "AttributeName": "ttl"})
	counters = countersTbl

	romancesTbl := awsdynamodb.NewTable(stack, jsii.String(persistence.RomancesTableName), &awsdynamodb.TableProps{
		TableName:    jsii.String(persistence.RomancesTableName),
		PartitionKey: &awsdynamodb.Attribute{Name: jsii.String(persistence.PkUserIdAttrName), Type: awsdynamodb.AttributeType_STRING},
		SortKey:      &awsdynamodb.Attribute{Name: jsii.String(persistence.SkUserIdAttrName), Type: awsdynamodb.AttributeType_STRING},
		BillingMode:  awsdynamodb.BillingMode_PAY_PER_REQUEST,
	})
	cfnRomances := romancesTbl.Node().DefaultChild().(awscdk.CfnResource)
	cfnRomances.AddOverride(jsii.String("Properties.TimeToLiveSpecification"),
		map[string]interface{}{"Enabled": true, "AttributeName": "ttl"})
	romancesTbl.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
		IndexName:      jsii.String("gsiByMaxMinUser"),
		PartitionKey:   &awsdynamodb.Attribute{Name: jsii.String(persistence.SkUserIdAttrName), Type: awsdynamodb.AttributeType_STRING},
		SortKey:        &awsdynamodb.Attribute{Name: jsii.String(persistence.PkUserIdAttrName), Type: awsdynamodb.AttributeType_STRING},
		ProjectionType: awsdynamodb.ProjectionType_KEYS_ONLY,
	})
	romances = romancesTbl

	if props != nil && props.GrantRwToRole != nil {
		counters.GrantReadWriteData(props.GrantRwToRole)
		romances.GrantReadWriteData(props.GrantRwToRole)
	}

	var topic1, topic2 awssns.ITopic
	var queue1, queue2 awssqs.IQueue

	topic1 = awssns.NewTopic(stack, jsii.String("DeleteRomancesFifoTopic"), &awssns.TopicProps{
		TopicName: jsii.String("delete-romances.fifo"),
		Fifo:      jsii.Bool(true),
	})
	queue1 = awssqs.NewQueue(stack, jsii.String("DeleteRomancesFifoQueue"), &awssqs.QueueProps{
		QueueName: jsii.String("delete-romances-queue.fifo"),
		Fifo:      jsii.Bool(true),
	})

	topic2 = awssns.NewTopic(stack, jsii.String("DeleteRomancesGroupFifoTopic"), &awssns.TopicProps{
		TopicName: jsii.String("delete-romances-group.fifo"),
		Fifo:      jsii.Bool(true),
	})
	queue2 = awssqs.NewQueue(stack, jsii.String("DeleteRomancesGroupFifoQueue"), &awssqs.QueueProps{
		QueueName: jsii.String("delete-romances-group-queue.fifo"),
		Fifo:      jsii.Bool(true),
	})

	return stack, &DataOutputs{
		Counters:                     counters,
		Romances:                     romances,
		DeleteRomancesFifoTopic:      topic1,
		DeleteRomancesFifoQueue:      queue1,
		DeleteRomancesGroupFifoTopic: topic2,
		DeleteRomancesGroupFifoQueue: queue2,
	}
}
