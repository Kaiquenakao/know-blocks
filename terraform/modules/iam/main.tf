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
    actions   = ["s3:HeadObject"]
    resources = ["${var.documents_bucket_arn}/*"]
  }

  statement {
    effect    = "Allow"
    actions   = ["s3:PutObject"]
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
    effect    = "Allow"
    actions   = ["s3:HeadObject"]
    resources = ["${var.documents_bucket_arn}/*"]
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
