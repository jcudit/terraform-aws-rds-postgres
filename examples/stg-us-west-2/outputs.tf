output "aws_rds_cluster" {
  description = "An `aws_rds_cluster` JSON object"
  value       = module.cluster.aws_rds_cluster
}

output "aws_rds_cluster_identifier" {
  description = "A cluster identifier for an `aws_rds_cluster` resource"
  value       = module.cluster.aws_rds_cluster_identifier
}

output "aws_rds_cluster_db_subnet_group_name" {
  description = "A subnet group name for an `aws_rds_cluster` resource"
  value       = module.cluster.aws_rds_cluster_db_subnet_group_name
}

output "aws_rds_cluster_security_groups" {
  description = "The security groups attached to `aws_rds_cluster_instance` resources"
  value       = module.cluster.aws_rds_cluster_security_groups
}

output "aws_rds_cluster_instance_endpoints" {
  description = "The endpoints of all `aws_rds_cluster_instance` resources"
  value       = module.cluster.aws_rds_cluster_instance_endpoints
}

output "aws_rds_cluster_instance_ids" {
  description = "The IDs of all `aws_rds_cluster_instance` resources"
  value       = module.cluster.aws_rds_cluster_instance_ids
}

output "reader_connection_string" {
  description = "Connection string for validating read access"
  value       = module.cluster.reader_connection_string
}
