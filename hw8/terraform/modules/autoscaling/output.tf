output "autoscaling_target_id" {
  description = "ID of the auto scaling target"
  value       = aws_appautoscaling_target.ecs.id
}

output "autoscaling_policy_name" {
  description = "Name of the auto scaling policy"
  value       = aws_appautoscaling_policy.ecs_cpu.name
}

output "autoscaling_policy_arn" {
  description = "ARN of the auto scaling policy"
  value       = aws_appautoscaling_policy.ecs_cpu.arn
}