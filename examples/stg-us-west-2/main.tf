# ------------------------------------------------------------------------------
# FOUNDATION
# ------------------------------------------------------------------------------

resource "random_string" "id" {
  length  = 6
  special = false
}

module "foundation" {
  source = "github.com/jcudit/terraform-aws-foundation-minimal?ref=v0.0.1"

  region      = "us-west-1"
  environment = "staging"
}

# ------------------------------------------------------------------------------
# DATABASE
# ------------------------------------------------------------------------------

resource "random_string" "database_password" {
  length  = 16
  special = false
}

module "cluster" {

  source = "../../"

  environment = var.environment
  region      = var.region

  database_name   = "ptfe"
  master_username = "ptfe"
  master_password = random_string.database_password.result

  # Foundation
  vpc_id      = module.foundation.vpc_id
  cidr_blocks = module.foundation.public_cidr_blocks
  subnet_ids  = module.foundation.public_subnet_ids

}
