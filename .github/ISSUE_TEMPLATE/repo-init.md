---
name: Initialize Repository Template
title: 'Initialize Terraform Module Repository'
about: A template for kicking off repo initialization Actions
---

## Overview

An issue for kicking off repository initialization Actions.  Creating this issue with the default title causes an Action to run that generates a PR for further customization of the repository.

## Progress

- [ ] Update the repository README with a `PROVIDER` and `NAME`
   ```
   .repo-init <provider> <name> <description>

   # example
   .repo-init aws example-application provides an example application installation
   ```
   - [x] `terraform-`: standard prefix
   - [x] `provider`: infrastructure provider (`aws`, `google`, `azurerm`, `github`)
   - [x] `name`: A `-`-separated identifier describing what this module provides
   - [x] `description`: A one-liner describing the repository


- [ ] Add a test case to the repository
  ```
    .repo-add-test <environment>-<region>

    # example
    .repo-add-test stg-us-west-1
  ```

- [ ] Update the `Usage` section of the repository README

- [ ] Create an initial tag (or release)
