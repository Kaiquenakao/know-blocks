output "executions_table_name" {
  value = aws_dynamodb_table.executions.name
}

output "executions_table_arn" {
  value = aws_dynamodb_table.executions.arn
}

output "documents_table_name" {
  value = aws_dynamodb_table.documents.name
}

output "documents_table_arn" {
  value = aws_dynamodb_table.documents.arn
}

output "chunks_table_name" {
  value = aws_dynamodb_table.chunks.name
}

output "chunks_table_arn" {
  value = aws_dynamodb_table.chunks.arn
}
