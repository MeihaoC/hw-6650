# Wire together four focused modules: network, ecr, logging, ecs.

module "network" {
  source         = "./modules/network"
  service_name   = var.service_name
  container_port = var.container_port
}

# Added an application load balancer
module "alb" {
  source             = "./modules/alb"
  service_name       = var.service_name
  container_port     = var.container_port
  vpc_id             = module.network.vpc_id
  subnet_ids         = module.network.subnet_ids
  security_group_id  = module.network.security_group_id
}

# RDS MySQL Module - MySQL 8.0 on db.t3.micro (Free tier) for persistent storage
# Per homework: skip_final_snapshot=true, deletion_protection=false
module "rds" {
  source            = "./modules/rds"
  service_name      = var.service_name
  vpc_id            = module.network.vpc_id
  subnet_ids        = module.network.subnet_ids
  security_group_id = module.network.security_group_id
}

# DynamoDB Module - NoSQL database for shopping carts (STEP II)
# Design: Single table with cart_id as partition key
# Items embedded in cart document (NoSQL pattern)
module "dynamodb" {
  source       = "./modules/dynamodb"
  service_name = var.service_name
}

module "ecr" {
  source          = "./modules/ecr"
  repository_name = var.ecr_repository_name
}

module "logging" {
  source            = "./modules/logging"
  service_name      = var.service_name
  retention_in_days = var.log_retention_days
}

# Reuse an existing IAM role for ECS tasks
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

module "ecs" {
  source             = "./modules/ecs"
  service_name       = var.service_name
  image              = "${module.ecr.repository_url}:latest"
  container_port     = var.container_port
  subnet_ids         = module.network.subnet_ids
  security_group_ids = [module.network.security_group_id]
  execution_role_arn = data.aws_iam_role.lab_role.arn
  task_role_arn      = data.aws_iam_role.lab_role.arn
  log_group_name     = module.logging.log_group_name
  ecs_count          = var.ecs_count
  region             = var.aws_region
  target_group_arn   = module.alb.target_group_arn  # Added this
  
  # Database connection - passed to ECS tasks as environment variables
  # MySQL (RDS) - for STEP I
  db_host     = module.rds.db_host
  db_port     = tostring(module.rds.db_port)
  db_user     = module.rds.db_username
  db_password = module.rds.db_password
  db_name     = module.rds.db_name
  
  # DynamoDB - for STEP II
  dynamodb_table_name = module.dynamodb.table_name
  aws_region         = var.aws_region
  use_dynamodb       = var.use_dynamodb
  
  depends_on = [module.rds, module.dynamodb]
}

# Auto Scaling module - ADD THIS ENTIRE BLOCK
module "autoscaling" {
  source                  = "./modules/autoscaling"
  service_name            = module.ecs.service_name
  cluster_name            = module.ecs.cluster_name
  min_capacity            = var.min_capacity
  max_capacity            = var.max_capacity
  target_cpu_utilization  = var.target_cpu_utilization
  scale_in_cooldown       = var.scale_in_cooldown
  scale_out_cooldown      = var.scale_out_cooldown

  depends_on = [module.ecs]
}

// Build & push the Go app image into ECR
resource "docker_image" "app" {
  # Use the URL from the ecr module, and tag it "latest"
  name = "${module.ecr.repository_url}:latest"

  build {
    # relative path from terraform/ → src/
    context = "../src"
    # Dockerfile defaults to "Dockerfile" in that context
  }
}

resource "docker_registry_image" "app" {
  # this will push :latest → ECR
  name = docker_image.app.name
}
