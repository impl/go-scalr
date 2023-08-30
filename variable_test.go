package scalr

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVariablesCreate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	wsTest, wsTestCleanup := createWorkspace(t, client, nil)
	account := &Account{ID: defaultAccountID}

	defer wsTestCleanup()

	t.Run("when options has an empty string value", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:         String(randomVariableKey(t)),
			Value:       String(""),
			Category:    Category(CategoryShell),
			Description: String("random variable test"),
			Workspace:   wsTest,
		}

		v, err := client.Variables.Create(ctx, options)
		require.NoError(t, err)

		assert.NotEmpty(t, v.ID)
		assert.Equal(t, *options.Key, v.Key)
		assert.Equal(t, *options.Value, v.Value)
		assert.Equal(t, *options.Category, v.Category)
		assert.Equal(t, *options.Description, v.Description)
	})

	t.Run("when options is missing value", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:       String(randomVariableKey(t)),
			Category:  Category(CategoryShell),
			Workspace: wsTest,
		}

		v, err := client.Variables.Create(ctx, options)
		require.NoError(t, err)

		assert.NotEmpty(t, v.ID)
		assert.Equal(t, *options.Key, v.Key)
		assert.Equal(t, "", v.Value)
		assert.Equal(t, *options.Category, v.Category)
	})

	t.Run("when options is missing key", func(t *testing.T) {
		options := VariableCreateOptions{
			Value:     String(randomString(t)),
			Category:  Category(CategoryShell),
			Workspace: wsTest,
		}

		_, err := client.Variables.Create(ctx, options)
		assert.EqualError(t, err, "key is required")
	})

	t.Run("when options has an empty key", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:       String(""),
			Value:     String(randomString(t)),
			Category:  Category(CategoryShell),
			Workspace: wsTest,
		}

		_, err := client.Variables.Create(ctx, options)
		assert.EqualError(t, err, "key is required")
	})

	t.Run("when options is missing category", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:       String(randomVariableKey(t)),
			Value:     String(randomString(t)),
			Workspace: wsTest,
		}

		_, err := client.Variables.Create(ctx, options)
		assert.EqualError(t, err, "category is required")
	})

	t.Run("when options is missing account", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:         String(randomVariableKey(t)),
			Value:       String(randomString(t)),
			Category:    Category(CategoryShell),
			Environment: wsTest.Environment,
			Workspace:   wsTest,
		}

		_, err := client.Variables.Create(ctx, options)
		require.NoError(t, err)
	})

	t.Run("when options is missing environment", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:       String(randomVariableKey(t)),
			Value:     String(randomString(t)),
			Category:  Category(CategoryShell),
			Account:   account,
			Workspace: wsTest,
		}

		_, err := client.Variables.Create(ctx, options)
		require.NoError(t, err)
	})

	t.Run("when options is missing workspace", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:         String(randomVariableKey(t)),
			Value:       String(randomString(t)),
			Category:    Category(CategoryShell),
			Account:     account,
			Environment: wsTest.Environment,
		}

		_, err := client.Variables.Create(ctx, options)
		require.NoError(t, err)
	})

	t.Run("when options is missing account, environment, workspace", func(t *testing.T) {
		options := VariableCreateOptions{
			Key:       String(randomVariableKey(t)),
			Value:     String(randomString(t)),
			Category:  Category(CategoryShell),
			Workspace: wsTest,
		}

		v, err := client.Variables.Create(ctx, options)
		require.NoError(t, err)

		assert.NotEmpty(t, v.ID)
		assert.Equal(t, *options.Key, v.Key)
		assert.Equal(t, *options.Value, v.Value)
		assert.Equal(t, *options.Category, v.Category)
		assert.Equal(t, options.Workspace.ID, v.Workspace.ID)
	})

}

func TestVariablesRead(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vTest, vTestCleanup := createVariable(t, client, nil, nil, &Account{ID: defaultAccountID})
	defer vTestCleanup()

	t.Run("when the variable exists", func(t *testing.T) {
		v, err := client.Variables.Read(ctx, vTest.ID)
		require.NoError(t, err)
		assert.Equal(t, vTest.ID, v.ID)
		assert.Equal(t, vTest.Category, v.Category)
		assert.Equal(t, vTest.Description, v.Description)
		assert.Equal(t, vTest.HCL, v.HCL)
		assert.Equal(t, vTest.Key, v.Key)
		assert.Equal(t, vTest.Sensitive, v.Sensitive)
		assert.Equal(t, vTest.Value, v.Value)
	})

	t.Run("when the variable does not exist", func(t *testing.T) {
		var variableId = "nonexisting"
		v, err := client.Variables.Read(ctx, variableId)
		assert.Nil(t, v)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Variable with ID '%s' not found or user unauthorized.", variableId),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("without a valid variable ID", func(t *testing.T) {
		v, err := client.Variables.Read(ctx, badIdentifier)
		assert.Nil(t, v)
		assert.EqualError(t, err, "invalid value for variable ID")
	})
}

