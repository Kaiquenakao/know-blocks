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

output "extract_text_arn" {
  value = aws_lambda_function.extract_text.arn
}

output "extract_text_name" {
  value = aws_lambda_function.extract_text.function_name
}

output "chunk_document_arn" {
  value = aws_lambda_function.chunk_document.arn
}

output "chunk_document_name" {
  value = aws_lambda_function.chunk_document.function_name
}

output "save_metadata_arn" {
  value = aws_lambda_function.save_metadata.arn
}

output "save_metadata_name" {
  value = aws_lambda_function.save_metadata.function_name
}
