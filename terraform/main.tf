provider "aws" {
  region = var.region
}

module "cognito" {
  source         = "./modules/cognito"
  user_pool_name = "${var.application_name}-user-pool"
}
