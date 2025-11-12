#!/usr/bin/env sh

set -eux
trap 'echo "[infra] EXIT status=$?"' EXIT

#  Env
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-dummy}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-dummy}"
export AWS_SESSION_TOKEN="${AWS_SESSION_TOKEN:-}"
export AWS_REGION="${AWS_REGION:-us-east-2}"
export AWS_DEFAULT_REGION="${AWS_DEFAULT_REGION:-$AWS_REGION}"
export CDK_DEFAULT_ACCOUNT="${CDK_DEFAULT_ACCOUNT:-000000000000}"
export ENV_TYPE="${ENV_TYPE:-local}"

# Point SDKs to LocalStack
export AWS_ENDPOINT_URL="${AWS_ENDPOINT_URL:-http://localstack:4566}"
export AWS_ENDPOINT_URL_S3="${AWS_ENDPOINT_URL}"
export AWS_EC2_METADATA_DISABLED="${AWS_EC2_METADATA_DISABLED:-true}"

# Toolchain
export PATH="/usr/local/go/bin:$PATH"
export GOTOOLCHAIN="${GOTOOLCHAIN:-auto}"

echo "[infra] ENV SUMMARY: ACCOUNT=$CDK_DEFAULT_ACCOUNT REGION=$AWS_REGION ENDPOINT=$AWS_ENDPOINT_URL"

# Install tools
apk add --no-cache nodejs npm jq curl aws-cli >/dev/null || true
# npm config set fund false audit false progress=false >/dev/null
npm i -g aws-cdk@2 aws-cdk-local >/dev/null

echo "[infra] versions"
which go && go version
node --version
npm --version
cdklocal --version
aws --version || true
jq --version

# Ensure LocalStack is up
echo "[infra] LocalStack health (one shot):"
set +e
curl -sf "$AWS_ENDPOINT_URL/_localstack/health" | jq . || true
set -e

# Synthesize & verify JSON 
cdklocal context --clear
ENV_TYPE="$ENV_TYPE" cdklocal synth -j DataStack 1>/work/infra/tmp.cfn.json 2>/work/infra/synth.stderr.log || true

echo "[infra] resource types in template:"
jq -r '.Resources | to_entries[] | .value.Type' /work/infra/tmp.cfn.json | sort -u

# Deploy (NO LOOKUPS, NO BOOTSTRAP) 
ENV_TYPE="$ENV_TYPE" \
AWS_ENVAR_ALLOWLIST=AWS_REGION AWS_REGION=${AWS_REGION} cdklocal deploy DataStack \
  --require-approval never \
  --progress events \
  --verbose \
  --no-lookups
  
#  Verify against LocalStack 
AWS="aws --endpoint-url $AWS_ENDPOINT_URL --region $AWS_REGION"
$AWS dynamodb list-tables || true
$AWS sns list-topics || true
$AWS sqs list-queues || true

echo "[infra] DONE"
