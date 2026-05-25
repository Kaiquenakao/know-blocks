variable "project" {
  type = string
}

variable "documents_bucket_name" {
  type = string
}

variable "doc_validator_role_arn" {
  type = string
}

variable "pipe_trigger_role_arn" {
  type = string
}

variable "state_machine_arn" {
  type    = string
  default = ""
}

variable "extract_text_role_arn" {
  type = string
}

variable "chunk_document_role_arn" {
  type = string
}

variable "executions_table_name" {
  type    = string
  default = ""
}

variable "save_metadata_role_arn" {
  type = string
}

variable "documents_table_name" {
  type    = string
  default = ""
}

variable "chunks_table_name" {
  type    = string
  default = ""
}

variable "vector_bucket_name" {
  type    = string
  default = ""
}

variable "vector_index_name" {
  type    = string
  default = "global-chunks"
}
