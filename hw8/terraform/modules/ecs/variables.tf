variable "service_name" {
  type        = string
  description = "Base name for ECS resources"
}

variable "image" {
  type        = string
  description = "ECR image URI (with tag)"
}

variable "container_port" {
  type        = number
  description = "Port your app listens on"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnets for FARGATE tasks"
}

variable "security_group_ids" {
  type        = list(string)
  description = "SGs for FARGATE tasks"
}

variable "execution_role_arn" {
  type        = string
  description = "ECS Task Execution Role ARN"
}

variable "task_role_arn" {
  type        = string
  description = "IAM Role ARN for app permissions"
}

variable "log_group_name" {
  type        = string
  description = "CloudWatch log group name"
}

variable "ecs_count" {
  type        = number
  default     = 1
  description = "Desired Fargate task count"
}

variable "region" {
  type        = string
  description = "AWS region (for awslogs driver)"
}

variable "cpu" {
  type        = string
  default     = "256"
  description = "vCPU units"
}

variable "memory" {
  type        = string
  default     = "512"
  description = "Memory (MiB)"
}

# Added this variable
variable "target_group_arn" {
  type        = string
  description = "ALB target group ARN"
  default     = null
}

# Database connection variables
variable "db_host" {
  type        = string
  description = "Database host"
  default     = ""
}

variable "db_port" {
  type        = string
  description = "Database port"
  default     = "3306"
}

variable "db_user" {
  type        = string
  description = "Database username"
  default     = ""
}

variable "db_password" {
  type        = string
  description = "Database password"
  sensitive   = true
  default     = ""
}

variable "db_name" {
  type        = string
  description = "Database name"
  default     = "shoppingcart"
}

# DynamoDB variables (for STEP II)
variable "dynamodb_table_name" {
  type        = string
  description = "DynamoDB table name for shopping carts"
  default     = ""
}

variable "aws_region" {
  type        = string
  description = "AWS region for DynamoDB client"
  default     = "us-west-2"
}

# Database selection variable
variable "use_dynamodb" {
  type        = bool
  description = "Set to true to use DynamoDB, false to use MySQL (default: false)"
  default     = false
}
