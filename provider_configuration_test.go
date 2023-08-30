package scalr

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getGoogleTestingCreds(t *testing.T) (credentials, project string) {
	credentials = os.Getenv("TEST_GOOGLE_CREDENTIALS")
	project = os.Getenv("TEST_GOOGLE_PROJECT")
	if len(credentials) == 0 ||
		len(project) == 0 {
		t.Skip("Please set TEST_GOOGLE_CREDENTIALS, TEST_GOOGLE_PROJECT env variables to run this test.")
	}
	return
}

func getAzureTestingCreds(t *testing.T) (armClientId, armClientSecret, armSubscriptionId, armTenantId string) {
	armClientId = os.Getenv("TEST_ARM_CLIENT_ID")
	armClientSecret = os.Getenv("TEST_ARM_CLIENT_SECRET")
	armSubscriptionId = os.Getenv("TEST_ARM_SUBSCRIPTION_ID")
	armTenantId = os.Getenv("TEST_ARM_TENANT_ID")
	if len(armClientId) == 0 ||
		len(armClientSecret) == 0 ||
		len(armSubscriptionId) == 0 ||
		len(armTenantId) == 0 {
		t.Skip("Please set TEST_ARM_CLIENT_ID, TEST_ARM_CLIENT_SECRET, TEST_ARM_SUBSCRIPTION_ID and TEST_ARM_TENANT_ID env variables to run this test.")
	}
	return
}

func getAwsTestingCreds(t *testing.T) (accessKeyId, secretAccessKey, roleArn, externalId string) {
	accessKeyId = os.Getenv("TEST_AWS_ACCESS_KEY")
	secretAccessKey = os.Getenv("TEST_AWS_SECRET_KEY")
	roleArn = os.Getenv("TEST_AWS_ROLE_ARN")
	externalId = os.Getenv("TEST_AWS_EXTERNAL_ID")
	if len(accessKeyId) == 0 ||
		len(secretAccessKey) == 0 ||
		len(roleArn) == 0 ||
		len(externalId) == 0 {
		t.Skip("Please set TEST_AWS_ACCESS_KEY, TEST_AWS_SECRET_KEY, TEST_AWS_ROLE_ARN and TEST_AWS_EXTERNAL_ID env variables to run this test.")
	}
	return
}

