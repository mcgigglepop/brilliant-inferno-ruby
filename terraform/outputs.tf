output "cognito_user_pool_id" {
  description = "The ID of the Cognito User Pool."
  value       = module.cognito.cognito_user_pool_id
}

output "cognito_client_id" {
  description = "The ID of the Cognito User Pool Client."
  value       = module.cognito.cognito_client_id
}
