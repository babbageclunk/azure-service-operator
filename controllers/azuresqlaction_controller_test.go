// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

// +build all azuresqlserver azuresqlservercombined testaction

package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	azurev1alpha1 "github.com/Azure/azure-service-operator/api/v1alpha1"
	"github.com/Azure/azure-service-operator/pkg/secrets"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RunSQLActionHappy(t *testing.T, server string) {
	defer PanicRecover(t)
	ctx := context.Background()
	assert := assert.New(t)

	// Add any setup steps that needs to be executed before each test
	rgName := tc.resourceGroupName

	//Get SQL credentials to compare after rollover
	var secret map[string][]byte
	assert.Eventually(func() bool {
		secretName := getSecretName(server)
		var err error
		secret, err = tc.secretClient.Get(ctx, secrets.SecretKey{Name: secretName, Namespace: "default"})
		if err != nil {
			return false
		}
		return true
	}, tc.timeoutFast, tc.retry, "wait for server to return secret")

	sqlActionName := GenerateTestResourceNameWithRandom("azuresqlaction-dev", 10)

	// Create the Sql Action object and expect the Reconcile to be created
	sqlActionInstance := &azurev1alpha1.AzureSqlAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      sqlActionName,
			Namespace: "default",
		},
		Spec: azurev1alpha1.AzureSqlActionSpec{
			ActionName:    "rolladmincreds",
			ServerName:    server,
			ResourceGroup: rgName,
		},
	}

	EnsureInstance(ctx, t, tc, sqlActionInstance)

	// makre sure credentials are not the same as previous
	var secretAfter map[string][]byte
	assert.Eventually(func() bool {
		var secretName string
		if tc.secretClient.GetSecretNamingVersion() == secrets.SecretNamingV1 {
			secretName = server
		} else {
			secretName = "azuresqlserver-" + server
		}
		var err error
		secretAfter, err = tc.secretClient.Get(ctx, secrets.SecretKey{Name: secretName, Namespace: "default"})
		if err != nil {
			return false
		}
		return true
	}, tc.timeoutFast, tc.retry, "wait for server to return secret")

	assert.Equal(secret["username"], secretAfter["username"], "username should still be the same")
	assert.NotEqual(string(secret["password"]), string(secretAfter["password"]), "password should have changed")

	EnsureDelete(ctx, t, tc, sqlActionInstance)
}

func getSecretName(server string) string {
	var secretName string
	if tc.secretClient.GetSecretNamingVersion() == secrets.SecretNamingV1 {
		secretName = server
	} else {
		secretName = "azuresqlserver-" + server
	}
	return secretName
}
