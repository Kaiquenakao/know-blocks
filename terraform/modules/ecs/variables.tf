variable "project" {
  type = string
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "ollama_image_url" {
  type        = string
  description = "ECR URL da imagem Ollama com modelos embutidos"
}

variable "embed_chunks_image_url" {
  type        = string
  description = "ECR URL da imagem embed-chunks"
}

variable "documents_bucket_name" {
  type = string
}

variable "executions_table_name" {
  type = string
}

variable "ecs_execution_role_arn" {
  type        = string
  description = "Role para o ECS agent (pull imagem, logs)"
}

variable "ecs_task_role_arn" {
  type        = string
  description = "Role para o código dentro do container (S3, DynamoDB)"
}
