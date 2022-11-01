# terraform-provider-azureakscommand

Terraform provider for running commands on private AKS clusters without reach them

# Examples

## Simple command execution

```terraform
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}

resource "azurerm_resource_group" "example" {
  name     = "example-resources"
  location = "West Europe"
}

resource "azurerm_kubernetes_cluster" "example" {
  name                = "example-aks1"
  location            = azurerm_resource_group.example.location
  resource_group_name = azurerm_resource_group.example.name

  default_node_pool {
    name       = "default"
    node_count = 1
    vm_size    = "Standard_D2_v2"
  }
}


provider "azureakscommand" {
  tenant_id       = data.azurerm_subscription.current.tenant_id
  subscription_id = data.azurerm_subscription.current.subscription_id
}

resource "azureakscommand_invoke" "example" {
  resource_group_name = azurerm_resource_group.example.name
  name                = azurerm_kubernetes_cluster.example.name

  command = "kubectl cluster-info"

  # Re-run command, if cluster gets recreated.
  triggers = {
    id = azurerm_kubernetes_cluster.example.id
  }

  # Precondition and Postcondition checks are available with Terraform v1.2.0 and later.
  lifecycle {
    postcondition {
      condition     = self.exit_code == 0
      error_message = "exit code invalid"
    }
  }
}

output "invoke_output" {
  value = azureakscommand_invoke.example.output
}
```

## Command execution with additional files

```terraform
provider "azurerm" {
  features {}
}

data "azurerm_subscription" "current" {}


provider "archive" {}

data "archive_file" "manifests" {
  type        = "zip"
  output_path = "${path.module}/manifests.zip"
  source_dir  = "${path.module}/manifests/"
}


provider "azureakscommand" {
  tenant_id       = data.azurerm_subscription.current.tenant_id
  subscription_id = data.azurerm_subscription.current.subscription_id
}

resource "azureakscommand_invoke" "example" {
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

resource "azureakscommand_invoke" "example" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = join("&&", [
    "helm repo add bitnami https://charts.bitnami.com/bitnami",
    "helm repo update",
    "helm install my-release --repo bitnami/nginx"
  ])
}
```
