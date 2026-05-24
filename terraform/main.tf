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

module "dynamodb" {
  source  = "./modules/dynamodb"
  project = var.project
}

module "ecr" {
  source  = "./modules/ecr"
  project = var.project
}

module "iam" {
  source                = "./modules/iam"
  project               = var.project
  documents_bucket_arn  = module.s3.documents_bucket_arn
  executions_table_arn  = module.dynamodb.executions_table_arn
  state_machine_arn     = module.step_functions.state_machine_arn
  extract_text_arn      = module.lambdas.extract_text_arn
  chunk_document_arn    = module.lambdas.chunk_document_arn
}

module "lambdas" {
  source                  = "./modules/lambdas"
  project                 = var.project
  documents_bucket_name   = module.s3.documents_bucket_name
  doc_validator_role_arn  = module.iam.doc_validator_role_arn
  pipe_trigger_role_arn   = module.iam.pipe_trigger_role_arn
  extract_text_role_arn   = module.iam.extract_text_role_arn
  chunk_document_role_arn = module.iam.chunk_document_role_arn
  executions_table_name   = module.dynamodb.executions_table_name
  state_machine_arn       = module.step_functions.state_machine_arn
}

module "ecs" {
  source                 = "./modules/ecs"
  project                = var.project
  aws_region             = var.aws_region
  ollama_image_url       = module.ecr.ollama_repository_url
  embed_chunks_image_url = module.ecr.embed_chunks_repository_url
  documents_bucket_name  = module.s3.documents_bucket_name
  executions_table_name  = module.dynamodb.executions_table_name
  ecs_execution_role_arn = module.iam.ecs_execution_role_arn
  ecs_task_role_arn      = module.iam.ecs_task_role_arn
}

module "step_functions" {
  source                    = "./modules/step_functions"
  project                   = var.project
  step_functions_role_arn   = module.iam.step_functions_role_arn
  extract_text_arn          = module.lambdas.extract_text_arn
  chunk_document_arn        = module.lambdas.chunk_document_arn
  ecs_cluster_arn           = module.ecs.cluster_arn
  embed_task_definition_arn = module.ecs.embed_task_definition_arn
  subnet_ids                = var.subnet_ids
}

module "api_gateway" {
  source                   = "./modules/api_gateway"
  project                  = var.project
  doc_validator_invoke_arn = module.lambdas.doc_validator_invoke_arn
  doc_validator_name       = module.lambdas.doc_validator_name
  pipe_trigger_invoke_arn  = module.lambdas.pipe_trigger_invoke_arn
  pipe_trigger_name        = module.lambdas.pipe_trigger_name
}
