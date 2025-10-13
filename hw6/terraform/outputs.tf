output "ecs_cluster_name" {
  description = "Name of the created ECS cluster"
  value       = module.ecs.cluster_name
}

output "ecs_service_name" {
  description = "Name of the running ECS service"
  value       = module.ecs.service_name
}

# Added this - Very important to get the API endpoint
output "ecr_repository_url" {
  description = "URL of the ECR repository"
  value       = module.ecr.repository_url
}

# Added this - Shows the CloudWatch logs location
output "log_group_name" {
  description = "CloudWatch log group for debugging"
  value       = module.logging.log_group_name
}

# Added this - Region where everything is deployed
output "aws_region" {
  description = "AWS region where resources are deployed"
  value       = var.aws_region
}

# Added this
output "api_url" {
  description = "Public URL to access the API"
  value       = "http://${module.alb.alb_dns_name}"
}

# Auto Scaling outputs - ADD THESE
output "autoscaling_min_capacity" {
  description = "Minimum capacity for auto scaling"
  value       = var.min_capacity
}

output "autoscaling_max_capacity" {
  description = "Maximum capacity for auto scaling"
  value       = var.max_capacity
}

output "autoscaling_target_cpu" {
  description = "Target CPU utilization for auto scaling"
  value       = var.target_cpu_utilization
}