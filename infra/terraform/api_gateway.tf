# HTTP API v2 proxying everything to the EC2 instance's public DNS:port.
# Provisioned only when var.enable_api_gateway is true because the primary
# client path is meant to hit the EC2 directly (API Gateway's ~30s
# integration timeout is too short for the /prestador flow).

resource "aws_apigatewayv2_api" "nfse" {
  count         = var.enable_api_gateway ? 1 : 0
  name          = "${var.project}-http-api"
  protocol_type = "HTTP"
  description   = "HTTP API in front of the nfse EC2 instance. Prefer hitting the EC2 directly for /prestador because of the 30s API Gateway integration timeout."
}

resource "aws_apigatewayv2_integration" "nfse" {
  count                  = var.enable_api_gateway ? 1 : 0
  api_id                 = aws_apigatewayv2_api.nfse[0].id
  integration_type       = "HTTP_PROXY"
  integration_method     = "ANY"
  integration_uri        = "http://${aws_instance.nfse.public_dns}:${var.app_port}/{proxy}"
  payload_format_version = "1.0"
  timeout_milliseconds   = 30000
}

resource "aws_apigatewayv2_route" "nfse" {
  count     = var.enable_api_gateway ? 1 : 0
  api_id    = aws_apigatewayv2_api.nfse[0].id
  route_key = "ANY /{proxy+}"
  target    = "integrations/${aws_apigatewayv2_integration.nfse[0].id}"
}

resource "aws_apigatewayv2_stage" "nfse" {
  count       = var.enable_api_gateway ? 1 : 0
  api_id      = aws_apigatewayv2_api.nfse[0].id
  name        = var.api_stage_name
  auto_deploy = true
}
