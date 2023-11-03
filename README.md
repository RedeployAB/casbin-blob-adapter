# blobadapter

> Azure Blob Storage adapter for casbin.

[Casbin](https://github.com/casbin/casbin) adapter implementation for Azure Blob Storage.

* [Installation](#installation)
* [Example usage](#example-usage)
* [Constructor functions](#constructor-functions)


## Installation

```sh
go get github.com/RedeployAB/casbin-blob-adapter 
```

## Example usage

This example uses `azcore.TokenCredential` as credentials for the
adapter. See [Constructor functions](#constructor-functions) below
for other options.

```go
package main

import (
    "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
    blobadapter "github.com/RedeployAB/casbin-blob-adapter"
    "github.com/casbin/casbin/v2"
)

func main() {
    // Create credentials for Azure Blob Storage (service principal, managed identity, az cli).
    cred, err := azidentity.NewDefaultAzureCredential(nil)
    if err != nil {
        // Handle error.
    }

    // Create the adapter for Azure Blob Storage. Provide account (storage account name),
    // container name, blob name and credentials. If the container and blob does not exist,
    // they will be created.
    a, err := blobadapter.NewAdapter("account", "container", "policy.csv", cred)
    if err != nil {
        // Handle error.
    }

    e, err := casbin.NewEnforcer("rbac_with_domains_model.conf", a)
    if err != nil {
        // Handle error.
    }

    // Load the policy from the specified blob in Azure Blob Storage manually.
    // NOTE: Like all implicit and explicit adapters the policies is loaded
    // automatically when calling NewEnforcer. This method can be used at
    // runtime to reload policy.
    if err := e.LoadPolicy(); err != nil {
        // Handle error.
    }

    // Check the permission.
    ok, err := e.Enforce("alice", "domain1", "data1", "read")
    if err != nil {
        // Handle error.
    }

    // Modify policy.
    // e.AddPolicy(...)
    // e.RemovePolicy(...)

    // Save policy back to the blob in Azure Blob Storage.
    if err := e.SavePolicy(); err != nil {
        // Handle error.
    }
}
```

## Constructor functions

**`NewAdapter(account string, container string, blob string, cred azcore.TokenCredential, options ...Option) (*Adapter, error)`**

Uses `azcore.TokenCredential`. See [`azidentity`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity) for
more options on creating credentials.

```go
cred, err := azidentity.NewDefaultAzureCredential(nil)
if err != nil {
    // Handle error.
}

a, err := blobadapter.NewAdapter("account", "container", "policy.csv", cred)
if err != nil {
    // Handle error.
}
```

**`NewAdapterFromConnectionString(connectionString string, container string, blob string, options ...Option) (*Adapter, error)`**

Uses a connection string for an Azure Storage account.

```go
a, err := blobadapter.NewAdapterFromConnectionString("connectionstring", "container", "policy.csv")
if err != nil {
    // Handle error.
}
```

**`NewAdapterFromSharedKeyCredential(account string, key string, container string, blob string, options ...Option) (*Adapter, error)`**

Uses storage account name and key for an Azure Storage account.

```go
a, err := blobadapter.NewAdapterFromSharedKeyCredential("account", "key", "container", "policy.csv")
if err != nil {
    // Handle error.
}
```