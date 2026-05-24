output "api_url" {
  value       = module.api_gateway.api_url
  description = "URL base da API — cole no .env como API_BASE_URL"
}

output "state_machine_arn" {
  value       = module.step_functions.state_machine_arn
  description = "pipe-trigger"
}
