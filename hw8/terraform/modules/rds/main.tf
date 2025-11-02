# RDS Security Group - Allows MySQL (port 3306) from ECS tasks only
resource "aws_security_group" "rds" {
  name        = "${var.service_name}-rds-sg"
  description = "Security group for RDS MySQL instance"
  vpc_id      = var.vpc_id

  ingress {
    from_port       = 3306
    to_port         = 3306
    protocol        = "tcp"
    security_groups = [var.security_group_id]  # ECS security group
    description     = "MySQL access from ECS tasks"
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    description = "Allow all outbound"
  }

  tags = {
    Name = "${var.service_name}-rds-sg"
  }
}

# DB Subnet Group - Defines which subnets RDS can use
# NOTE: RDS requires at least 2 subnets in different Availability Zones
# If you get an error, check your VPC has subnets in at least 2 AZs
resource "aws_db_subnet_group" "main" {
  name       = "${var.service_name}-db-subnet-group"
  subnet_ids = var.subnet_ids

  tags = {
    Name = "${var.service_name}-db-subnet-group"
  }
}

# RDS MySQL Instance - MySQL 8.0 on db.t3.micro (Free tier)
# Per homework: skip_final_snapshot=true, deletion_protection=false
resource "aws_db_instance" "main" {
  identifier = "${var.service_name}-mysql"

  engine         = "mysql"
  engine_version = var.engine_version  # MySQL 8.0
  instance_class = var.instance_class   # db.t3.micro

  allocated_storage     = var.allocated_storage
  max_allocated_storage = 100
  storage_type          = "gp2"
  storage_encrypted      = false

  db_name  = var.db_name
  username = var.db_username
  password = random_password.db_password.result

  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible     = false  # Private subnet only

  backup_retention_period = 0
  skip_final_snapshot     = true   # Required for homework
  deletion_protection      = false  # Required for homework

  performance_insights_enabled = false
  monitoring_interval          = 0

  tags = {
    Name = "${var.service_name}-mysql"
  }
}

# Random password for RDS - passed to ECS via env vars
# RDS MySQL restrictions: cannot contain '/', '@', '"', or spaces
# Using override_special to only allow safe special characters
resource "random_password" "db_password" {
  length         = 16
  special        = true
  override_special = "!#$%&*()-_=+[]{}<>:?"
  # Excludes: / @ " and space (not allowed by RDS MySQL)
}

