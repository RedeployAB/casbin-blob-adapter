# blobadapter

> Azure Blob Storage adapter for casbin.

[Casbin](https://github.com/casbin/casbin) adapter implementation for Azure Blob Storage.

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
    "github.com/casbin/casbin"
    "github.com/RedeployAB/casbin-blob-adapter"
)

func main() {
    // Create credentials for Azure Blob Storage (service principal, managed identity, az cli).
    cred, err := azidentity.NewDefaultAzureCredential(nil)
    if err != nil {
        // Handle error.
    }

    // Create the adapter for Azure Blob Storage. Provide account (storage account name),
    // container name, blob name and credentials. If the container does not exist,
    // it will be created when loading/saving the policy.
    a, err := blobadapter.NewAdapter("account", "container", "policy.csv", cred)
    if err != nil {
        // Handle error.
    }

    e, err := casbin.NewEnforcer("rbac_model.conf", a)
    if err != nil {
        // Handle error.
    }

    // Load the policy from the specified blob in Azure Blob Storage.
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

### Constructor functions

**`NewAdapter(account string, container string, blob string, cred azcore.TokenCredential) (*Adapter, error)`**

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

**`NewAdapterFromConnectionString(connectionString string, container string, blob string) (*Adapter, error)`**

Uses a connection string for an Azure Storage account.

```go
a, err := blobadapter.NewAdapterFromConnectionString("connectionstring", "container", "policy.csv")
if err != nil {
    // Handle error.
}
```

**`NewAdapterFromSharedKeyCredential(account string, key string, container string, blob string) (*Adapter, error)`**

Uses storage account name and key for an Azure Storage account.

```go
a, err := blobadapter.NewAdapterFromSharedKeyCredential("account", "key", "container", "policy.csv")
if err != nil {
    // Handle error.
}
```