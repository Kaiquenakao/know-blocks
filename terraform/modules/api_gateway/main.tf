# ── REST API ──────────────────────────────────────────────────────────────────

resource "aws_api_gateway_rest_api" "main" {
  name        = "${var.project}-api"
  description = "Know-Blocks REST API"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  tags = { Project = var.project }
}

# ── /documents ────────────────────────────────────────────────────────────────

resource "aws_api_gateway_resource" "documents" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_rest_api.main.root_resource_id
  path_part   = "documents"
}

# ── /documents/deploy ─────────────────────────────────────────────────────────

resource "aws_api_gateway_resource" "deploy" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  parent_id   = aws_api_gateway_resource.documents.id
  path_part   = "deploy"
}

# GET → doc-validator
resource "aws_api_gateway_method" "deploy_get" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.deploy.id
  http_method   = "GET"
  authorization = "NONE"

  request_parameters = {
    "method.request.querystring.filename" = true
  }
}

resource "aws_api_gateway_integration" "deploy_get" {
  rest_api_id             = aws_api_gateway_rest_api.main.id
  resource_id             = aws_api_gateway_resource.deploy.id
  http_method             = aws_api_gateway_method.deploy_get.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.doc_validator_invoke_arn
}

# POST → pipe-trigger
resource "aws_api_gateway_method" "deploy_post" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  resource_id   = aws_api_gateway_resource.deploy.id
  http_method   = "POST"
  authorization = "NONE"
}

resource "aws_api_gateway_integration" "deploy_post" {
  rest_api_id             = aws_api_gateway_rest_api.main.id
  resource_id             = aws_api_gateway_resource.deploy.id
  http_method             = aws_api_gateway_method.deploy_post.http_method
  integration_http_method = "POST"
  type                    = "AWS_PROXY"
  uri                     = var.pipe_trigger_invoke_arn
}

# ── Deploy da API ─────────────────────────────────────────────────────────────

resource "aws_api_gateway_deployment" "main" {
  rest_api_id = aws_api_gateway_rest_api.main.id

  triggers = {
    redeployment = sha1(jsonencode([
      aws_api_gateway_resource.documents,
      aws_api_gateway_resource.deploy,
      aws_api_gateway_method.deploy_get,
      aws_api_gateway_method.deploy_post,
      aws_api_gateway_integration.deploy_get,
      aws_api_gateway_integration.deploy_post,
    ]))
  }

  lifecycle {
    create_before_destroy = true
  }

  depends_on = [
    aws_api_gateway_integration.deploy_get,
    aws_api_gateway_integration.deploy_post,
  ]
}

resource "aws_api_gateway_stage" "prod" {
  rest_api_id   = aws_api_gateway_rest_api.main.id
  deployment_id = aws_api_gateway_deployment.main.id
  stage_name    = "prod"

  tags = { Project = var.project }
}

# ── Throttling ────────────────────────────────────────────────────────────────

resource "aws_api_gateway_method_settings" "throttle" {
  rest_api_id = aws_api_gateway_rest_api.main.id
  stage_name  = aws_api_gateway_stage.prod.stage_name
  method_path = "*/*"

  settings {
    throttling_burst_limit = 50
    throttling_rate_limit  = 100
  }
}

# ── Permissões ────────────────────────────────────────────────────────────────

resource "aws_lambda_permission" "doc_validator" {
  statement_id  = "AllowAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = var.doc_validator_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.main.execution_arn}/*/GET/documents/deploy"
}

resource "aws_lambda_permission" "pipe_trigger" {
  statement_id  = "AllowAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = var.pipe_trigger_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = "${aws_api_gateway_rest_api.main.execution_arn}/*/POST/documents/deploy"
}
