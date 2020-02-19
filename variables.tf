# ------------------------------------------------------------------------------
# ENVIRONMENT VARIABLES
# Define these secrets as environment variables
# ------------------------------------------------------------------------------

# AWS_ACCESS_KEY_ID
# AWS_SECRET_ACCESS_KEY
# AWS_DEFAULT_REGION

# ------------------------------------------------------------------------------
# REQUIRED PARAMETERS
# You must provide a value for each of these parameters.
# ------------------------------------------------------------------------------

variable "environment" {
  description = "The environment this module will run in"
  type        = string
}

variable "region" {
  description = "The region this module will run in"
  type        = string
}

variable "vpc_id" {
  description = "A VPC ID used to give subnets access to the database"
  type        = string
}

variable "database_name" {
  description = "`database_name` of a `aws_rds_cluster` resource"
  type        = string
}

variable "master_username" {
  description = "`master_username` of a `aws_rds_cluster` resource"
  type        = string
}

variable "master_password" {
  description = "`master_password` of a `aws_rds_cluster` resource"
  type        = string
}

variable "subnet_ids" {
  description = "Subnet IDs to receive access to the database"
  type        = list(string)
}

variable "cidr_blocks" {
  description = "CIDR blocks to allow access to the database"
  type        = list(string)
}

# ------------------------------------------------------------------------------
# OPTIONAL PARAMETERS
# These parameters have reasonable defaults.
# ------------------------------------------------------------------------------
