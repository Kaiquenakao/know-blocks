output "documents_bucket_name" {
  value = aws_s3_bucket.documents.bucket
}

output "documents_bucket_arn" {
  value = aws_s3_bucket.documents.arn
}

# S3 Vectors criados via CLI
output "vector_bucket_name" {
  value = "know-blocks-vectors"
}

output "vector_index_name" {
  value = "global-chunks"
}
