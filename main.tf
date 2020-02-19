# ------------------------------------------------------------------------------
# Prerequisites
# ------------------------------------------------------------------------------

locals {
  postgres_port = "5432"
}

resource "random_string" "id" {
  length  = 6
  upper   = false
  special = false
}

# ------------------------------------------------------------------------------
# Security
# ------------------------------------------------------------------------------

resource "aws_security_group" "private_access" {
  description = "allow access from private CIDR blocks"
  vpc_id      = var.vpc_id

  ingress {
    protocol  = "tcp"
    from_port = local.postgres_port
    to_port   = local.postgres_port

    cidr_blocks = var.cidr_blocks
  }
}

resource "aws_security_group" "public_access" {
  description = "allow public access to the database (testing, debugging)"
  vpc_id      = var.vpc_id

  ingress {
    protocol  = "tcp"
    from_port = local.postgres_port
    to_port   = local.postgres_port

    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_subnet_group" "cluster" {
  subnet_ids = var.subnet_ids

  tags = {
    Name = random_string.id.result
  }
}

# ------------------------------------------------------------------------------
# RDS Cluster
# ------------------------------------------------------------------------------

resource "aws_rds_cluster" "cluster" {

  # General
  database_name = var.database_name
  engine        = "aurora-postgresql"

  # Security
  master_username      = var.master_username
  master_password      = var.master_password
  storage_encrypted    = true
  db_subnet_group_name = aws_db_subnet_group.cluster.name
  vpc_security_group_ids = (
    var.environment == "production" ? [aws_security_group.private_access.id] :
    [aws_security_group.private_access.id, aws_security_group.public_access.id]
  )

  # Maintenance
  backup_retention_period      = 14
  preferred_backup_window      = "04:00-04:30"
  preferred_maintenance_window = "sun:12:00-sun:12:30"

  # Lifecycle
  final_snapshot_identifier = "${random_string.id.result}-final"
  deletion_protection       = var.environment == "production" ? true : false
  skip_final_snapshot       = var.environment == "production" ? false : true

}

resource "aws_rds_cluster_instance" "per_subnet" {
  count = length(var.subnet_ids)

  # General
  cluster_identifier = aws_rds_cluster.cluster.id
  identifier_prefix  = random_string.id.result
  engine             = "aurora-postgresql"
  apply_immediately  = true

  # Scale
  instance_class = "db.r5.large"

  # Security
  db_subnet_group_name = aws_db_subnet_group.cluster.name
  publicly_accessible  = var.environment == "production" ? false : true

}
