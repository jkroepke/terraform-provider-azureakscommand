provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}

provider "azureakscommand" {
  subscription_id = data.azurerm_subscription.current.subscription_id
}

provider "archive" {}

terraform {
  required_version = ">=0.13.0"
  required_providers {
    azureakscommand = {
      version = "0.0.1"
      source  = "terraform.example.com/local/azureakscommand"
    }
  }
}

resource "azureakscommand_invoke" "this" {
  count = 5

  resource_group_name = "default"
  name                = "jok"
  command             = "exit ${count.index}"
}

data "archive_file" "hello" {
  type        = "zip"
  output_path = "${path.module}/hello.zip"

  source {
    content  = "world"
    filename = "hello"
  }
}

resource "azureakscommand_invoke" "world" {
  resource_group_name = "default"
  name                = "jok"
  command             = "cat hello"
  context             = filebase64(data.archive_file.hello.output_path)
}

output "this" {
  value = azureakscommand_invoke.this[*].exit_code
}

output "notfund" {
  value = azureakscommand_invoke.world.output
}
