data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

data "aws_iam_policy_document" "sfn_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["states.amazonaws.com"]
    }
  }
}

# ── doc-validator ─────────────────────────────────────────────────────────────

resource "aws_iam_role" "doc_validator" {
  name               = "${var.project}-doc-validator"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "doc_validator" {
  statement {
    effect    = "Allow"
    actions   = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"]
    resources = ["arn:aws:logs:*:*:*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["s3:HeadObject", "s3:PutObject"]
    resources = ["${var.documents_bucket_arn}/*"]
  }
}

resource "aws_iam_role_policy" "doc_validator" {
  name   = "${var.project}-doc-validator-policy"
  role   = aws_iam_role.doc_validator.id
  policy = data.aws_iam_policy_document.doc_validator.json
}

# ── pipe-trigger ──────────────────────────────────────────────────────────────

resource "aws_iam_role" "pipe_trigger" {
  name               = "${var.project}-pipe-trigger"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "pipe_trigger" {
  statement {
    effect    = "Allow"
    actions   = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"]
    resources = ["arn:aws:logs:*:*:*"]
  }
  statement {
    effect  = "Allow"
    actions = ["s3:HeadObject", "s3:GetObject", "s3:ListBucket"]
    resources = [
      var.documents_bucket_arn,
      "${var.documents_bucket_arn}/*"
    ]
  }
  statement {
    effect    = "Allow"
    actions   = ["states:StartExecution"]
    resources = [var.state_machine_arn != "" ? var.state_machine_arn : "*"]
  }
}

resource "aws_iam_role_policy" "pipe_trigger" {
  name   = "${var.project}-pipe-trigger-policy"
  role   = aws_iam_role.pipe_trigger.id
  policy = data.aws_iam_policy_document.pipe_trigger.json
}

# ── extract-text ──────────────────────────────────────────────────────────────

resource "aws_iam_role" "extract_text" {
  name               = "${var.project}-extract-text"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "extract_text" {
  statement {
    effect    = "Allow"
    actions   = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"]
    resources = ["arn:aws:logs:*:*:*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["s3:GetObject", "s3:PutObject"]
    resources = ["${var.documents_bucket_arn}/*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["dynamodb:PutItem"]
    resources = [var.executions_table_arn]
  }
}

resource "aws_iam_role_policy" "extract_text" {
  name   = "${var.project}-extract-text-policy"
  role   = aws_iam_role.extract_text.id
  policy = data.aws_iam_policy_document.extract_text.json
}

# ── chunk-document ────────────────────────────────────────────────────────────

resource "aws_iam_role" "chunk_document" {
  name               = "${var.project}-chunk-document"
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
}

data "aws_iam_policy_document" "chunk_document" {
  statement {
    effect    = "Allow"
    actions   = ["logs:CreateLogGroup", "logs:CreateLogStream", "logs:PutLogEvents"]
    resources = ["arn:aws:logs:*:*:*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["s3:GetObject", "s3:PutObject"]
    resources = ["${var.documents_bucket_arn}/*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["dynamodb:PutItem"]
    resources = [var.executions_table_arn]
  }
}

resource "aws_iam_role_policy" "chunk_document" {
  name   = "${var.project}-chunk-document-policy"
  role   = aws_iam_role.chunk_document.id
  policy = data.aws_iam_policy_document.chunk_document.json
}

# ── step-functions ────────────────────────────────────────────────────────────

resource "aws_iam_role" "step_functions" {
  name               = "${var.project}-step-functions"
  assume_role_policy = data.aws_iam_policy_document.sfn_assume_role.json
}

data "aws_iam_policy_document" "step_functions" {
  statement {
    effect    = "Allow"
    actions   = ["lambda:InvokeFunction"]
    resources = [
      var.extract_text_arn,
      var.chunk_document_arn,
    ]
  }

  statement {
    effect  = "Allow"
    actions = ["ecs:RunTask", "ecs:StopTask", "ecs:DescribeTasks"]
    resources = ["*"]
  }

  statement {
    effect    = "Allow"
    actions   = ["iam:PassRole"]
    resources = ["*"]
  }

  statement {
    effect = "Allow"
    actions = [
      "events:PutTargets",
      "events:PutRule",
      "events:DescribeRule",
      "events:DeleteRule",
      "events:RemoveTargets",
    ]
    resources = ["arn:aws:events:*:*:rule/StepFunctionsGetEventsForECSTaskRule"]
  }
}

resource "aws_iam_role_policy" "step_functions" {
  name   = "${var.project}-step-functions-policy"
  role   = aws_iam_role.step_functions.id
  policy = data.aws_iam_policy_document.step_functions.json
}

# ── ECS execution role (pull imagem + logs) ───────────────────────────────────

data "aws_iam_policy_document" "ecs_assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "ecs_execution" {
  name               = "${var.project}-ecs-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
}

resource "aws_iam_role_policy_attachment" "ecs_execution" {
  role       = aws_iam_role.ecs_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

# ── ECS task role (S3 + DynamoDB) ────────────────────────────────────────────

resource "aws_iam_role" "ecs_task" {
  name               = "${var.project}-ecs-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume_role.json
}

data "aws_iam_policy_document" "ecs_task" {
  statement {
    effect    = "Allow"
    actions   = ["s3:GetObject", "s3:PutObject", "s3:DeleteObject"]
    resources = ["${var.documents_bucket_arn}/*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["dynamodb:PutItem", "dynamodb:GetItem"]
    resources = [var.executions_table_arn]
  }
}

resource "aws_iam_role_policy" "ecs_task" {
  name   = "${var.project}-ecs-task-policy"
  role   = aws_iam_role.ecs_task.id
  policy = data.aws_iam_policy_document.ecs_task.json
}

# ── Step Functions — adiciona permissão para acionar ECS ─────────────────────

resource "aws_iam_role_policy" "step_functions_ecs" {
  name = "${var.project}-step-functions-ecs-policy"
  role = aws_iam_role.step_functions.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["ecs:RunTask", "ecs:StopTask", "ecs:DescribeTasks"]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = ["iam:PassRole"]
        Resource = "*"
      }
    ]
  })
}
