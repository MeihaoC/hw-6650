terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 6.7.0"
    }
  }
}

provider "aws" {
  region = "us-west-2"
}

# ECR Repositories
resource "aws_ecr_repository" "order_api" {
  name                 = "order-api"
  image_tag_mutability = "MUTABLE"
  
  image_scanning_configuration {
    scan_on_push = false
  }
}

resource "aws_ecr_repository" "order_processor" {
  name                 = "order-processor"
  image_tag_mutability = "MUTABLE"
  
  image_scanning_configuration {
    scan_on_push = false
  }
}

# SNS Topic for order events
resource "aws_sns_topic" "order_processing_events" {
  name = "order-processing-events"
  
  tags = {
    Name = "Order Processing Events"
    Project = "HW7"
  }
}

# SQS Queue for order processing
resource "aws_sqs_queue" "order_processing_queue" {
  name                       = "order-processing-queue"
  visibility_timeout_seconds = 30
  message_retention_seconds  = 345600  # 4 days
  receive_wait_time_seconds  = 20      # Long polling
  
  tags = {
    Name = "Order Processing Queue"
    Project = "HW7"
  }
}

# Subscribe SQS queue to SNS topic
resource "aws_sns_topic_subscription" "order_queue_subscription" {
  topic_arn = aws_sns_topic.order_processing_events.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.order_processing_queue.arn
}

# Allow SNS to send messages to SQS
resource "aws_sqs_queue_policy" "order_queue_policy" {
  queue_url = aws_sqs_queue.order_processing_queue.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action = "SQS:SendMessage"
        Resource = aws_sqs_queue.order_processing_queue.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.order_processing_events.arn
          }
        }
      }
    ]
  })
}

# Output the ARNs for use in our application
output "sns_topic_arn" {
  value = aws_sns_topic.order_processing_events.arn
  description = "ARN of the SNS topic for order events"
}

output "sqs_queue_url" {
  value = aws_sqs_queue.order_processing_queue.url
  description = "URL of the SQS queue for order processing"
}

output "sqs_queue_arn" {
  value = aws_sqs_queue.order_processing_queue.arn
  description = "ARN of the SQS queue"
}

# Outputs for ECR
output "ecr_order_api_url" {
  value = aws_ecr_repository.order_api.repository_url
}

output "ecr_order_processor_url" {
  value = aws_ecr_repository.order_processor.repository_url
}