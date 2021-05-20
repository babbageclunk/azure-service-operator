// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package nic

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	vnetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-09-01/network"
	"github.com/Azure/azure-service-operator/pkg/helpers"
	"github.com/Azure/azure-service-operator/pkg/resourcemanager/config"
	"github.com/Azure/azure-service-operator/pkg/resourcemanager/iam"
	"github.com/Azure/azure-service-operator/pkg/secrets"
)

type AzureNetworkInterfaceClient struct {
	Creds        config.Credentials
	SecretClient secrets.SecretClient
	Scheme       *runtime.Scheme
}

func NewAzureNetworkInterfaceClient(creds config.Credentials, secretclient secrets.SecretClient, scheme *runtime.Scheme) *AzureNetworkInterfaceClient {
	return &AzureNetworkInterfaceClient{
		Creds:        creds,
		SecretClient: secretclient,
		Scheme:       scheme,
	}
}

func getNetworkInterfaceClient(creds config.Credentials) (vnetwork.InterfacesClient, error) {
	nicClient := vnetwork.NewInterfacesClientWithBaseURI(config.BaseURI(), creds.SubscriptionID())
	a, err := iam.GetResourceManagementAuthorizer(creds)
	if err != nil {
		return vnetwork.InterfacesClient{}, errors.Wrapf(err, "getting authorizer")
	}
	nicClient.Authorizer = a
	nicClient.AddToUserAgent(config.UserAgent())
	return nicClient, nil
}

func (m *AzureNetworkInterfaceClient) CreateNetworkInterface(ctx context.Context, location string, resourceGroupName string, resourceName string, vnetName string, subnetName string, publicIPAddressName string) (vnetwork.InterfacesCreateOrUpdateFuture, error) {

	client, err := getNetworkInterfaceClient(m.Creds)
	if err != nil {
		return vnetwork.InterfacesCreateOrUpdateFuture{}, err
	}

	subnetIDInput := helpers.MakeResourceID(
		client.SubscriptionID,
		resourceGroupName,
		"Microsoft.Network",
		"virtualNetworks",
		vnetName,
		"subnets",
		subnetName,
	)

	publicIPAddressIDInput := helpers.MakeResourceID(
		client.SubscriptionID,
		resourceGroupName,
		"Microsoft.Network",
		"publicIPAddresses",
		publicIPAddressName,
		"",
		"",
	)

	var ipConfigsToAdd []vnetwork.InterfaceIPConfiguration
	ipConfigsToAdd = append(
		ipConfigsToAdd,
		vnetwork.InterfaceIPConfiguration{
			Name: &resourceName,
			InterfaceIPConfigurationPropertiesFormat: &vnetwork.InterfaceIPConfigurationPropertiesFormat{
				Subnet: &vnetwork.Subnet{
					ID: &subnetIDInput,
				},
				PublicIPAddress: &vnetwork.PublicIPAddress{
					ID: &publicIPAddressIDInput,
				},
			},
		},
	)

	future, err := client.CreateOrUpdate(
		ctx,
		resourceGroupName,
		resourceName,
		vnetwork.Interface{
			Location: &location,
			InterfacePropertiesFormat: &vnetwork.InterfacePropertiesFormat{
				IPConfigurations: &ipConfigsToAdd,
			},
		},
	)

	return future, err
}

func (m *AzureNetworkInterfaceClient) DeleteNetworkInterface(ctx context.Context, nicName string, resourcegroup string) (string, error) {

	client, err := getNetworkInterfaceClient(m.Creds)
	if err != nil {
		return "", err
	}

	_, err = client.Get(ctx, resourcegroup, nicName, "")
	if err == nil { // nic present, so go ahead and delete
		future, err := client.Delete(ctx, resourcegroup, nicName)
		return future.Status(), err
	}
	// nic not present so return success anyway
	return "nic not present", nil

}

func (m *AzureNetworkInterfaceClient) GetNetworkInterface(ctx context.Context, resourcegroup string, nicName string) (vnetwork.Interface, error) {

	client, err := getNetworkInterfaceClient(m.Creds)
	if err != nil {
		return vnetwork.Interface{}, err
	}

	return client.Get(ctx, resourcegroup, nicName, "")
}
