package scalr

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProviderConfigurationParameterCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	configuration, deleteConfiguration := createProviderConfiguration(
		t, client, "kubernetes", "kubernetes_dev",
	)
	defer deleteConfiguration()

	t.Run("success", func(t *testing.T) {
		options := ProviderConfigurationParameterCreateOptions{
			Key:         String("config_path"),
			Sensitive:   Bool(false),
			Value:       String("~/.kube/config"),
			Description: String("A path to a kube config file."),
		}
		parameter, err := client.ProviderConfigurationParameters.Create(ctx, configuration.ID, options)
		if err != nil {
			t.Fatal(err)
		}

		parameter, err = client.ProviderConfigurationParameters.Read(ctx, parameter.ID)
		require.NoError(t, err)

		assert.Equal(t, *options.Key, parameter.Key)
		assert.Equal(t, *options.Sensitive, parameter.Sensitive)
		assert.Equal(t, *options.Value, parameter.Value)
		assert.Equal(t, *options.Description, parameter.Description)
	})

	t.Run("success sensitive", func(t *testing.T) {
		options := ProviderConfigurationParameterCreateOptions{
			Key:       String("client_certificate"),
			Sensitive: Bool(true),
			Value:     String("-----BEGIN CERTIFICATE-----\nMIIB9TCCAWACAQAwgbgxGTAXB"),
		}
		parameter, err := client.ProviderConfigurationParameters.Create(ctx, configuration.ID, options)
		if err != nil {
			t.Fatal(err)
		}

		parameter, err = client.ProviderConfigurationParameters.Read(ctx, parameter.ID)
		require.NoError(t, err)

		assert.Equal(t, *options.Key, parameter.Key)
		assert.Equal(t, *options.Sensitive, parameter.Sensitive)
		assert.Equal(t, "", parameter.Value)
	})
}

func TestProviderConfigurationParametersList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		configuration, removeConfiguration := createProviderConfiguration(
			t, client, "kubernetes", "kubernetes_dev",
		)
		defer removeConfiguration()

		providerTestingDataSet := []struct {
			Key       string
			Value     string
			Sensitive bool
		}{
			{Key: "config_path", Value: "~/.kube/config", Sensitive: false},
			{Key: "config_context", Value: "my-context", Sensitive: false},
			{Key: "client_certificate", Value: "--BEGIN CERTIFICATE--\nMIIB9", Sensitive: true},
		}

		var createdParameterIDs []string
		for _, parameterData := range providerTestingDataSet {
			parameter, err := client.ProviderConfigurationParameters.Create(ctx, configuration.ID, ProviderConfigurationParameterCreateOptions{
				Key:       String(parameterData.Key),
				Sensitive: Bool(parameterData.Sensitive),
				Value:     String(parameterData.Value),
			})
			if err != nil {
				t.Fatal(err)
			}

			createdParameterIDs = append(createdParameterIDs, parameter.ID)
		}

		parametersList, err := client.ProviderConfigurationParameters.List(ctx, configuration.ID, ProviderConfigurationParametersListOptions{})

		require.NoError(t, err)
		assert.Equal(t, 3, len(parametersList.Items))

		var resultIDs []string
		for _, configuration := range parametersList.Items {
			resultIDs = append(resultIDs, configuration.ID)
		}

		for _, parameterID := range createdParameterIDs {
			assert.Contains(t, resultIDs, parameterID)
		}
	})
}

func TestProviderConfigurationParameterUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("success all attributes", func(t *testing.T) {
		configuration, removeConfiguration := createProviderConfiguration(
			t, client, "kubernetes", "kubernetes_dev",
		)
		defer removeConfiguration()

		parameter, err := client.ProviderConfigurationParameters.Create(ctx, configuration.ID, ProviderConfigurationParameterCreateOptions{
			Key:       String("config_context"),
			Sensitive: Bool(false),
			Value:     String("my-context"),
		})
		if err != nil {
			t.Fatal(err)
		}

		options := ProviderConfigurationParameterUpdateOptions{
			Key:         String("config_path"),
			Sensitive:   Bool(true),
			Value:       String("~/.kube/config"),
			Description: String("A path to a kube config file."),
		}
		updatedParameter, err := client.ProviderConfigurationParameters.Update(
			ctx, parameter.ID, options,
		)
		require.NoError(t, err)

		assert.Equal(t, *options.Key, updatedParameter.Key)
		assert.Equal(t, *options.Sensitive, updatedParameter.Sensitive)
		assert.Equal(t, *options.Description, updatedParameter.Description)
	})
}

func TestProviderConfigurationParameterDelete(t *testing.T) {
	client := testClient(t)

	ctx := context.Background()

	configuration, removeConfiguration := createProviderConfiguration(t, client, "kubernetes", "kubernetes_dev")
	defer removeConfiguration()

	t.Run("success", func(t *testing.T) {
		parameter, err := client.ProviderConfigurationParameters.Create(ctx, configuration.ID, ProviderConfigurationParameterCreateOptions{
			Key:       String("config_context"),
			Sensitive: Bool(true),
			Value:     String("my-context"),
		})
		require.NoError(t, err)

		err = client.ProviderConfigurationParameters.Delete(ctx, parameter.ID)
		require.NoError(t, err)

		// Try loading the configuration - it should fail.
		_, err = client.ProviderConfigurationParameters.Read(ctx, parameter.ID)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("ProviderConfigurationParameter with ID '%s' not found or user unauthorized.", parameter.ID),
			}.Error(),
			err.Error(),
		)
	})
}
