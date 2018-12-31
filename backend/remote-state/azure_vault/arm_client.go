package azure_vault

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/keyvault/mgmt/keyvault"
	keyvaultSecrets "github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"log"
	"os"
	"time"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform/httpclient"
)

type ArmClient struct {
	vaultsClient      keyvault.VaultsClient
	secretsClient     keyvaultSecrets.BaseClient
	environment       azure.Environment
	resourceGroupName string
	vaultName         string
	subscriptionId    string
}

func buildArmClient(ctx context.Context, config BackendConfig) (*ArmClient, error) {
	env, err := buildArmEnvironment(config)
	if err != nil {
		return nil, err
	}

	client := ArmClient{
		environment:       *env,
		resourceGroupName: config.ResourceGroupName,
		vaultName:         config.keyvaultName,
		subscriptionId:    config.SubscriptionID,
	}

	builder := authentication.Builder{
		ClientID:                      config.ClientID,
		ClientSecret:                  config.ClientSecret,
		SubscriptionID:                config.SubscriptionID,
		TenantID:                      config.TenantID,
		CustomResourceManagerEndpoint: config.CustomResourceManagerEndpoint,
		Environment:                   config.Environment,

		// Feature Toggles
		SupportsAzureCliToken:    true,
		SupportsClientSecretAuth: true,
		// TODO: support for Client Certificate auth
	}
	armConfig, err := builder.Build()
	if err != nil {
		return nil, fmt.Errorf("Error building ARM Config: %+v", err)
	}

	oauthConfig, err := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, armConfig.TenantID)
	if err != nil {
		return nil, err
	}

	auth, err := armConfig.GetAuthorizationToken(oauthConfig, env.TokenAudience)
	if err != nil {
		return nil, err
	}

	vaultsClient := keyvault.NewVaultsClient(armConfig.SubscriptionID)

	client.configureVaultsClient(&vaultsClient.Client, auth)
	client.vaultsClient = vaultsClient

	vault, err := client.vaultsClient.Get(ctx, config.ResourceGroupName, config.keyvaultName)
	if err != nil {
		return nil, err
	}

	fmt.Printf("%s vault found\n", *vault.Name)


	secretClient := keyvaultSecrets.New()
	client.configureVaultsClient(&secretClient.Client, auth)
	client.secretsClient = secretClient
	//maxResults := int32(25)
	fmt.Printf(*vault.Properties.VaultURI)
	t := context.TODO()
	key, err := secretClient.GetSecret(t, *vault.Properties.VaultURI, config.keyvaultPrefix, "")
	if err != nil  {
		return nil, fmt.Errorf("Error fetching secret %s ", config.keyvaultPrefix)
	}

	fmt.Println(key)

	return &client, nil
}

func buildArmEnvironment(config BackendConfig) (*azure.Environment, error) {
	if config.CustomResourceManagerEndpoint != "" {
		log.Printf("[DEBUG] Loading Environment from Endpoint %q", config.CustomResourceManagerEndpoint)
		return authentication.LoadEnvironmentFromUrl(config.CustomResourceManagerEndpoint)
	}

	log.Printf("[DEBUG] Loading Environment %q", config.Environment)
	return authentication.DetermineEnvironment(config.Environment)
}

func (c *ArmClient) configureClient(client *autorest.Client, auth autorest.Authorizer) {
	client.UserAgent = buildUserAgent()
	client.Authorizer = auth
	client.Sender = buildSender()
	client.SkipResourceProviderRegistration = false
	client.PollingDuration = 60 * time.Minute
}

func (c *ArmClient) configureVaultsClient(client *autorest.Client, auth autorest.Authorizer) {
	client.UserAgent = buildUserAgent()
	client.Authorizer = auth
	client.Sender = buildSender()
	client.SkipResourceProviderRegistration = false
	client.PollingDuration = 60 * time.Minute
}

func buildUserAgent() string {
	userAgent := httpclient.UserAgentString()

	// append the CloudShell version to the user agent if it exists
	if azureAgent := os.Getenv("AZURE_HTTP_USER_AGENT"); azureAgent != "" {
		userAgent = fmt.Sprintf("%s %s", userAgent, azureAgent)
	}

	return userAgent
}
