provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}

provider "azureakscommand" {
  tenant_id       = data.azurerm_subscription.current.tenant_id
  subscription_id = data.azurerm_subscription.current.subscription_id
}
