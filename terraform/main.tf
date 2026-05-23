terraform {
  required_version = ">= 1.5"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = "know-blocks-terraform-state"
    key    = "prod/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

module "s3" {
  source  = "./modules/s3"
  project = var.project
}

module "iam" {
  source               = "./modules/iam"
  project              = var.project
  documents_bucket_arn = module.s3.documents_bucket_arn
  state_machine_arn    = ""
}

module "lambdas" {
  source                 = "./modules/lambdas"
  project                = var.project
  documents_bucket_name  = module.s3.documents_bucket_name
  doc_validator_role_arn = module.iam.doc_validator_role_arn
  pipe_trigger_role_arn  = module.iam.pipe_trigger_role_arn
  state_machine_arn      = ""
}

module "api_gateway" {
  source                   = "./modules/api_gateway"
  project                  = var.project
  doc_validator_invoke_arn = module.lambdas.doc_validator_invoke_arn
  doc_validator_name       = module.lambdas.doc_validator_name
  pipe_trigger_invoke_arn  = module.lambdas.pipe_trigger_invoke_arn
  pipe_trigger_name        = module.lambdas.pipe_trigger_name
}
