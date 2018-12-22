package azure_vault

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/resources/mgmt/resources"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/hashicorp/go-azure-helpers/authentication"
	"github.com/hashicorp/terraform/httpclient"
)

type ArmClient struct {
	// These Clients are only initialized if an Access Key isn't provided
	groupsClient      *resources.GroupsClient
	vaultsClient      *keyvault.BaseClient
	environment       azure.Environment
	resourceGroupName string
	vaultName         string
}

func buildArmClient(config BackendConfig) (*ArmClient, error) {
	env, err := buildArmEnvironment(config)
	if err != nil {
		return nil, err
	}

	client := ArmClient{
		environment:        *env,
		resourceGroupName:  config.ResourceGroupName,
		vaultName: config.VaultName,

	}

	builder := authentication.Builder{
		ClientID:                      config.ClientID,
		ClientSecret:                  config.ClientSecret,
		SubscriptionID:                config.SubscriptionID,
		TenantID:                      config.TenantID,
		CustomResourceManagerEndpoint: config.CustomResourceManagerEndpoint,
		Environment:                   config.Environment,

		// Feature Toggles
		SupportsAzureCliToken:          true,
		SupportsClientSecretAuth:       true,
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

	//TODO: implement vault client

	groupsClient := resources.NewGroupsClientWithBaseURI(env.ResourceManagerEndpoint, armConfig.SubscriptionID)
	client.configureClient(&groupsClient.Client, auth)
	client.groupsClient = &groupsClient

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

func (c ArmClient) getBlobClient(ctx context.Context) (*keyvault.BaseClient, error) {


	log.Printf("[DEBUG] Building the Blob Client from an Access Token (using user credentials)")
//TODO LIST VAULT KEYS
//	if keys.Keys == nil {
//		return nil, fmt.Errorf("Nil key returned for storage account %q", c.storageAccountName)
//	}

	client =: keyvault.BaseClient()
	//accessKeys := *keys.Keys
	//accessKey := accessKeys[0].Value

	//storageClient, err := storage.NewBasicClientOnSovereignCloud(c.storageAccountName, *accessKey, c.environment)
	//if err != nil {
	//	return nil, fmt.Errorf("Error creating storage client for storage account %q: %s", c.storageAccountName, err)
	//}
	//client := storageClient.GetBlobService()
	return &client, nil
}

func (c *ArmClient) configureClient(client *autorest.Client, auth autorest.Authorizer) {
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
