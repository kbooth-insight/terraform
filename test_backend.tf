terraform {
  backend "azurerm_vault" {
    keyvault_prefix = "booth"
    keyvault_name = "goldentoe"
  }
}