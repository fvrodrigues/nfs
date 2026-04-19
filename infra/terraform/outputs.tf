output "instance_id" {
  description = "EC2 instance ID."
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
  description = "URL for calling /prestador on the EC2 host."
  value       = "http://${aws_instance.nfse.public_dns}:${var.app_port}"
}
