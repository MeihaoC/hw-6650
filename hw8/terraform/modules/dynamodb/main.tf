# DynamoDB Table for Shopping Carts
# Design Decision: Single table with cart_id as partition key
# Items are embedded in the cart document (NoSQL pattern)
resource "aws_dynamodb_table" "shopping_carts" {
  name         = "${var.service_name}-carts"
  billing_mode = "PAY_PER_REQUEST"  # On-demand pricing (simpler, scales automatically)

  # Partition key: cart_id (string) - UUID ensures even distribution
  # No sort key needed for access patterns: get by cart_id only
  hash_key = "cart_id"

  attribute {
    name = "cart_id"
    type = "S"  # String
  }

  # Note: customer_id can still be stored in items, but doesn't need to be
  # declared here unless it's used as a key or in a Global Secondary Index (GSI)

  # Tags for resource management
  tags = {
    Name        = "${var.service_name}-carts"
    Service     = var.service_name
    Environment = "homework"
  }

  # Point-in-time recovery (optional, can disable for cost savings)
  point_in_time_recovery {
    enabled = false
  }

  # Stream settings (disabled for now)
  stream_enabled   = false
}

