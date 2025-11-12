#!/usr/bin/env bash
set -eu

AWS_BASE="aws --region ${AWS_REGION} --endpoint-url ${DYNAMO_DB_ENDPOINT}"

${AWS_BASE} dynamodb create-table \
--table-name Counters \
--attribute-definitions AttributeName=u,AttributeType=S AttributeName=h,AttributeType=N \
--key-schema AttributeName=u,KeyType=HASH AttributeName=h,KeyType=RANGE \
--provisioned-throughput ReadCapacityUnits=10000,WriteCapacityUnits=2400

${AWS_BASE} dynamodb update-time-to-live \
  --table-name Counters \
  --time-to-live-specification "Enabled=true, AttributeName=ttl"

${AWS_BASE} dynamodb create-table \
--table-name Romances \
--attribute-definitions AttributeName=a,AttributeType=S AttributeName=b,AttributeType=S \
--key-schema AttributeName=a,KeyType=HASH AttributeName=b,KeyType=RANGE \
--provisioned-throughput ReadCapacityUnits=10000,WriteCapacityUnits=2400 \
--global-secondary-indexes '[
{"IndexName":"gsiByMaxMinUser",
"KeySchema":[{"AttributeName":"b","KeyType":"HASH"},{"AttributeName":"a","KeyType":"RANGE"}],
"Projection":{"ProjectionType":"KEYS_ONLY"},
"ProvisionedThroughput":{"ReadCapacityUnits":100,"WriteCapacityUnits":1300}}]'

${AWS_BASE} dynamodb update-time-to-live \
  --table-name Romances \
  --time-to-live-specification "Enabled=true, AttributeName=ttl"

echo "DynamoDB tables ready."

${AWS_BASE} sns create-topic --name delete-romances.fifo --attributes '{"FifoTopic":"true"}'
${AWS_BASE} sqs create-queue --queue-name delete-romances-queue.fifo --attributes '{"FifoQueue":"true"}'

${AWS_BASE} sns create-topic --name delete-romances-group.fifo --attributes '{"FifoTopic":"true"}'
${AWS_BASE} sqs create-queue --queue-name delete-romances-group-queue.fifo --attributes '{"FifoQueue":"true"}'

echo "SNS ready."