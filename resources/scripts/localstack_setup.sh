#!/bin/bash
# Setup script for LocalStack SNS topics and SQS queues

set -e

ENDPOINT_URL="${AWS_ENDPOINT_URL:-http://localstack:4566}"
REGION="${AWS_REGION:-us-east-1}"

# Override AWS credentials for LocalStack (LocalStack doesn't validate these)
export AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-test}"
export AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-test}"
export AWS_DEFAULT_REGION="${REGION}"

echo "Setting up LocalStack resources..."
echo "Endpoint: $ENDPOINT_URL"
echo "Region: $REGION"
echo ""

# Create SNS topic for events
echo "Creating SNS topic..."
TOPIC_ARN=$(aws --endpoint-url=$ENDPOINT_URL sns create-topic \
  --name events-topic \
  --region $REGION \
  --output text \
  --query 'TopicArn' 2>&1) || {
  echo "Error creating SNS topic:" >&2
  echo "$TOPIC_ARN" >&2
  exit 1
}

if [ -z "$TOPIC_ARN" ] || [[ ! "$TOPIC_ARN" =~ ^arn: ]]; then
  echo "Error: Failed to create SNS topic. Got: $TOPIC_ARN" >&2
  exit 1
fi

echo "Created SNS topic: $TOPIC_ARN"

# Create SQS queues
echo "Creating SQS queues..."

echo "Creating user-events queue..."
USER_EVENTS_QUEUE=$(aws --endpoint-url=$ENDPOINT_URL sqs create-queue \
  --queue-name user-events \
  --region $REGION \
  --output text \
  --query 'QueueUrl' 2>&1) || {
  echo "Error creating user-events queue:" >&2
  echo "$USER_EVENTS_QUEUE" >&2
  exit 1
}
echo "Created queue: $USER_EVENTS_QUEUE"

echo "Creating debug queue..."
DEBUG_QUEUE=$(aws --endpoint-url=$ENDPOINT_URL sqs create-queue \
  --queue-name debug \
  --region $REGION \
  --output text \
  --query 'QueueUrl' 2>&1) || {
  echo "Error creating debug queue:" >&2
  echo "$DEBUG_QUEUE" >&2
  exit 1
}
echo "Created queue: $DEBUG_QUEUE"

# Get queue ARNs for subscription
echo "Getting queue ARNs..."

USER_EVENTS_QUEUE_ARN=$(aws --endpoint-url=$ENDPOINT_URL sqs get-queue-attributes \
  --queue-url "$USER_EVENTS_QUEUE" \
  --attribute-names QueueArn \
  --region $REGION \
  --output text \
  --query 'Attributes.QueueArn' 2>&1) || {
  echo "Error getting user-events queue ARN:" >&2
  echo "$USER_EVENTS_QUEUE_ARN" >&2
  exit 1
}

DEBUG_QUEUE_ARN=$(aws --endpoint-url=$ENDPOINT_URL sqs get-queue-attributes \
  --queue-url "$DEBUG_QUEUE" \
  --attribute-names QueueArn \
  --region $REGION \
  --output text \
  --query 'Attributes.QueueArn' 2>&1) || {
  echo "Error getting debug queue ARN:" >&2
  echo "$DEBUG_QUEUE_ARN" >&2
  exit 1
}

# Subscribe queues to SNS topic with filter policies on message attributes
echo "Subscribing queues to SNS topic..."

echo "Subscribing user-events queue (prefix filter: user.)..."
USER_EVENTS_SUBSCRIPTION_ARN=$(aws --endpoint-url=$ENDPOINT_URL sns subscribe \
  --topic-arn "$TOPIC_ARN" \
  --protocol sqs \
  --notification-endpoint "$USER_EVENTS_QUEUE_ARN" \
  --attributes '{"FilterPolicy":"{\"event_type\":[{\"prefix\":\"user.\"}]}","FilterPolicyScope":"MessageAttributes","RawMessageDelivery":"true"}' \
  --region $REGION \
  --output text \
  --query 'SubscriptionArn' 2>&1) || {
  echo "Error subscribing user-events queue:" >&2
  echo "$USER_EVENTS_SUBSCRIPTION_ARN" >&2
  exit 1
}
echo "Subscribed user-events queue to topic with prefix filter policy"

echo "Subscribing debug queue to all events..."
DEBUG_SUBSCRIPTION_ARN=$(aws --endpoint-url=$ENDPOINT_URL sns subscribe \
  --topic-arn "$TOPIC_ARN" \
  --protocol sqs \
  --notification-endpoint "$DEBUG_QUEUE_ARN" \
  --attributes '{"RawMessageDelivery":"true"}' \
  --region $REGION \
  --output text \
  --query 'SubscriptionArn' 2>&1) || {
  echo "Error subscribing debug queue:" >&2
  echo "$DEBUG_SUBSCRIPTION_ARN" >&2
  exit 1
}
echo "Subscribed debug queue to all events"

echo ""
echo "Setup complete!"
echo ""
echo "Add these environment variables to your .env.local file:"
echo ""
echo "AWS_ENDPOINT=$ENDPOINT_URL"
echo "AWS_REGION=$REGION"
echo "EVENTS_TOPIC_ARN=$TOPIC_ARN"
echo "EVENTS_QUEUE_URL_USER=$USER_EVENTS_QUEUE"
echo ""
echo "Note: Copy .env.local.example to .env.local if you haven't already."
