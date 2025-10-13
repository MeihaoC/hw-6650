# Region to deploy into
variable "aws_region" {
  type    = string
  default = "us-west-2"
}

# ECR & ECS settings
variable "ecr_repository_name" {
  type    = string
  default = "product-search-repo" # Updated
}

variable "service_name" {
  type    = string
  default = "product-search-service" # Updated
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "ecs_count" {
  type    = number
  default = 2 # Updated this
}

# How long to keep logs
variable "log_retention_days" {
  type    = number
  default = 7
}

# Auto Scaling settings - ADD THESE
variable "min_capacity" {
  type        = number
  default     = 2
  description = "Minimum number of ECS tasks"
}

variable "max_capacity" {
  type        = number
  default     = 4
  description = "Maximum number of ECS tasks"
}

variable "target_cpu_utilization" {
  type        = number
  default     = 70
  description = "Target CPU utilization percentage for auto scaling"
}

variable "scale_in_cooldown" {
  type        = number
  default     = 300
  description = "Cooldown period (seconds) before scaling in"
}

variable "scale_out_cooldown" {
  type        = number
  default     = 300
  description = "Cooldown period (seconds) before scaling out"
}