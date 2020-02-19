# Architecture

This architecture was [recommended by Hashicorp](https://www.youtube.com/watch?v=G06j6HLWyYo&feature=youtu.be&t=620) in mid-2018.


- [Modules](#modules)
  - [Infrastructure Architecture](#infrastructure-architecture)
    - [Terminology](#terminology)
    - [Services](#services)
    - [Platforms](#platforms)
    - [Foundations](#foundations)
  - [Module Architecture](#module-architecture)
      - [Module Structure](#module-structure)
      - [Module Creation](#module-creation)
      - [Module Composition](#module-composition)
    - [Versioning](#versioning)
      - [Breaking Changes](#breaking-changes)
      - [Adopting Changes](#adopting-changes)
      - [Incentivizing Changes](#incentivizing-changes)
    - [Testing](#testing)
      - [Unit Testing](#unit-testing)
      - [Policy Testing](#policy-testing)

---

## Infrastructure Architecture

- [Modules](#modules)
  - [Infrastructure Architecture](#infrastructure-architecture)
    - [Terminology](#terminology)
    - [Services](#services)
    - [Platforms](#platforms)
    - [Foundations](#foundations)

Design choices with this architecture are motivated by the following one-liners:

> allows change to be constant at the service layer in order to create new features and new business value

> as you scale, you will find boundaries ... with separated responsibilities, work is accomplished independently with minimal conflict

> enables testing of smaller pieces of the overall system ... blast radiuses decrease

> organizational changes can kill infrastructure-as-code if unprepared

### Services

_Services_ are the topmost layer of infrastructure. They tend to only consume _outputs_ of other modules.

<img align="center" alt="Screen Shot 2019-10-31 at 10 36 56 AM" src="https://user-images.githubusercontent.com/5034582/67956492-93042a00-fbca-11e9-9eb4-9d482096d670.png">

---

A _Service_ is something that a domain team operates to provide business value:

> workloads (`ec2`, `fargate`), storage (`s3`), ingress configuration (`alb`)

> a single cloud provider account, which has one domain team with full responsibility and privileges ... domain teams operate acceptance and production environments that are 99% the same, facilitated by Terraform Enterprise workspaces

An example _Service_ is a backend engine for rich client-side data visualization in GitHub Issue comments.

### Platforms

_Platforms_ both consume _inputs_ and produce _outputs_ for other layers.

<img align="center" alt="Screen Shot 2019-10-31 at 10 36 32 AM" src="https://user-images.githubusercontent.com/5034582/67956451-7ec02d00-fbca-11e9-8485-6e4bbbc4abdc.png">

---

Domain teams build _Services_ on _Platforms_:

> a traditional platform team's responsibilities ... databases make sense here as well

Examples include:
- A container Runtime _Platform_ for _Services_ to operate on. `Kubernetes` is used to satisy this interface.
- A persistence _Platform_ to satisfy storage and caching needs of _Services_. `ElasticSearch` is used to satisfy this interface.


### Foundations

_Foundations_ are the bottom layer of infrastructure. They tend to only produce _outputs_ for other modules.

> if you build a foundation of a house, you do not have a house yet, you just have the foundation of a house

<img align="center" alt="Screen Shot 2019-11-01 at 11 35 45 AM" src="https://user-images.githubusercontent.com/5034582/68036321-cebaf480-fc9b-11e9-9966-ccfbc1e9a022.png">

---

Changes in a _Foundation_ layer can be frequent and independent of upper layers.

Resources created by _Foundation_ modules are:
- _Service_-specific
- Created per _Service_ instead of once and shared globally
- Within a _Service_ billing subaccount

Examples include:

> VPN / DirectConnect, VPCs, subnet routes, cloudwatch log groups, SNS alerting topics

> ingress and egress, logging, shared message bus, shared persistence, Identity and Access Management


## Module Architecture

The big picture described in [Infrastructure Architecture](#infrastructure-architecture) is assembled using modules. This design has opted to use standard module structure and composability patterns that already exist within the Terraform community. This section goes into further detail.

- [Modules](#modules)
  - [Module Architecture](#module-architecture)
      - [Module Structure](#module-structure)
      - [Module Creation](#module-creation)
      - [Module Composition](#module-composition)
    - [Versioning](#versioning)
      - [Breaking Changes](#breaking-changes)
      - [Adopting Changes](#adopting-changes)
      - [Incentivizing Changes](#incentivizing-changes)
    - [Testing](#testing)
      - [Unit Testing](#unit-testing)
      - [Policy Testing](#policy-testing)

#### Module Structure

This section describes the paved path module layout.

A module lives in its own VCS repository. The repository has the following folder structure:

* **root directory**: The root directory contains a single, standalone, reusable, production-grade module
  * The Terraform Registry requires the root of every repo to contain Terraform code
* **modules directory**: This directory may contain helper modules called by the standalone module in the root directory
* [examples](./examples): This directory contains working examples demonstrating different ways a root module can be configured
* [test](./test): Automated tests for modules, submodules and examples

```
.
├── README.md             # README explaining the root module
├── examples
│   ├── main.tf           # runnable example
│   └── terraform.tfstate # byproduct of running `terraform` locally via tests
├── main.tf               # root module declaration
├── modules               # submodules used by the root module
│   ├── teams
│   │   ├── README.md     # README explaining the submodule
│   │   ├── main.tf
│   │   ├── outputs.tf
│   │   ├── variables.tf
│   │   └── versions.tf
│   └── workspaces
│       ├── README.md
│       ├── main.tf
│       ├── outputs.tf
│       ├── variables.tf
│       └── versions.tf
├── outputs.tf
├── test                  # empty test directory (explained in detail later)
├── variables.tf
└── versions.tf
```

See the [official docs](https://www.terraform.io/docs/modules/index.html) for further details as we do not stray far from the standard.

---

#### Module Creation

Module creation is typically an automated process. An example implementation could be composed of:

- [Terraform](https://github.com/terraform-providers/terraform-provider-github)
- A [Template Repository](https://github.blog/2019-06-06-generate-new-repositories-with-repository-templates/)
- Automated repo-initialization through Actions chatops via Issue Comments
- Manual repo-initialization (following a checklist until automated)

#### Module Composition

Infrastructure is a versioned dependency of a _Service_.

A _Service_ can obtain infrastructure by writing Terraform beside their code.

```
.
├── app/
│   └── github.rb
├── .config
│   └── terraform
│       ├── .terraform
│       ├── data.tf
│       ├── main.tf
│       ├── modules
│       ├── terraform.auto.tfvars
│       ├── terraform.tfstate
│       ├── terraform.tfstate.backup
│       └── variables.tf
├── .gitignore
└── README.md
```

For example, repositories configured with a `.config/terraform` directory declare which _Platform_ and _Foundation_ modules will support the _Service_. `main.tf` assembles modules together:

```hcl
# ------------------------------------------------------------------------------
# FOUNDATION
# ------------------------------------------------------------------------------

module "foundation" {
  source  = "..."
  version = "v0.0.3"

  service     = "terraform-enterprise"
  region      = "us-west-1"
  environment = "staging"
}

# ------------------------------------------------------------------------------
# PLATFORMS
# ------------------------------------------------------------------------------

module "database" {

  source  = "..."
  version = "v0.0.2"

  environment = var.environment
  region      = var.region
  service     = var.service
  password    = random_string.database_password.result

  # Foundation
  subnet_ids           = module.foundation.object.database_subnets
  sg_ids               = [module.foundation.object.default_security_group_id]
  db_subnet_group_name = module.foundation.object.database_subnet_group

}

# ------------------------------------------------------------------------------
# ROOT MODULE
# ------------------------------------------------------------------------------

module "example_service" {
  source = "..."

  # Service
  environment  = var.environment
  region       = var.region
  owner        = var.owner
  ttl          = "99"

  # Platform
  database_endpoint = module.database.object.endpoint
  database_password = random_string.database_password.result

  # Foundation
  vpc_id     = module.foundation.vpc_id
  subnet_ids = module.foundation.object.private_subnets
  sg_ids     = [module.foundation.object.default_security_group_id]

}
```

### Versioning

Design choices for module versioning are biased towards solving current pain points of operating infrastructure at GitHub:

- The long tail of _Services_ are seldomly re-deployed, which delays upgrades to more secure, efficient and cost effective _Platforms_ and _Foundations_

- Owners of _Services_ are uninformed of the cost and security risk of their design choices when choosing _Platform_ and _Foundation_ layers

- While changes to _Services_ occur in discrete steps, changes to _Platforms_ and _Foundations_ are continuous. We lack a deterministic option for a _Service_ owner to layer upon.

---

[_Platforms_](#Platforms) and [_Foundations_](#Foundations) are versioned dependencies of _Services_. This table displays example compositions of modules:

| Foundation | Platform | Service |
| ---------  | -------- | ------- |
| `terraform-aws-foundation`@`v0.0.1` | `terraform-aws-database`@`v0.0.1` | `github/sparkles`@`v1.0.0`
| `terraform-gcp-foundation`@`v0.0.1` | `terraform-gcp-database`@`v0.0.1` | `github/sparkles`@`v1.0.0`
| `terraform-azure-foundation`@`v0.0.1` | `terraform-azure-database`@`v0.0.1` | `github/sparkles`@`v1.0.0`

Hypothetical scenarios that versioning allows for follow:

#### Breaking Changes

`v0.0.1` of `terraform-aws-container-runtime` operates a `Kubernetes` cluster leveraging EKS configured with a control plane exposed to the Internet over HTTPS. A change in `v1.0.0` of this _Platform_ replaces the HTTPS interface with a bastion host. The major version bump signals a breaking change as access to configure the cluster will now be incompatible for older clients.

#### Adopting Changes

`v0.0.1` of `terraform-aws-foundation`, a collection of general purpose IAM, VPC and networking resources, does not provide a shared message bus to communicate with _Services_ hosted on-prem. A change in `v0.1.0` of this _Foundation_ adds this feature.

_Platforms_ that want to offer this feature migrate from `v0.0.1` to `v0.1.0`.

_Services_ that want to take advantage of this feature migrate to a _Platform_ that supports `v0.1.0` or greater of `terraform-aws-foundation`.

_Services_ and _Platforms_ that benefit more from the stability of remaining on `v0.0.1` are not forced to change.  Additionally, _Services_ and _Platforms_ that remain on `v0.0.1` are unaffected by the availability of `v0.1.0` of the _Foundation_, which operates independently of `v0.0.1` infrastructure.

#### Incentivizing Changes

A security vulnerability affecting `v0.0.1` of `terraform-aws-vm-runtime`, GitHub's _Platform_ for VM workloads, is fixed in version `v0.0.2`.

All but one _Service_ re-provisions onto hypervisors tagged with `v0.0.2` to avoid exposure to increased vulnerability. Infrastructure builds begin failing for the remaining _Service_ due to the vulnerable `v0.0.1` dependency.

Owners of the remaining _Service_ have unique budget, availability and threat model perspectives.  To inform a decision of how to react to this situation, _Service_ owners have access to the following information:

- The cost of the underlying _Platform_ and _Foundation_ resources
- A value indicating an amount of increased vulnerability due to the underlying _Platform_ and _Foundation_ layers

### Testing

Adopting Sentinel and Terratest means joining a larger community of Terraform users that have already solved for [governance](https://github.com/hashicorp/terraform-guides/tree/master/governance/second-generation), and [unit](https://github.com/gruntwork-io/terratest/tree/master/examples) testing with these tools.

#### Unit Testing

Adopting [Terratest](https://github.com/gruntwork-io/terratest) transitively adopts the project's [best practices](https://github.com/gruntwork-io/terratest#testing-best-practices):

> Testing infrastructure as code (IaC) is hard. With general purpose programming languages (e.g., Java, Python, Ruby), you have a "localhost" environment where you can run and test the code before you commit. You can also isolate parts of your code from external dependencies to create fast, reliable unit tests. With IaC, neither of these advantages is typically available, as there isn't a "localhost" equivalent for most IaC code (e.g., I can't use Terraform to deploy an AWS VPC on my own laptop) and there's no way to isolate your code from the outside world (i.e., the whole point of a tool like Terraform is to make calls to AWS, so if you remove AWS, there's nothing left).

> That means that most of the tests are going to be integration tests that deploy into a real AWS account. This makes the tests effective at catching real-world bugs, but it also makes them much slower and more brittle.

Tests for modules and their submodules are collocated within a module repo:

```
.
├── README.md
├── examples
│   └── aws-foundation-vpc  # workspace used for test invocations of `terraform`
│       ├── main.tf
│       └── variables.tf
├── go.mod                  # generated via `go mod init` / `go mod tidy`
├── go.sum                  # generated via `go mod init` / `go mod tidy`
├── main.tf
├── modules
│   └── aws-foundation-vpc  # resources exercised by tests
│       ├── README.md
│       ├── main.tf
│       ├── outputs.tf
│       └── variables.tf
├── outputs.tf
├── test
│   └── aws-foundation-vpc  # tests for submodule resources
│       └── vpc_test.go
├── variables.tf
├── vendor                  # generated via `go mod vendor`
│   ├── github.com
│   │   ├── aws
│   │   ├── boombuler
│   │   ├── davecgh
│   │   ├── go-sql-driver
│   │   ├── google
│   │   ├── gruntwork-io
│   │   ├── jmespath
│   │   ├── pmezard
│   │   ├── pquerna
│   │   └── stretchr
│   ├── golang.org
│   │   └── x
│   ├── google.golang.org
│   │   └── appengine
│   ├── gopkg.in
│   │   └── yaml.v2
│   └── modules.txt
└── versions.tf
```

Check out the available [packages](https://github.com/gruntwork-io/terratest/tree/master/modules) used for testing different types of infrastructure.


#### Policy Testing

Adopting [Sentinel](https://docs.hashicorp.com/sentinel/language/spec) gives us deep visibility and convenient security checkpoints through the `plan` and `apply` phases of managing infrastructure. [Config, Plan, State and Run integrations](https://www.terraform.io/docs/cloud/sentinel/index.html) exist to allow scripting policy enforcement.

There are benefits to operating infrastructure in different environments and regions. Policies are associated globally for each environment. This design allows for:
- Graduation of policy towards `production`
- Enforcement of policy with increasing stringency as changes graduate toward `production`
- Enforcement of policy with differing stringency for arbirtrary regions

The following is a rough example of how policy is organized and gives a snapshot of the mocking / testing capabilities that Sentinel provides:

```
.
├── README.md
├── examples
│   ├── main.tf
│   └── stg-us-east-1
│       └── main.tf
├── main.tf
├── modules
│   ├── policy-aws
│   │   ├── README.md
│   │   ├── main.tf
│   │   ├── outputs.tf
│   │   └── variables.tf
│   ├── ...
│   │   ├── README.md
│   │   ├── main.tf
│   │   ├── outputs.tf
│   │   └── variables.tf
│   ├── policy-common
│   │   ├── README.md
│   │   ├── main.tf
│   │   ├── outputs.tf
│   │   ├── policies
│   │   │   ├── blacklist-datasources.sentinel
│   │   │   ├── blacklist-providers.sentinel
│   │   │   ├── *.sentinel
│   │   │   └── variable_description_present.sentinel
│   │   ├── sentinel.hcl
│   │   └── variables.tf
│   └── policy-gcp
│       ├── README.md
│       ├── main.tf
│       ├── outputs.tf
│       └── variables.tf
├── test
│   └── policy-common
│       ├── blacklist-datasources
│       │   ├── fail-0.11.json
│       │   ├── fail-0.12.json
│       │   ├── mock-tfconfig-fail-0.11.sentinel
│       │   ├── mock-tfconfig-fail-0.12.sentinel
│       │   ├── mock-tfconfig-pass-0.11.sentinel
│       │   ├── mock-tfconfig-pass-0.12.sentinel
│       │   ├── pass-0.11.json
│       │   └── pass-0.12.json
│       ├── blacklist-providers
│       │   ├── fail-0.11.json
│       │   ├── fail-0.12.json
│       │   ├── mock-tfconfig-fail-0.11.sentinel
│       │   ├── mock-tfconfig-fail-0.12.sentinel
│       │   ├── mock-tfconfig-pass-0.11.sentinel
│       │   ├── mock-tfconfig-pass-0.12.sentinel
│       │   ├── pass-0.11.json
│       │   └── pass-0.12.json
│       ├── ...
│       └── validate-all-variables-have-descriptions
│           ├── fail-0.11.json
│           ├── fail-0.12.json
│           ├── mock-tfconfig-fail-0.11.sentinel
│           ├── mock-tfconfig-fail-0.12.sentinel
│           ├── mock-tfconfig-pass-0.11.sentinel
│           ├── mock-tfconfig-pass-0.12.sentinel
│           ├── mock-tfplan-fail-0.11.sentinel
│           ├── mock-tfplan-fail-0.12.sentinel
│           ├── mock-tfplan-pass-0.11.sentinel
│           ├── mock-tfplan-pass-0.12.sentinel
│           ├── pass-0.11.json
│           └── pass-0.12.json
├── variables.tf
├── outputs.tf
└── versions.tf
```

Check out the available [governance examples](https://github.com/hashicorp/terraform-guides/tree/master/governance/second-generation) provided by Hashicorp for practical examples of policies that can be enforced.