func TestProviderConfigurationCreateAzurerm(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	armClientId, armClientSecret, armSubscriptionId, armTenantId := getAzureTestingCreds(t)

	t.Run("success azurerm", func(t *testing.T) {
		options := ProviderConfigurationCreateOptions{
			Account:               &Account{ID: defaultAccountID},
			Name:                  String("azurerm_dev"),
			ProviderName:          String("azurerm"),
			ExportShellVariables:  Bool(false),
			AzurermClientId:       String(armClientId),
			AzurermClientSecret:   String(armClientSecret),
			AzurermSubscriptionId: String(armSubscriptionId),
			AzurermTenantId:       String(armTenantId),
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Equal(t, *options.ExportShellVariables, pcfg.ExportShellVariables)
		assert.Equal(t, *options.AzurermClientId, pcfg.AzurermClientId)
		assert.Equal(t, "", pcfg.AzurermClientSecret)
		assert.Equal(t, *options.AzurermSubscriptionId, pcfg.AzurermSubscriptionId)
		assert.Equal(t, *options.AzurermTenantId, pcfg.AzurermTenantId)
	})
}

func TestProviderConfigurationCreateScalr(t *testing.T) {
	client := testClient(t)
	scalrHostname := client.baseURL.Host
	scalrToken := client.token
	ctx := context.Background()

	t.Run("success scalr", func(t *testing.T) {
		options := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("scalr_dev"),
			ProviderName:         String("scalr"),
			ExportShellVariables: Bool(false),
			ScalrHostname:        String(scalrHostname),
			ScalrToken:           String(scalrToken),
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Equal(t, *options.ExportShellVariables, pcfg.ExportShellVariables)
		assert.Equal(t, *options.ScalrHostname, pcfg.ScalrHostname)
		assert.Equal(t, "", pcfg.ScalrToken)
	})
}

func TestProviderConfigurationCreateAws(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	accessKeyId, secretAccessKey, roleArn, externalId := getAwsTestingCreds(t)

	t.Run("success aws access keys auth", func(t *testing.T) {
		options := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("AWS_dev_account_us_east_1"),
			ProviderName:         String("aws"),
			ExportShellVariables: Bool(false),
			AwsAccessKey:         String(accessKeyId),
			AwsSecretKey:         String(secretAccessKey),
			AwsAccountType:       String("regular"),
			AwsCredentialsType:   String("access_keys"),
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Equal(t, *options.ExportShellVariables, pcfg.ExportShellVariables)
		assert.Equal(t, *options.AwsAccessKey, pcfg.AwsAccessKey)
		assert.Equal(t, *options.AwsAccountType, pcfg.AwsAccountType)
		assert.Equal(t, *options.AwsCredentialsType, pcfg.AwsCredentialsType)
		assert.Equal(t, "", pcfg.AwsSecretKey)
	})

	t.Run("success aws role delegation auth service entity", func(t *testing.T) {
		options := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("AWS_dev_account_us_east_1"),
			ProviderName:         String("aws"),
			ExportShellVariables: Bool(false),
			AwsAccountType:       String("regular"),
			AwsCredentialsType:   String("role_delegation"),
			AwsTrustedEntityType: String("aws_service"),
			AwsRoleArn:           String(roleArn),
			AwsExternalId:        String(externalId),
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Equal(t, *options.ExportShellVariables, pcfg.ExportShellVariables)
		assert.Equal(t, *options.AwsAccountType, pcfg.AwsAccountType)
		assert.Equal(t, *options.AwsCredentialsType, pcfg.AwsCredentialsType)
		assert.Equal(t, *options.AwsTrustedEntityType, pcfg.AwsTrustedEntityType)
		assert.Equal(t, *options.AwsRoleArn, pcfg.AwsRoleArn)
		assert.Equal(t, *options.AwsExternalId, pcfg.AwsExternalId)

	})

	t.Run("success aws role delegation auth account entity", func(t *testing.T) {
		options := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("AWS_dev_account_us_east_1"),
			ProviderName:         String("aws"),
			ExportShellVariables: Bool(false),
			AwsAccountType:       String("regular"),
			AwsCredentialsType:   String("role_delegation"),
			AwsTrustedEntityType: String("aws_account"),
			AwsAccessKey:         String(accessKeyId),
			AwsSecretKey:         String(secretAccessKey),
			AwsRoleArn:           String(roleArn),
			AwsExternalId:        String(externalId),
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Equal(t, *options.ExportShellVariables, pcfg.ExportShellVariables)
		assert.Equal(t, *options.AwsAccountType, pcfg.AwsAccountType)
		assert.Equal(t, *options.AwsCredentialsType, pcfg.AwsCredentialsType)
		assert.Equal(t, *options.AwsTrustedEntityType, pcfg.AwsTrustedEntityType)
		assert.Equal(t, *options.AwsAccessKey, pcfg.AwsAccessKey)
		assert.Equal(t, "", pcfg.AwsSecretKey)
		assert.Equal(t, *options.AwsRoleArn, pcfg.AwsRoleArn)
		assert.Equal(t, *options.AwsExternalId, pcfg.AwsExternalId)
	})
}

func TestProviderConfigurationCreateGoogle(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	credentials, project := getGoogleTestingCreds(t)

	t.Run("success google", func(t *testing.T) {
		options := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("google_dev_project"),
			ProviderName:         String("google"),
			ExportShellVariables: Bool(false),
			IsCustom:             Bool(false),
			GoogleProject:        String(project),
			GoogleCredentials:    String(credentials),
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Equal(t, *options.ExportShellVariables, pcfg.ExportShellVariables)
		assert.Equal(t, *options.GoogleProject, pcfg.GoogleProject)
		assert.Equal(t, "", pcfg.GoogleCredentials)
		assert.Equal(t, "service-account-key", pcfg.GoogleAuthType)
	})
}

func TestProviderConfigurationCreateWithLinkage(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		environment, deleteEnvironment := createEnvironment(t, client)
		defer deleteEnvironment()

		options := ProviderConfigurationCreateOptions{
			Account:      &Account{ID: defaultAccountID},
			Name:         String("consul_dev"),
			ProviderName: String("consul"),
			Environments: []*Environment{environment},
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Len(t, pcfg.Environments, 1)
		assert.Equal(t, environment.ID, pcfg.Environments[0].ID)
	})
}

