output "ollama_repository_url" {
  value = aws_ecr_repository.ollama.repository_url
}

output "embed_chunks_repository_url" {
  value = aws_ecr_repository.embed_chunks.repository_url
}

output "ollama_repository_arn" {
  value = aws_ecr_repository.ollama.arn
}

output "embed_chunks_repository_arn" {
  value = aws_ecr_repository.embed_chunks.arn
}
