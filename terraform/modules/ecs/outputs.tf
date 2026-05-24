output "cluster_arn" {
  value = aws_ecs_cluster.main.arn
}

output "embed_task_definition_arn" {
  value = aws_ecs_task_definition.embed.arn
}
