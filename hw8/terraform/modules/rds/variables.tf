# Service name used for resource naming (e.g., "product-search-service-mysql")
variable "service_name" {
  type        = string
  description = "Service name for resource naming"
}

# VPC where the RDS instance will be deployed
variable "vpc_id" {
  type        = string
  description = "VPC ID for RDS instance"
}

# List of subnet IDs where RDS can be deployed (typically private subnets)
# REQUIRED: Must have at least 2 subnets in different Availability Zones
variable "subnet_ids" {
  type        = list(string)
  description = "List of subnet IDs for RDS subnet group (must be in at least 2 AZs)"
}

# Security group ID of ECS tasks - used to allow MySQL connections from ECS only
variable "security_group_id" {
  type        = string
  description = "Security group ID for ECS tasks that need to connect to RDS"
}

# Initial database name created when RDS instance starts
variable "db_name" {
  type        = string
  default     = "shoppingcart"
  description = "Initial database name"
}

# Master username for the MySQL database
variable "db_username" {
  type        = string
  default     = "admin"
  description = "Master username for RDS instance"
}

# Initial storage allocation in GB (can auto-scale up to max_allocated_storage)
variable "allocated_storage" {
  type        = number
  default     = 20
  description = "Allocated storage in GB"
}

# RDS instance class - db.t3.micro is Free tier eligible
variable "instance_class" {
  type        = string
  default     = "db.t3.micro"
  description = "RDS instance class"
}

# MySQL engine version (8.0 per homework requirements)
variable "engine_version" {
  type        = string
  default     = "8.0"
  description = "MySQL engine version"
}

