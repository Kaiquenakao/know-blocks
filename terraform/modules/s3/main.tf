resource "aws_s3_bucket" "documents" {
  bucket = "${var.project}-documents-${var.env}"

  tags = {
    Project     = var.project
    Environment = var.env
  }
}

resource "aws_s3_bucket_ownership_controls" "documents" {
  bucket = aws_s3_bucket.documents.id

  rule {
    object_ownership = "BucketOwnerEnforced"
  }
}

resource "aws_s3_bucket_versioning" "documents" {
  bucket = aws_s3_bucket.documents.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "documents" {
  bucket = aws_s3_bucket.documents.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

resource "aws_s3_bucket_public_access_block" "documents" {
  bucket = aws_s3_bucket.documents.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ── S3 Vectors ────────────────────────────────────────────────────────────────


# S3 Vectors são criados via CLI — não têm suporte no Terraform provider ainda
# aws s3vectors create-vector-bucket --vector-bucket-name know-blocks-vectors
# aws s3vectors create-index --vector-bucket-name know-blocks-vectors --index-name global-chunks --data-type float32 --dimension 768 --distance-metric cosine
