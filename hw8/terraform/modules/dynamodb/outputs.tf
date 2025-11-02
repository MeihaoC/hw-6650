# Output the table name (for environment variables)
output "table_name" {
  description = "DynamoDB table name for shopping carts"
  value       = aws_dynamodb_table.shopping_carts.name
}

# Output the table ARN (useful for IAM permissions if needed)
output "table_arn" {
  description = "DynamoDB table ARN"
  value       = aws_dynamodb_table.shopping_carts.arn
}