func TestVariablesUpdate(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	vTest, vTestCleanup := createVariable(t, client, nil, nil, &Account{ID: defaultAccountID})
	defer vTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		options := VariableUpdateOptions{
			Key:         String("newname"),
			Value:       String("newvalue"),
			Description: String("newdescription"),
			HCL:         Bool(true),
		}

		v, err := client.Variables.Update(ctx, vTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Key, v.Key)
		assert.Equal(t, *options.HCL, v.HCL)
		assert.Equal(t, *options.Value, v.Value)
		assert.Equal(t, *options.Description, v.Description)
	})

	t.Run("when updating a subset of values", func(t *testing.T) {
		options := VariableUpdateOptions{
			Key: String("someothername"),
			HCL: Bool(false),
		}

		v, err := client.Variables.Update(ctx, vTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Key, v.Key)
		assert.Equal(t, *options.HCL, v.HCL)
	})

	t.Run("with sensitive set", func(t *testing.T) {
		options := VariableUpdateOptions{
			Sensitive: Bool(true),
		}

		v, err := client.Variables.Update(ctx, vTest.ID, options)
		require.NoError(t, err)

		assert.Equal(t, *options.Sensitive, v.Sensitive)
		assert.Empty(t, v.Value) // Because its now sensitive
	})

	t.Run("without any changes", func(t *testing.T) {
		created, vTestCleanup := createVariable(t, client, nil, nil, vTest.Account)
		defer vTestCleanup()

		updated, err := client.Variables.Update(ctx, created.ID, VariableUpdateOptions{})
		require.NoError(t, err)

		assert.Equal(t, created, updated)
	})

	t.Run("with invalid variable ID", func(t *testing.T) {
		_, err := client.Variables.Update(ctx, badIdentifier, VariableUpdateOptions{})
		assert.EqualError(t, err, "invalid value for variable ID")
	})
}

func TestVariablesDelete(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	wTest, wTestCleanup := createWorkspace(t, client, nil)
	defer wTestCleanup()

	t.Run("with valid options", func(t *testing.T) {
		vTest, _ := createVariable(t, client, wTest, nil, nil)
		err := client.Variables.Delete(ctx, vTest.ID)
		assert.NoError(t, err)
	})

	t.Run("with non existing variable ID", func(t *testing.T) {
		var variableId = "nonexisting"
		err := client.Variables.Delete(ctx, variableId)
		assert.Equal(
			t,
			ResourceNotFoundError{
				Message: fmt.Sprintf("Variable with ID '%s' not found or user unauthorized.", variableId),
			}.Error(),
			err.Error(),
		)
	})

	t.Run("with invalid variable ID", func(t *testing.T) {
		err := client.Variables.Delete(ctx, badIdentifier)
		assert.EqualError(t, err, "invalid value for variable ID")
	})
}

