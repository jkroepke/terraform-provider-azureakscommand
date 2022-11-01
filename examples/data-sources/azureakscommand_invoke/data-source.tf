# The following example shows how to run the command kubectl cluster-info inside a AKS cluster

data "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "kubectl cluster-info"
}

output "invoke_output" {
  value = data.azureakscommand_invoke.this.output
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

data "azureakscommand_invoke" "this" {
  resource_group_name = "rg-default"
  name                = "cluster-name"

  command = "cat hello"
  context = filebase64(data.archive_file.context.output_path)
}

