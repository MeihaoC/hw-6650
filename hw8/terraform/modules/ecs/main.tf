# ECS Cluster
resource "aws_ecs_cluster" "this" {
  name = "${var.service_name}-cluster"
}

# Task Definition
resource "aws_ecs_task_definition" "this" {
  family                   = "${var.service_name}-task"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.cpu
  memory                   = var.memory

  execution_role_arn = var.execution_role_arn
  task_role_arn      = var.task_role_arn

  container_definitions = jsonencode([{
    name      = "${var.service_name}-container"
    image     = var.image
    essential = true

    portMappings = [{
      containerPort = var.container_port
    }]

    environment = [
      {
        name  = "DB_HOST"
        value = var.db_host
      },
      {
        name  = "DB_PORT"
        value = var.db_port
      },
      {
        name  = "DB_USER"
        value = var.db_user
      },
      {
        name  = "DB_NAME"
        value = var.db_name
      },
      {
        name  = "PORT"
        value = tostring(var.container_port)
      },
      {
        name  = "DB_PASSWORD"
        value = var.db_password
      }
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = var.log_group_name
        "awslogs-region"        = var.region
        "awslogs-stream-prefix" = "ecs"
      }
    }
  }])
}

# ECS Service
resource "aws_ecs_service" "this" {
  name            = var.service_name
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.this.arn
  desired_count   = var.ecs_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = var.subnet_ids
    security_groups = var.security_group_ids
    assign_public_ip = true
  }

  # Added this load balancer block
  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = "${var.service_name}-container"
    container_port   = var.container_port
  }

  # Added this - Ensures tasks are registered before marking as healthy
  depends_on = [var.target_group_arn]

  # ADD THIS LIFECYCLE BLOCK - IMPORTANT! 
  # This prevents Terraform from resetting desired_count when auto scaling changes it.
  lifecycle {
    ignore_changes = [desired_count]
  }
}