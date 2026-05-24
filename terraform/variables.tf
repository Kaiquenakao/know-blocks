variable "project" {
  type    = string
  default = "know-blocks"
}

variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "subnet_ids" {
  type        = list(string)
  description = "Subnet IDs da VPC default para o Fargate"
}
