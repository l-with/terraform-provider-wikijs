# Terraform Provider Wiki.js

Terraform provider for [Wiki.js](https://js.wiki/).

## Docs

All documentation for this provider can now be found on the Terraform Registry: https://registry.terraform.io/providers/camjjack/wikijs/latest/docs

## Installation

This provider can be installed automatically using Terraform >=0.13 by using the `terraform` configuration block:

```hcl
terraform {
  required_providers {
    wikijs = {
      source = "camjjack/wikijs"
      version = ">= 0.0.1"
    }
  }
}
```

## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

## Running tests

In order to run the full suite of Acceptance tests, you will either need docker (+ docker-compose) or minikube.

To run tests with docker-compose  run `make testacc-compose`.

To run tests with minikuve run  `make testacc`.
