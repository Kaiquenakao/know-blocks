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
