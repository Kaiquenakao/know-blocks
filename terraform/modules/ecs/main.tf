# ── ECS Cluster ───────────────────────────────────────────────────────────────

resource "aws_ecs_cluster" "main" {
  name = "${var.project}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = { Project = var.project }
}

# ── CloudWatch Log Groups ─────────────────────────────────────────────────────

resource "aws_cloudwatch_log_group" "ollama" {
  name              = "/ecs/${var.project}/ollama"
  retention_in_days = 7
}

resource "aws_cloudwatch_log_group" "embed_chunks" {
  name              = "/ecs/${var.project}/embed-chunks"
  retention_in_days = 7
}

# ── Task Definition ───────────────────────────────────────────────────────────

resource "aws_ecs_task_definition" "embed" {
  family                   = "${var.project}-embed-task"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "2048"   # 2 vCPU
  memory                   = "8192"   # 8GB — Ollama precisa de memória
  execution_role_arn       = var.ecs_execution_role_arn
  task_role_arn            = var.ecs_task_role_arn

  container_definitions = jsonencode([

    # ── Container 1: Ollama (sidecar) ─────────────────────────────────────────
    {
      name      = "ollama"
      image     = "${var.ollama_image_url}:latest"
      essential = false   # se ollama parar, não encerra a task imediatamente

      portMappings = [
        {
          containerPort = 11434
          protocol      = "tcp"
        }
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.ollama.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "ollama"
        }
      }

      healthCheck = {
        command     = ["CMD-SHELL", "/bin/ollama list > /dev/null 2>&1 || exit 1"]
        interval    = 15
        timeout     = 10
        retries     = 5
        startPeriod = 60
      }
    },

    # ── Container 2: embed-chunks (main) ──────────────────────────────────────
    {
      name      = "embed-chunks"
      image     = "${var.embed_chunks_image_url}:latest"
      essential = true   # quando terminar, encerra a task

      dependsOn = [
        {
          containerName = "ollama"
          condition     = "HEALTHY"
        }
      ]

      environment = [
        { name = "OLLAMA_URL",       value = "http://localhost:11434" },
        { name = "DOCUMENTS_BUCKET", value = var.documents_bucket_name },
        { name = "EXECUTIONS_TABLE", value = var.executions_table_name },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.embed_chunks.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "embed-chunks"
        }
      }
    }
  ])

  tags = { Project = var.project }
}
