resource "aws_sfn_state_machine" "pipeline" {
  name     = "${var.project}-pipeline"
  role_arn = var.step_functions_role_arn

  definition = templatefile("${path.module}/state_machine.json", {
    extract_text_arn         = var.extract_text_arn
    chunk_document_arn       = var.chunk_document_arn
    ecs_cluster_arn          = var.ecs_cluster_arn
    embed_task_definition_arn = var.embed_task_definition_arn
    subnets                  = jsonencode(var.subnet_ids)
  })

  tags = { Project = var.project }
}
