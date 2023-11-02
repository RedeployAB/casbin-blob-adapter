package blobadapter

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/casbin/casbin/v2/model"
	"github.com/casbin/casbin/v2/persist"
	"github.com/casbin/casbin/v2/util"
)

// client is the interface that wraps around methods CreateContainer, DownloadStream and UploadStream.
type client interface {
	CreateContainer(ctx context.Context, containerName string, o *azblob.CreateContainerOptions) (azblob.CreateContainerResponse, error)
	DownloadStream(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error)
	UploadStream(ctx context.Context, containerName string, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error)
}

// Adapter is an Azure Blob Storage adapter for casbin.
type Adapter struct {
	c         client
	container string
	blob      string
	timeout   time.Duration
	initiated bool
}

// NewAdapter returns a new adapter with the given account, container, blob and credentials.
func NewAdapter(account, container, blob string, cred azcore.TokenCredential, options ...Option) (*Adapter, error) {
	if err := checkAccountCredentialsArguments(account, cred); err != nil {
		return nil, err
	}

	a, err := newAdapter(container, blob, options...)
	if err != nil {
		return nil, err
	}

	if a.c == nil {
		a.c, err = azblob.NewClient(serviceURL(account), cred, nil)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

// NewAdapterFromConnectionString returns a new adapter with the given connection string, container and blob.
func NewAdapterFromConnectionString(connectionString, container, blob string, options ...Option) (*Adapter, error) {
	if len(connectionString) == 0 {
		return nil, ErrInvalidConnectionString
	}

	a, err := newAdapter(container, blob, options...)
	if err != nil {
		return nil, err
	}

	if a.c == nil {
		a.c, err = azblob.NewClientFromConnectionString(connectionString, nil)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

// NewAdapterFromSharedKeyCredential returns a new adapter with the given account, key, container and blob.
func NewAdapterFromSharedKeyCredential(account, key, container, blob string, options ...Option) (*Adapter, error) {
	if err := checkAccountKeyArguments(account, key); err != nil {
		return nil, err
	}

	a, err := newAdapter(container, blob, options...)
	if err != nil {
		return nil, err
	}

	if a.c == nil {
		cred, err := azblob.NewSharedKeyCredential(account, key)
		if err != nil {
			return nil, err
		}
		a.c, err = azblob.NewClientWithSharedKeyCredential(serviceURL(account), cred, nil)
		if err != nil {
			return nil, err
		}
	}
	return a, nil
}

// newAdapter returns a new adapter with the given container, blob and options.
func newAdapter(container, blob string, options ...Option) (*Adapter, error) {
	if err := checkContainerBlobArguments(container, blob); err != nil {
		return nil, err
	}

	a := &Adapter{
		container: container,
		blob:      blob,
		timeout:   time.Second * 10,
	}

	for _, option := range options {
		option(a)
	}
	return a, nil
}

// serviceURL returns the service URL for the provided account.
func serviceURL(account string) string {
	return strings.Replace("https://{account}.blob.core.windows.net/", "{account}", account, 1)
}

// LoadPolicy loads all policy rules from the storage.
func (a *Adapter) LoadPolicy(model model.Model) error {
	if err := checkContainerBlobArguments(a.container, a.blob); err != nil {
		return err
	}
	return a.loadPolicyBlob(model, persist.LoadPolicyLine)
}

// loadPolicyBlob loads all policy rules from the storage by downloading
// the blob and reading it line by line.
func (a *Adapter) loadPolicyBlob(model model.Model, handler func(string, model.Model) error) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	res, err := a.c.DownloadStream(ctx, a.container, a.blob, nil)
	if err != nil {
		// When creating the enforcer with the adapter it will attempt to load the policy,
		// to avoid an error at this point, return nil and set initiated to true.
		// Next call to LoadPolicy will return the error.
		if bloberror.HasCode(err, bloberror.ContainerNotFound) {
			if !a.initiated {
				a.initiated = true
				return nil
			}
			return fmt.Errorf("%w: %s", ErrContainerDoesNotExist, a.container)
		} else if bloberror.HasCode(err, bloberror.BlobNotFound) {
			if !a.initiated {
				a.initiated = true
				return nil
			}
			return fmt.Errorf("%w: %s", ErrBlobDoesNotExist, a.blob)
		} else {
			return err
		}
	}

	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if err := handler(line, model); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// SavePolicy saves all policy rules to the storage.
func (a *Adapter) SavePolicy(model model.Model) error {
	if err := checkContainerBlobArguments(a.container, a.blob); err != nil {
		return err
	}

	var buf bytes.Buffer
	for ptype, ast := range model["p"] {
		for _, rule := range ast.Policy {
			writeRule(&buf, ptype, rule)
		}
	}

	for ptype, ast := range model["g"] {
		for _, rule := range ast.Policy {
			writeRule(&buf, ptype, rule)
		}
	}

	return a.savePolicyBlob(strings.TrimRight(buf.String(), "\n"))
}

// savePolicyBlob saves all policy rules to the storage by uploading
// the blob.
func (a *Adapter) savePolicyBlob(text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), a.timeout)
	defer cancel()

	if _, err := a.c.CreateContainer(ctx, a.container, nil); err != nil && !bloberror.HasCode(err, bloberror.ContainerAlreadyExists, bloberror.ResourceAlreadyExists) {
		return err
	}
	_, err := a.c.UploadStream(ctx, a.container, a.blob, bytes.NewReader([]byte(text)), nil)
	return err
}

// AddPolicy adds a policy rule to the storage.
// NOTE: This method is not implemented.
func (a *Adapter) AddPolicy(sec, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemovePolicy removes a policy rule from the storage.
// NOTE: This method is not implemented.
func (a *Adapter) RemovePolicy(sec, ptype string, rule []string) error {
	return errors.New("not implemented")
}

// RemoveFilteredPolicy removes policy rules that match the filter from the storage.
// NOTE: This method is not implemented.
func (a *Adapter) RemoveFilteredPolicy(sec, ptype string, fieldIndex int, fieldValues ...string) error {
	return errors.New("not implemented")
}

// writeRule writes ptype and rule to the buffer.
func writeRule(buf *bytes.Buffer, ptype string, rule []string) {
	buf.WriteString(ptype + ", ")
	buf.WriteString(util.ArrayToString(rule))
	buf.WriteString("\n")
}

// checkAccountCredentialsArguments checks if the provided account and credentials are not empty.
func checkAccountCredentialsArguments(account string, cred azcore.TokenCredential) error {
	if len(account) == 0 {
		return ErrInvalidAccount
	}
	if cred == nil {
		return ErrInvalidCredential
	}
	return nil
}

// checkContainerBlobArguments checks if the provided container and blob are not empty.
func checkContainerBlobArguments(container, blob string) error {
	if len(container) == 0 {
		return ErrInvalidContainer
	}
	if len(blob) == 0 {
		return ErrInvalidBlob
	}
	return nil
}

// checkAccountKeyArguments checks if the provided account and key are not empty.
func checkAccountKeyArguments(account, key string) error {
	if len(account) == 0 {
		return ErrInvalidAccount
	}
	if len(key) == 0 {
		return ErrInvalidKey
	}
	return nil
}
