# terraform-aws-rds-postgres

## Overview

This repository holds a module for AWS that provides a PostgreSQL RDS database.

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:-----:|
| cidr\_blocks | CIDR blocks to allow access to the database | `list(string)` | n/a | yes |
| database\_name | `database\_name` of a `aws\_rds\_cluster` resource | `string` | n/a | yes |
| environment | The environment this module will run in | `string` | n/a | yes |
| master\_password | `master\_password` of a `aws\_rds\_cluster` resource | `string` | n/a | yes |
| master\_username | `master\_username` of a `aws\_rds\_cluster` resource | `string` | n/a | yes |
| region | The region this module will run in | `string` | n/a | yes |
| service | The service this database will be owned by | `string` | n/a | yes |
| subnet\_ids | Subnet IDs to receive access to the database | `list(string)` | n/a | yes |
| vpc\_id | A VPC ID used to give subnets access to the database | `string` | n/a | yes |

## Outputs

| Name | Description |
|------|-------------|
| aws\_rds\_cluster | An `aws\_rds\_cluster` JSON object |
| aws\_rds\_cluster\_instance\_endpoints | The endpoints of all `aws\_rds\_cluster\_instance` resources |
| aws\_rds\_cluster\_instance\_ids | The IDs of all `aws\_rds\_cluster\_instance` resources |
| reader\_connection\_string | Connection string for validating read access |

## Usage

```hcl
module "cluster" {

  source = "github.com/jcudit/terraform-aws-rds-postgres?ref=v0.0.1"

  environment = var.environment
  region      = var.region

  database_name   = "ptfe"
  master_username = "ptfe"
  master_password = "changeme"

  # Foundation
  vpc_id      = module.foundation.vpc_id
  cidr_blocks = module.foundation.cidr_blocks
  subnet_ids  = module.foundation.subnet_ids

}
```

Check out the [examples](../examples) for fully-working sample code that the [tests](../test) exercise. Paved path testing patterns are documented [here](https://github.com/github/terraform-enterprise/blob/master/docs/modules.md#testing).

---

This repo has the following folder structure:

* root folder: The root folder contains a single, standalone, reusable, production-grade module.
* [modules](./modules): This folder may contain supporting modules to the root module.
* [examples](./examples): This folder shows examples of different ways to configure the root module and is typically exercised by tests.
* [test](./test): Automated tests for the modules and examples.

See the [official docs](https://www.terraform.io/docs/modules/index.html) for further details.

---

This repository was initialized with an Issue Template.
[See here](https://github.com/github/terraform-aws-rds-postgres/issues/new/choose).
