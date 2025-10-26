# ECS Cluster
resource "aws_ecs_cluster" "main" {
  name = "hw7-cluster"
  
  tags = {
    Name = "hw7-cluster"
  }
}

# CloudWatch Log Group
resource "aws_cloudwatch_log_group" "ecs" {
  name              = "/ecs/hw7"
  retention_in_days = 7
}

# Get AWS account ID
data "aws_caller_identity" "current" {}

# Order API Task Definition
resource "aws_ecs_task_definition" "order_api" {
  family                   = "hw7-order-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = data.aws_iam_role.lab_role.arn
  task_role_arn            = data.aws_iam_role.lab_role.arn
  
  container_definitions = jsonencode([
    {
      name  = "order-api"
      image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.us-west-2.amazonaws.com/order-api:latest"
      
      portMappings = [
        {
          containerPort = 8080
          protocol      = "tcp"
        }
      ]
      
      environment = [
        {
          name  = "SNS_TOPIC_ARN"
          value = aws_sns_topic.order_processing_events.arn
        }
      ]
      
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = "us-west-2"
          "awslogs-stream-prefix" = "order-api"
        }
      }
    }
  ])
}

# Order Processor Task Definition
resource "aws_ecs_task_definition" "order_processor" {
  family                   = "hw7-order-processor"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = data.aws_iam_role.lab_role.arn
  task_role_arn            = data.aws_iam_role.lab_role.arn
  
  container_definitions = jsonencode([
    {
      name  = "order-processor"
      image = "${data.aws_caller_identity.current.account_id}.dkr.ecr.us-west-2.amazonaws.com/order-processor:latest"
      
      environment = [
        {
          name  = "SQS_QUEUE_URL"
          value = aws_sqs_queue.order_processing_queue.url
        },
        {
          name  = "NUM_WORKERS"
          value = "100" # Change this: 1 → 5 → 20 → 100
        }
      ]
      
      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ecs.name
          "awslogs-region"        = "us-west-2"
          "awslogs-stream-prefix" = "order-processor"
        }
      }
    }
  ])
}

# Order API ECS Service
resource "aws_ecs_service" "order_api" {
  name            = "hw7-order-api"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.order_api.arn
  desired_count   = 1
  launch_type     = "FARGATE"
  
  network_configuration {
    subnets          = [aws_subnet.private_1.id, aws_subnet.private_2.id]
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }
  
  load_balancer {
    target_group_arn = aws_lb_target_group.order_api.arn
    container_name   = "order-api"
    container_port   = 8080
  }
  
  depends_on = [aws_lb_listener.http]
}

# Order Processor ECS Service
resource "aws_ecs_service" "order_processor" {
  name            = "hw7-order-processor"
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.order_processor.arn
  desired_count   = 1
  launch_type     = "FARGATE"
  
  network_configuration {
    subnets          = [aws_subnet.private_1.id, aws_subnet.private_2.id]
    security_groups  = [aws_security_group.ecs_tasks.id]
    assign_public_ip = false
  }
}