output "doc_validator_invoke_arn" {
  value = aws_lambda_function.doc_validator.invoke_arn
}

output "doc_validator_name" {
  value = aws_lambda_function.doc_validator.function_name
}

output "pipe_trigger_invoke_arn" {
  value = aws_lambda_function.pipe_trigger.invoke_arn
}

output "pipe_trigger_name" {
  value = aws_lambda_function.pipe_trigger.function_name
}
