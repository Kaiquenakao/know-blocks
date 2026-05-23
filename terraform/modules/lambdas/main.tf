# ── doc-validator ─────────────────────────────────────────────────────────────

resource "aws_lambda_function" "doc_validator" {
  function_name = "${var.project}-doc-validator"
  role          = var.doc_validator_role_arn
  runtime       = "provided.al2"
  handler       = "bootstrap"
  filename      = "${path.module}/functions/doc-validator/function.zip"

  timeout     = 15
  memory_size = 128

  environment {
    variables = {
      DOCUMENTS_BUCKET = var.documents_bucket_name
    }
  }

  tags = { Project = var.project }
}

# ── pipe-trigger ──────────────────────────────────────────────────────────────

resource "aws_lambda_function" "pipe_trigger" {
  function_name = "${var.project}-pipe-trigger"
  role          = var.pipe_trigger_role_arn
  runtime       = "provided.al2"
  handler       = "bootstrap"
  filename      = "${path.module}/functions/pipe-trigger/function.zip"

  timeout     = 30
  memory_size = 128

  environment {
    variables = {
      DOCUMENTS_BUCKET  = var.documents_bucket_name
      STATE_MACHINE_ARN = var.state_machine_arn
    }
  }

  tags = { Project = var.project }
}

# ── extract-text ──────────────────────────────────────────────────────────────

resource "aws_lambda_function" "extract_text" {
  function_name = "${var.project}-extract-text"
  role          = var.extract_text_role_arn
  runtime       = "provided.al2"
  handler       = "bootstrap"
  filename      = "${path.module}/functions/extract-text/function.zip"

  timeout     = 300
  memory_size = 512

  environment {
    variables = {
      DOCUMENTS_BUCKET = var.documents_bucket_name
      EXECUTIONS_TABLE = var.executions_table_name
    }
  }

  tags = { Project = var.project }
}

# ── chunk-document ────────────────────────────────────────────────────────────

resource "aws_lambda_function" "chunk_document" {
  function_name = "${var.project}-chunk-document"
  role          = var.chunk_document_role_arn
  runtime       = "provided.al2"
  handler       = "bootstrap"
  filename      = "${path.module}/functions/chunk-document/function.zip"

  timeout     = 300
  memory_size = 512

  environment {
    variables = {
      DOCUMENTS_BUCKET = var.documents_bucket_name
      EXECUTIONS_TABLE = var.executions_table_name
    }
  }

  tags = { Project = var.project }
}
