# Variables for configuration
variable "ssh_cidr" {
  type        = string
  description = "Your home IP in CIDR notation"
}

variable "ssh_key_name" {
  type        = string
  description = "Name of your existing AWS key pair"
}

# AWS Provider
provider "aws" {
  region = "us-west-2"
}

# Security Group with SSH and Application Port
resource "aws_security_group" "docker_app" {
  name        = "docker-go-app-sg"
  description = "Security group for Docker Go application"
  
  ingress {
    description = "SSH"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.ssh_cidr]
  }
  
  ingress {
    description = "Go App"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name = "docker-go-app-sg"
  }
}

# Latest Amazon Linux 2023 AMI
data "aws_ami" "al2023" {
  most_recent = true
  owners      = ["amazon"]
  filter {
    name   = "name"
    values = ["al2023-ami-*-x86_64-ebs"]
  }
}

# EC2 Instance with Docker and Git
resource "aws_instance" "docker_server" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = "t2.micro"
  iam_instance_profile   = "LabInstanceProfile"
  vpc_security_group_ids = [aws_security_group.docker_app.id]
  key_name               = var.ssh_key_name

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
    delete_on_termination = true
  }

  user_data = <<-EOF
    #!/bin/bash
    dnf update -y
    dnf install -y docker git
    systemctl start docker
    systemctl enable docker
    usermod -a -G docker ec2-user
    touch /tmp/setup-complete
  EOF
  
  tags = {
    Name = "hw2-docker-server"
    Project = "6650-HW2"
  }
}

# Outputs
output "ec2_public_dns" {
  value = aws_instance.docker_server.public_dns
  description = "Public DNS of the EC2 instance"
}

output "ec2_public_ip" {
  value = aws_instance.docker_server.public_ip
  description = "Public IP of the EC2 instance"
}

output "ssh_command" {
  value = "ssh -i ~/${var.ssh_key_name}.pem ec2-user@${aws_instance.docker_server.public_dns}"
  description = "SSH connection command"
}

output "app_url" {
  value = "http://${aws_instance.docker_server.public_ip}:8080/albums"
  description = "Application URL"
}

# Create a second EC2 instance
resource "aws_instance" "docker_server_2" {
  ami                    = data.aws_ami.al2023.id
  instance_type          = "t2.micro"
  iam_instance_profile   = "LabInstanceProfile"
  vpc_security_group_ids = [aws_security_group.docker_app.id]
  key_name               = var.ssh_key_name

  root_block_device {
    volume_size = 30
    volume_type = "gp3"
    delete_on_termination = true
  }

  user_data = <<-EOF
    #!/bin/bash
    dnf update -y
    dnf install -y docker git
    systemctl start docker
    systemctl enable docker
    usermod -a -G docker ec2-user
    touch /tmp/setup-complete
  EOF
  
  tags = {
    Name = "hw2-docker-server-2"
    Project = "6650-HW2"
  }
}

output "ec2_public_ip_2" {
  value = aws_instance.docker_server_2.public_ip
  description = "Public IP of the second EC2 instance"
}