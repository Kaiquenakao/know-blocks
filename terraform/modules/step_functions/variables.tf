variable "project" {
  type = string
}

variable "step_functions_role_arn" {
  type = string
}

variable "extract_text_arn" {
  type = string
}

variable "chunk_document_arn" {
  type = string
}

variable "ecs_cluster_arn" {
  type = string
}

variable "embed_task_definition_arn" {
  type = string
}

variable "save_metadata_arn" {
  type = string
}


variable "subnet_ids" {
  type        = list(string)
  description = "Subnets para o Fargate rodar"
}