func TestProviderConfigurationCreateShared(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		options := ProviderConfigurationCreateOptions{
			Account:      &Account{ID: defaultAccountID},
			Name:         String("consul_dev"),
			ProviderName: String("consul"),
			IsShared:     Bool(true),
		}
		pcfg, err := client.ProviderConfigurations.Create(ctx, options)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, pcfg.ID)

		pcfg, err = client.ProviderConfigurations.Read(ctx, pcfg.ID)
		require.NoError(t, err)

		assert.Equal(t, options.Account.ID, pcfg.Account.ID)
		assert.Equal(t, *options.Name, pcfg.Name)
		assert.Equal(t, *options.ProviderName, pcfg.ProviderName)
		assert.Equal(t, *options.IsShared, pcfg.IsShared)
	})
}

func TestProviderConfigurationRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("with parameters", func(t *testing.T) {
		configuration, removeConfiguration := createProviderConfiguration(
			t, client, "kubernetes", "kubernetes_dev",
		)
		defer removeConfiguration()

		optionsList := []ProviderConfigurationParameterCreateOptions{
			{
				Key:         String("config_path"),
				Sensitive:   Bool(false),
				Value:       String("~/.kube/config"),
				Description: String("A path to a kube config file."),
			},
			{
				Key:       String("client_certificate"),
				Sensitive: Bool(true),
				Value:     String("-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxGTAXB"),
			},
			{
				Key:   String("host"),
				Value: String("my-host"),
			},
		}
		for _, option := range optionsList {
			_, err := client.ProviderConfigurationParameters.Create(
				ctx, configuration.ID, option,
			)
			require.NoError(t, err)
		}

		configuration, err := client.ProviderConfigurations.Read(ctx, configuration.ID)
		require.NoError(t, err)
		assert.Equal(t, len(optionsList), len(configuration.Parameters))

		includedParameters := make(map[string]ProviderConfigurationParameter)
		for _, p := range configuration.Parameters {
			includedParameters[p.Key] = *p
		}

		for _, option := range optionsList {
			includedParameter := includedParameters[*option.Key]
			assert.Equal(t, *option.Key, includedParameter.Key)
			var description string
			if option.Description != nil {
				description = *option.Description
			}
			assert.Equal(t, description, includedParameter.Description)
			assert.Equal(t, option.Sensitive != nil && *option.Sensitive, includedParameter.Sensitive)
		}
	})
}

func TestProviderConfigurationList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("filtering", func(t *testing.T) {
		type providerTestingData struct {
			Name         string
			ProviderName string
		}
		providerTestingDataSet := []providerTestingData{
			{Name: "kubernetes_prod_us_east_1", ProviderName: "kubernetes"},
			{Name: "kubernetes_prod_us_east_2", ProviderName: "kubernetes"},
			{Name: "kubernetes_dev_us_east1", ProviderName: "kubernetes"},
			{Name: "consul_prod_us_west1_b", ProviderName: "consul"},
		}

		for _, providerData := range providerTestingDataSet {
			_, removeConfiguration := createProviderConfiguration(
				t, client, providerData.ProviderName, providerData.Name,
			)

			defer removeConfiguration()
		}

		requestOptions := ProviderConfigurationsListOptions{
			Filter: &ProviderConfigurationFilter{
				ProviderName: "kubernetes",
				Name:         "like:_prod_",
			},
		}
		configurationsList, err := client.ProviderConfigurations.List(ctx, requestOptions)

		require.NoError(t, err)
		assert.Equal(t, 2, len(configurationsList.Items))

		var resultNames []string
		for _, configuration := range configurationsList.Items {
			resultNames = append(resultNames, configuration.Name)
		}
		assert.Contains(t, resultNames, "kubernetes_prod_us_east_1")
		assert.Contains(t, resultNames, "kubernetes_prod_us_east_2")
	})
}

