# terraform-provider-azureakscommand

Terraform provider for running commands on private AKS clusters without reach them

# Examples

## Simple command execution

```terraform
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}

provider "azureakscommand" {
  tenant_id       = data.azurerm_subscription.current.tenant_id
  subscription_id = data.azurerm_subscription.current.subscription_id
}

resource "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "kubectl cluster-info"
  
  # Precondition and Postcondition checks are available with Terraform v1.2.0 and later.
  lifecycle {
    postcondition {
      condition     = self.exit_code == 0
      error_message = "exit code invalid"
    }
  }
}

output "invoke_output" {
  value = azureakscommand_invoke.this.output
}
```

## Command execution with additional files

```terraform
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}

provider "azureakscommand" {
  tenant_id       = data.azurerm_subscription.current.tenant_id
  subscription_id = data.azurerm_subscription.current.subscription_id
}

provider "archive" {}

data "archive_file" "manifests" {
  type        = "zip"
  output_path = "${path.module}/manifests.zip"
  source_dir  = "${path.module}/manifests/"
}

resource "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "kubectl apply -f manifests/"
  context = filebase64(data.archive_file.manifests.output_path)
}
```

## Helm

```terraform
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}

provider "azureakscommand" {
  tenant_id       = data.azurerm_subscription.current.tenant_id
  subscription_id = data.azurerm_subscription.current.subscription_id
}

resource "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = join("&&", [
    "helm install my-release --repo https://charts.bitnami.com/bitnami nginx"
  ])
}
```