func TestVariablesList(t *testing.T) {
	client := testClient(t)
	ctx := context.Background()

	t.Run("scopes", func(t *testing.T) {
		variables, err := client.Variables.List(ctx, VariableListOptions{})
		if err != nil {
			log.Fatalf("Cant remove default variables before test: %v", err)
			return
		}
		for _, variable := range variables.Items {
			err = client.Variables.Delete(ctx, variable.ID)
			if err != nil {
				log.Fatalf("Cant remove default variables before test: %v", err)
				return
			}
		}

		accountVariable, deleteAccountVariable := createVariable(t, client, nil, nil, &Account{ID: defaultAccountID})
		defer deleteAccountVariable()

		requestedEnvironment, deleteRequestedEnvironment := createEnvironment(t, client)
		defer deleteRequestedEnvironment()

		environmentVariable, deleteEnvironmentVariable := createVariable(t, client, nil, requestedEnvironment, nil)
		defer deleteEnvironmentVariable()

		otherEnvironment, deleteOtherEnvironment := createEnvironment(t, client)
		defer deleteOtherEnvironment()

		_, deleteOtherEnvironmentVariable := createVariable(t, client, nil, otherEnvironment, nil)
		defer deleteOtherEnvironmentVariable()

		requestedWorkspace, deleteRequestedWorkspace := createWorkspace(t, client, requestedEnvironment)
		defer deleteRequestedWorkspace()

		workspaceVariable, deleteRequestedVariable := createVariable(t, client, requestedWorkspace, nil, nil)
		defer deleteRequestedVariable()

		otherWorkspace, deleteOtherWorkspace := createWorkspace(t, client, requestedEnvironment)
		defer deleteOtherWorkspace()

		_, deleteOtherWorkspaceVariable := createVariable(t, client, otherWorkspace, nil, nil)
		defer deleteOtherWorkspaceVariable()

		responseVariables, err := client.Variables.List(
			ctx, VariableListOptions{Filter: &VariableFilter{
				Workspace:   String("in:null," + requestedWorkspace.ID),
				Environment: String("in:null," + requestedEnvironment.ID),
				Account:     String("in:null," + defaultAccountID),
			}})
		require.NoError(t, err)

		expectedIds := []string{accountVariable.ID, environmentVariable.ID, workspaceVariable.ID}
		responseIds := make([]string, 0)
		for _, variable := range responseVariables.Items {
			responseIds = append(responseIds, variable.ID)
		}

		assert.ElementsMatch(t, expectedIds, responseIds)
	})

	t.Run("category", func(t *testing.T) {
		workspace, deleteWorkspace := createWorkspace(t, client, nil)
		defer deleteWorkspace()

		terraformVariable, err := client.Variables.Create(ctx, VariableCreateOptions{
			Key:       String(randomVariableKey(t)),
			Value:     String(randomString(t)),
			Category:  Category(CategoryTerraform),
			Workspace: workspace,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := client.Variables.Delete(ctx, terraformVariable.ID); err != nil {
				t.Errorf("Error destroying variable! WARNING: Dangling resources\n"+
					"may exist! The full error is shown below.\n\n"+
					"Variable: %s\nError: %s", terraformVariable.Key, err)
			}
		}()

		envVariable, err := client.Variables.Create(ctx, VariableCreateOptions{
			Key:       String(randomVariableKey(t)),
			Value:     String(randomString(t)),
			Category:  Category(CategoryEnv),
			Workspace: workspace,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := client.Variables.Delete(ctx, envVariable.ID); err != nil {
				t.Errorf("Error destroying variable! WARNING: Dangling resources\n"+
					"may exist! The full error is shown below.\n\n"+
					"Variable: %s\nError: %s", envVariable.Key, err)
			}
		}()
		responseVariables, err := client.Variables.List(
			ctx, VariableListOptions{Filter: &VariableFilter{
				Category: String(string(CategoryTerraform)),
			}},
		)
		if err != nil {
			t.Fatal(err)
		}
		assert.Len(t, responseVariables.Items, 1)
		assert.Equal(t, responseVariables.Items[0].ID, terraformVariable.ID)
	})

	t.Run("by id", func(t *testing.T) {
		fooVariable, deleteFooVariable := createVariable(t, client, nil, nil, &Account{ID: defaultAccountID})
		defer deleteFooVariable()

		_, deleteBarVariable := createVariable(t, client, nil, nil, &Account{ID: defaultAccountID})
		defer deleteBarVariable()

		responseVariables, err := client.Variables.List(
			ctx, VariableListOptions{Filter: &VariableFilter{
				Var: String(fooVariable.ID),
			}},
		)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, 1, responseVariables.TotalCount)
		assert.Equal(t, fooVariable.ID, responseVariables.Items[0].ID)
	})

	t.Run("name", func(t *testing.T) {
		workspace, deleteWorkspace := createWorkspace(t, client, nil)
		defer deleteWorkspace()

		fooVariable, err := client.Variables.Create(ctx, VariableCreateOptions{
			Key:       String("foo"),
			Value:     String(randomString(t)),
			Category:  Category(CategoryTerraform),
			Workspace: workspace,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := client.Variables.Delete(ctx, fooVariable.ID); err != nil {
				t.Errorf("Error destroying variable! WARNING: Dangling resources\n"+
					"may exist! The full error is shown below.\n\n"+
					"Variable: %s\nError: %s", fooVariable.Key, err)
			}
		}()

		barVariable, err := client.Variables.Create(ctx, VariableCreateOptions{
			Key:       String("bar"),
			Value:     String(randomString(t)),
			Category:  Category(CategoryTerraform),
			Workspace: workspace,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := client.Variables.Delete(ctx, barVariable.ID); err != nil {
				t.Errorf("Error destroying variable! WARNING: Dangling resources\n"+
					"may exist! The full error is shown below.\n\n"+
					"Variable: %s\nError: %s", barVariable.Key, err)
			}
		}()

		bazVariable, err := client.Variables.Create(ctx, VariableCreateOptions{
			Key:       String("baz"),
			Value:     String(randomString(t)),
			Category:  Category(CategoryEnv),
			Workspace: workspace,
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := client.Variables.Delete(ctx, bazVariable.ID); err != nil {
				t.Errorf("Error destroying variable! WARNING: Dangling resources\n"+
					"may exist! The full error is shown below.\n\n"+
					"Variable: %s\nError: %s", bazVariable.Key, err)
			}
		}()
		responseVariables, err := client.Variables.List(
			ctx, VariableListOptions{Filter: &VariableFilter{
				Key: String("in:bar,baz"),
			}},
		)
		if err != nil {
			t.Fatal(err)
		}

		expectedIds := []string{barVariable.ID, bazVariable.ID}
		responseIds := make([]string, 0)
		for _, variable := range responseVariables.Items {
			responseIds = append(responseIds, variable.ID)
		}

		assert.ElementsMatch(t, expectedIds, responseIds)
	})
}
