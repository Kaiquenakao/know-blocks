output "doc_validator_role_arn" {
  value = aws_iam_role.doc_validator.arn
}

output "pipe_trigger_role_arn" {
  value = aws_iam_role.pipe_trigger.arn
}

output "extract_text_role_arn" {
  value = aws_iam_role.extract_text.arn
}

output "chunk_document_role_arn" {
  value = aws_iam_role.chunk_document.arn
}

output "step_functions_role_arn" {
  value = aws_iam_role.step_functions.arn
}

output "ecs_execution_role_arn" {
  value = aws_iam_role.ecs_execution.arn
}

output "ecs_task_role_arn" {
  value = aws_iam_role.ecs_task.arn
}

output "save_metadata_role_arn" {
  value = aws_iam_role.save_metadata.arn
}