func TestProviderConfigurationUpdateAzurerm(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	armClientId, armClientSecret, armSubscriptionId, armTenantId := getAzureTestingCreds(t)

	t.Run("success", func(t *testing.T) {
		createOptions := ProviderConfigurationCreateOptions{
			Account:               &Account{ID: defaultAccountID},
			Name:                  String("azurerm_dev"),
			ProviderName:          String("azurerm"),
			ExportShellVariables:  Bool(false),
			AzurermClientId:       String(armClientId),
			AzurermClientSecret:   String(armClientSecret),
			AzurermSubscriptionId: String(armSubscriptionId),
			AzurermTenantId:       String(armTenantId),
		}
		configuration, err := client.ProviderConfigurations.Create(ctx, createOptions)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, configuration.ID)

		updateOptions := ProviderConfigurationUpdateOptions{
			Name:                  String("azurerm_dev_updated"),
			ExportShellVariables:  Bool(true),
			AzurermClientId:       String(armClientId),
			AzurermClientSecret:   String(armClientSecret),
			AzurermSubscriptionId: String(armSubscriptionId),
			AzurermTenantId:       String(armTenantId),
		}

		updatedConfiguration, err := client.ProviderConfigurations.Update(
			ctx, configuration.ID, updateOptions,
		)
		require.NoError(t, err)
		assert.Equal(t, *updateOptions.Name, updatedConfiguration.Name)
		assert.Equal(t, *updateOptions.ExportShellVariables, updatedConfiguration.ExportShellVariables)
		assert.Equal(t, *updateOptions.AzurermClientId, updatedConfiguration.AzurermClientId)
		assert.Equal(t, "", updatedConfiguration.AzurermClientSecret)
		assert.Equal(t, *updateOptions.AzurermSubscriptionId, updatedConfiguration.AzurermSubscriptionId)
		assert.Equal(t, *updateOptions.AzurermTenantId, updatedConfiguration.AzurermTenantId)
	})

}
func TestProviderConfigurationUpdateAws(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	accessKeyId, secretAccessKey, roleArn, externalId := getAwsTestingCreds(t)

	t.Run("success aws", func(t *testing.T) {
		createOptions := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("AWS_dev_account_us_east_1"),
			ProviderName:         String("aws"),
			ExportShellVariables: Bool(false),
			AwsAccountType:       String("regular"),
			AwsCredentialsType:   String("role_delegation"),
			AwsTrustedEntityType: String("aws_service"),
			AwsRoleArn:           String(roleArn),
			AwsExternalId:        String(externalId),
		}
		configuration, err := client.ProviderConfigurations.Create(ctx, createOptions)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, configuration.ID)

		updateOptions := ProviderConfigurationUpdateOptions{
			Name:                 String("aws_dev_us_east_2"),
			ExportShellVariables: Bool(true),
			AwsAccountType:       String("regular"),
			AwsCredentialsType:   String("role_delegation"),
			AwsTrustedEntityType: String("aws_account"),
			AwsAccessKey:         String(accessKeyId),
			AwsSecretKey:         String(secretAccessKey),
			AwsRoleArn:           String(roleArn),
			AwsExternalId:        String(externalId),
		}
		updatedConfiguration, err := client.ProviderConfigurations.Update(
			ctx, configuration.ID, updateOptions,
		)
		require.NoError(t, err)
		assert.Equal(t, *updateOptions.Name, updatedConfiguration.Name)
		assert.Equal(t, *updateOptions.ExportShellVariables, updatedConfiguration.ExportShellVariables)
		assert.Equal(t, *updateOptions.AwsCredentialsType, updatedConfiguration.AwsCredentialsType)
		assert.Equal(t, *updateOptions.AwsTrustedEntityType, updatedConfiguration.AwsTrustedEntityType)
		assert.Equal(t, *updateOptions.AwsAccessKey, updatedConfiguration.AwsAccessKey)
		assert.Equal(t, *updateOptions.AwsRoleArn, updatedConfiguration.AwsRoleArn)
		assert.Equal(t, *updateOptions.AwsExternalId, updatedConfiguration.AwsExternalId)
	})
}

