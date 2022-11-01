# The following example shows how to run the command kubectl cluster-info inside a AKS cluster

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

resource "azureakscommand_invoke" "example" {
  resource_group_name = azurerm_resource_group.example.name
  name                = azurerm_kubernetes_cluster.example.name

  command = "kubectl cluster-info"

  # Re-run command, if cluster gets recreated.
  triggers = {
    id = azurerm_kubernetes_cluster.example.id
  }
}

output "invoke_output" {
  value = azureakscommand_invoke.example.output
}



# Precondition and Postcondition checks are available with Terraform v1.2.0 and later.
resource "azureakscommand_invoke" "example" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "kubectl cluster-info"

  lifecycle {
    postcondition {
      condition     = self.exit_code == 0
      error_message = "exit code invalid"
    }
  }
}



# optional with additional file context
provider "archive" {}

data "archive_file" "context" {
  type        = "zip"
  output_path = "${path.module}/context.zip"

  source {
    content  = "world"
    filename = "hello"
  }
}

resource "azureakscommand_invoke" "example" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "cat hello"
  context = filebase64(data.archive_file.context.output_path)
}



# helm is natively supported.
resource "azureakscommand_invoke" "example" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update && helm install my-release bitnami/nginx"
}
