variable "project" {
  type = string
}

variable "documents_bucket_arn" {
  type = string
}

variable "executions_table_arn" {
  type    = string
  default = ""
}

variable "state_machine_arn" {
  type    = string
  default = ""
}

variable "extract_text_arn" {
  type    = string
  default = "*"
}

variable "chunk_document_arn" {
  type    = string
  default = "*"
}

variable "documents_table_arn" {
  type    = string
  default = ""
}

variable "chunks_table_arn" {
  type    = string
  default = ""
}

variable "save_metadata_arn" {
  type    = string
  default = "*"
}
