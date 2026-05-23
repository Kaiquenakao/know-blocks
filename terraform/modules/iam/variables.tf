variable "project" {
  type = string
}

variable "documents_bucket_arn" {
  type = string
}

variable "state_machine_arn" {
  type    = string
  default = ""
}
