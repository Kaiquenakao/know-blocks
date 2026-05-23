resource "aws_dynamodb_table" "executions" {
  name         = "${var.project}-executions"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "execution_id"

  attribute {
    name = "execution_id"
    type = "S"
  }

  ttl {
    attribute_name = "expires_at"
    enabled        = true
  }

  tags = { Project = var.project }
}

resource "aws_dynamodb_table" "documents" {
  name         = "${var.project}-documents"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "document_id"

  attribute {
    name = "document_id"
    type = "S"
  }

  tags = { Project = var.project }
}

resource "aws_dynamodb_table" "chunks" {
  name         = "${var.project}-chunks"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "chunk_id"
  range_key    = "document_id"

  attribute {
    name = "chunk_id"
    type = "S"
  }

  attribute {
    name = "document_id"
    type = "S"
  }

  tags = { Project = var.project }
}
