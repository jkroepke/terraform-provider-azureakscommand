# The following example shows how to run the command kubectl cluster-info inside a AKS cluster

resource "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "kubectl cluster-info"
}

output "invoke_output" {
  value = azureakscommand_invoke.this.output
}



# Precondition and Postcondition checks are available with Terraform v1.2.0 and later.
resource "azureakscommand_invoke" "this" {
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

resource "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "cat hello"
  context = filebase64(data.archive_file.context.output_path)
}



# helm is nativly supported.
resource "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "helm repo add bitnami https://charts.bitnami.com/bitnami && helm repo update && helm install my-release bitnami/nginx"
}
