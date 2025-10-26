# Use existing LabRole instead of creating new roles
data "aws_iam_role" "lab_role" {
  name = "LabRole"
}

# Output the role ARN for reference
output "lab_role_arn" {
  value       = data.aws_iam_role.lab_role.arn
  description = "ARN of the LabRole"
}