package azure_vault

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform/backend"
	"github.com/hashicorp/terraform/helper/schema"
)

// New creates a new backend for Azure remote state.
func New() backend.Backend {
	s := &schema.Backend{
		Schema: map[string]*schema.Schema{

			"container_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The container name.",
			},
			"key": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "TODO: not sure if this is the state's internal key in some ways yet.",
			},
			"keyvault_prefix": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The key/cert prefix.",
			},

			"keyvault_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the vault.",
			},

			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Azure cloud environment.",
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENVIRONMENT", "public"),
			},

			"resource_group_name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The resource group name.",
			},

			"client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Client ID.",
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_ID", ""),
			},

			"client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Client Secret.",
				DefaultFunc: schema.EnvDefaultFunc("ARM_CLIENT_SECRET", ""),
			},

			"subscription_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Subscription ID.",
				DefaultFunc: schema.EnvDefaultFunc("ARM_SUBSCRIPTION_ID", ""),
			},

			"tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Tenant ID.",
				DefaultFunc: schema.EnvDefaultFunc("ARM_TENANT_ID", ""),
			},

			"endpoint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A custom Endpoint used to access the Azure Resource Manager API's.",
				DefaultFunc: schema.EnvDefaultFunc("ARM_ENDPOINT", ""),
			},

			// Deprecated fields
			"arm_client_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Client ID.",
				Deprecated:  "`arm_client_id` has been replaced by `client_id`",
			},

			"arm_client_secret": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Client Secret.",
				Deprecated:  "`arm_client_secret` has been replaced by `client_secret`",
			},

			"arm_subscription_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Subscription ID.",
				Deprecated:  "`arm_subscription_id` has been replaced by `subscription_id`",
			},

			"arm_tenant_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The Tenant ID.",
				Deprecated:  "`arm_tenant_id` has been replaced by `tenant_id`",
			},
		},
	}

	result := &Backend{Backend: s}
	result.Backend.ConfigureFunc = result.configure
	return result
}

type Backend struct {
	*schema.Backend
	// The fields below are set from configure
	armClient      *ArmClient
	containerName  string
	keyvaultName   string
	keyvaultPrefix string
	keyName        string
}

type BackendConfig struct {
	// Required
	keyvaultPrefix string
	keyvaultName   string

	// Optional
	ClientID                      string
	ClientSecret                  string
	CustomResourceManagerEndpoint string
	Environment                   string
	ResourceGroupName             string
	SubscriptionID                string
	TenantID                      string
}

func (b *Backend) configure(ctx context.Context) error {
	if b.containerName != "" {
		return nil
	}

	// Grab the resource data
	data := schema.FromContextBackendConfig(ctx)

	b.keyName = data.Get("key").(string)
	b.keyvaultPrefix = data.Get("keyvault_prefix").(string)
	b.keyvaultName = data.Get("keyvault_name").(string)

	// support for previously deprecated fields
	clientId := valueFromDeprecatedField(data, "client_id", "arm_client_id")
	clientSecret := valueFromDeprecatedField(data, "client_secret", "arm_client_secret")
	subscriptionId := valueFromDeprecatedField(data, "subscription_id", "arm_subscription_id")
	tenantId := valueFromDeprecatedField(data, "tenant_id", "arm_tenant_id")

	config := BackendConfig{
		ClientID:                      clientId,
		ClientSecret:                  clientSecret,
		CustomResourceManagerEndpoint: data.Get("endpoint").(string),
		Environment:                   data.Get("environment").(string),
		ResourceGroupName:             data.Get("resource_group_name").(string),
		keyvaultPrefix:                data.Get("keyvault_prefix").(string),
		keyvaultName:                  data.Get("keyvault_name").(string),
		SubscriptionID:                subscriptionId,
		TenantID:                      tenantId,
	}

	if config.keyvaultName == "" {
		return fmt.Errorf("Need the name of vault to use")
	}

	armClient, err := buildArmClient(ctx, config)
	if err != nil {
		return err
	}

	b.armClient = armClient
	return nil
}

func valueFromDeprecatedField(d *schema.ResourceData, key, deprecatedFieldKey string) string {
	v := d.Get(key).(string)

	if v == "" {
		v = d.Get(deprecatedFieldKey).(string)
	}

	return v
}
