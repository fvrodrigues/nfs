output "instance_id" {
  description = "EC2 instance ID. Use with `aws ssm start-session --target <id>` for a shell without SSH."
  value       = aws_instance.nfse.id
}

output "instance_public_ip" {
  description = "Public IPv4 address of the EC2 host."
  value       = aws_instance.nfse.public_ip
}

output "instance_public_dns" {
  description = "Public DNS name of the EC2 host."
  value       = aws_instance.nfse.public_dns
}

output "app_url" {
  description = "Primary URL for calling /prestador. No API Gateway in front, no 30s cap."
  value       = "http://${aws_instance.nfse.public_dns}:${var.app_port}"
}

output "api_gateway_url" {
  description = "HTTP API Gateway invoke URL. null when enable_api_gateway = false."
  value       = var.enable_api_gateway ? aws_apigatewayv2_api.nfse[0].api_endpoint : null
}
