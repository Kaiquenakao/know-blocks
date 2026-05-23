output "api_url" {
  value       = aws_api_gateway_stage.prod.invoke_url
  description = "URL base da API — ex: https://xxx.execute-api.us-east-1.amazonaws.com/prod"
}

output "rest_api_id" {
  value = aws_api_gateway_rest_api.main.id
}