func TestProviderConfigurationUpdateGoogle(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	credentials, project := getGoogleTestingCreds(t)

	t.Run("success google", func(t *testing.T) {
		createOptions := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("google_dev_project"),
			ProviderName:         String("google"),
			ExportShellVariables: Bool(false),
			GoogleProject:        String(project),
			GoogleCredentials:    String(credentials),
		}
		configuration, err := client.ProviderConfigurations.Create(ctx, createOptions)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, configuration.ID)

		updateOptions := ProviderConfigurationUpdateOptions{
			Name:                 String("google_dev2"),
			ExportShellVariables: Bool(true),
			GoogleProject:        String(project),
			GoogleCredentials:    String(credentials),
		}
		updatedConfiguration, err := client.ProviderConfigurations.Update(
			ctx, configuration.ID, updateOptions,
		)
		require.NoError(t, err)
		assert.Equal(t, *updateOptions.Name, updatedConfiguration.Name)
		assert.Equal(t, *updateOptions.ExportShellVariables, updatedConfiguration.ExportShellVariables)
		assert.Equal(t, *updateOptions.GoogleProject, updatedConfiguration.GoogleProject)
		assert.Equal(t, "", updatedConfiguration.GoogleCredentials)
	})
}

func TestProviderConfigurationUpdateScalr(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()
	scalrHostname := client.baseURL.Host
	scalrToken := client.token

	environment, deleteEnvironment := createEnvironment(t, client)
	defer deleteEnvironment()

	t.Run("success scalr", func(t *testing.T) {
		createOptions := ProviderConfigurationCreateOptions{
			Account:              &Account{ID: defaultAccountID},
			Name:                 String("scalr_dev"),
			ProviderName:         String("scalr"),
			ExportShellVariables: Bool(false),
			ScalrHostname:        String(scalrHostname),
			ScalrToken:           String(scalrToken),
			IsShared:             Bool(false),
			Environments:         []*Environment{environment},
		}
		configuration, err := client.ProviderConfigurations.Create(ctx, createOptions)
		if err != nil {
			t.Fatal(err)
		}
		defer client.ProviderConfigurations.Delete(ctx, configuration.ID)

		updateOptions := ProviderConfigurationUpdateOptions{
			Name:                 String("scalr_prod"),
			ExportShellVariables: Bool(true),
			ScalrHostname:        String(scalrHostname + "/"),
			ScalrToken:           String(scalrToken),
			IsShared:             Bool(true),
			Environments:         []*Environment{},
		}
		updatedConfiguration, err := client.ProviderConfigurations.Update(
			ctx, configuration.ID, updateOptions,
		)
		require.NoError(t, err)
		assert.Equal(t, *updateOptions.Name, updatedConfiguration.Name)
		assert.Equal(t, *updateOptions.ExportShellVariables, updatedConfiguration.ExportShellVariables)
		assert.Equal(t, *updateOptions.ScalrHostname, updatedConfiguration.ScalrHostname)
		assert.Equal(t, *updateOptions.IsShared, updatedConfiguration.IsShared)
	})
}

func TestProviderConfigurationDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	configuration, _ := createProviderConfiguration(t, client, "kubernetes", "kubernetes1")

	t.Run("success", func(t *testing.T) {
		err := client.ProviderConfigurations.Delete(ctx, configuration.ID)
		require.NoError(t, err)

		// Try loading the configuration - it should fail.
		_, err = client.ProviderConfigurations.Read(ctx, configuration.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ProviderConfiguration with ID '%s' not found or user unauthorized.", configuration.ID),
			}.Error(),
			err.Error(),
		)
	})
}
