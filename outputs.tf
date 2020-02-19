output "aws_rds_cluster" {
  description = "An `aws_rds_cluster` JSON object"
  value       = aws_rds_cluster.cluster
}

output "aws_rds_cluster_identifier" {
  description = "A cluster identifier for an `aws_rds_cluster` resource"
  value       = aws_rds_cluster.cluster.cluster_identifier
}

output "aws_rds_cluster_db_subnet_group_name" {
  description = "A subnet group name for an `aws_rds_cluster` resource"
  value       = aws_rds_cluster.cluster.db_subnet_group_name
}

output "aws_rds_cluster_security_groups" {
  description = "The security groups attached to `aws_rds_cluster_instance` resources"
  value       = aws_rds_cluster.cluster.vpc_security_group_ids
}

output "aws_rds_cluster_instance_endpoints" {
  description = "The endpoints of all `aws_rds_cluster_instance` resources"
  value = [
    for instance in aws_rds_cluster_instance.per_subnet :
    instance.endpoint
  ]
}

output "aws_rds_cluster_instance_ids" {
  description = "The IDs of all `aws_rds_cluster_instance` resources"
  value = [
    for instance in aws_rds_cluster_instance.per_subnet :
    instance.id
  ]
}

output "reader_connection_string" {
  description = "Connection string for validating read access"
  value = format(
    "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
    aws_rds_cluster.cluster.reader_endpoint,
    aws_rds_cluster.cluster.port,
    aws_rds_cluster.cluster.master_username,
    aws_rds_cluster.cluster.master_password,
    aws_rds_cluster.cluster.database_name,
  )
}
