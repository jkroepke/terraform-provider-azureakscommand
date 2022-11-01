provider "azureakscommand" {
  # properties can be lookup from environment.
}


# If Azure CLI is used for auth, tenant_id and subscription_id needs to be passed from azurerm provider
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}

provider "azureakscommand" {
  tenant_id       = data.azurerm_subscription.current.tenant_id
  subscription_id = data.azurerm_subscription.current.subscription_id
}
