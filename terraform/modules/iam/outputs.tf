output "doc_validator_role_arn" {
  value = aws_iam_role.doc_validator.arn
}

output "pipe_trigger_role_arn" {
  value = aws_iam_role.pipe_trigger.arn
}
