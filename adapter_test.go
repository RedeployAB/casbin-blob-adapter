package blobadapter

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/casbin/casbin/v2"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestNewAdapter(t *testing.T) {
	var tests = []struct {
		name  string
		input struct {
			account   string
			container string
			blob      string
			cred      azcore.TokenCredential
			options   []Option
		}
		want    *Adapter
		wantErr error
	}{
		{
			name: "Create a new adapter",
			input: struct {
				account   string
				container string
				blob      string
				cred      azcore.TokenCredential
				options   []Option
			}{
				account:   "account",
				container: "container",
				blob:      "blob",
				cred:      &mockCredential{},
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want: &Adapter{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 10,
			},
			wantErr: nil,
		},
		{
			name: "Create a new adapter with options",
			input: struct {
				account   string
				container string
				blob      string
				cred      azcore.TokenCredential
				options   []Option
			}{
				account:   "account",
				container: "container",
				blob:      "blob",
				cred:      &mockCredential{},
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
					WithTimeout(time.Second * 20),
				},
			},
			want: &Adapter{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 20,
			},
		},
		{
			name: "Create a new adapter with a container and blob that already exist",
			input: struct {
				account   string
				container string
				blob      string
				cred      azcore.TokenCredential
				options   []Option
			}{
				account:   "account",
				container: "container",
				blob:      "blob",
				cred:      &mockCredential{},
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{
							containerFound: true,
							blobFound:      true,
						}
					},
				},
			},
			want: &Adapter{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 10,
			},
		},
		{
			name: "Create a new adapter with invalid account",
			input: struct {
				account   string
				container string
				blob      string
				cred      azcore.TokenCredential
				options   []Option
			}{
				account:   "",
				container: "container",
				blob:      "blob",
				cred:      &mockCredential{},
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want:    nil,
			wantErr: ErrInvalidAccount,
		},
		{
			name: "Create a new adapter with invalid credentials",
			input: struct {
				account   string
				container string
				blob      string
				cred      azcore.TokenCredential
				options   []Option
			}{
				account:   "account",
				container: "container",
				blob:      "blob",
				cred:      nil,
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want:    nil,
			wantErr: ErrInvalidCredential,
		},
		{
			name: "Create a new adapter with invalid container",
			input: struct {
				account   string
				container string
				blob      string
				cred      azcore.TokenCredential
				options   []Option
			}{
				account:   "account",
				container: "",
				blob:      "blob",
				cred:      &mockCredential{},
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want:    nil,
			wantErr: ErrInvalidContainer,
		},
		{
			name: "Create a new adapter with invalid blob",
			input: struct {
				account   string
				container string
				blob      string
				cred      azcore.TokenCredential
				options   []Option
			}{
				account:   "account",
				container: "container",
				blob:      "",
				cred:      &mockCredential{},
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want:    nil,
			wantErr: ErrInvalidBlob,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := NewAdapter(test.input.account, test.input.container, test.input.blob, test.input.cred, test.input.options...)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Adapter{}), cmpopts.IgnoreUnexported(mockBlobClient{})); diff != "" {
				t.Errorf("NewAdapter() unexpected result (-want +got):\n%s\n", diff)
			}

			if diff := cmp.Diff(test.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("NewAdapter() unexpected error (-want +got):\n%s\n", diff)
			}
		})
	}
}

func TestNewAdapterFromConnectionString(t *testing.T) {
	var tests = []struct {
		name  string
		input struct {
			connectionString string
			container        string
			blob             string
			options          []Option
		}
		want    *Adapter
		wantErr error
	}{
		{
			name: "Create a new adapter",
			input: struct {
				connectionString string
				container        string
				blob             string
				options          []Option
			}{
				connectionString: fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=<accountName>;AccountKey=%s;EndpointSuffix=core.windows.net", _testKey),
				container:        "container",
				blob:             "blob",
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want: &Adapter{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 10,
			},
			wantErr: nil,
		},
		{
			name: "Create a new adapter with options",
			input: struct {
				connectionString string
				container        string
				blob             string
				options          []Option
			}{
				connectionString: fmt.Sprintf("DefaultEndpointsProtocol=https;AccountName=<accountName>;AccountKey=%s;EndpointSuffix=core.windows.net", _testKey),
				container:        "container",
				blob:             "blob",
				options: []Option{
					WithTimeout(time.Second * 20),
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want: &Adapter{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 20,
			},
		},
		{
			name: "Create a new adapter with invalid connection string",
			input: struct {
				connectionString string
				container        string
				blob             string
				options          []Option
			}{
				connectionString: "",
				container:        "container",
				blob:             "blob",
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want:    nil,
			wantErr: ErrInvalidConnectionString,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := NewAdapterFromConnectionString(test.input.connectionString, test.input.container, test.input.blob, test.input.options...)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Adapter{}), cmpopts.IgnoreUnexported(mockBlobClient{})); diff != "" {
				t.Errorf("NewAdapterFromConnectionString() unexpected result (-want +got):\n%s\n", diff)
			}

			if diff := cmp.Diff(test.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("NewAdapterFromConnectionString() unexpected error (-want +got):\n%s\n", diff)
			}
		})
	}
}

func TestNewAdapterFromSharedKeyCredential(t *testing.T) {
	var tests = []struct {
		name  string
		input struct {
			account   string
			key       string
			container string
			blob      string
			options   []Option
		}
		want    *Adapter
		wantErr error
	}{
		{
			name: "Create a new adapter",
			input: struct {
				account   string
				key       string
				container string
				blob      string
				options   []Option
			}{
				account:   "account",
				key:       _testKey,
				container: "container",
				blob:      "blob",
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want: &Adapter{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 10,
			},
		},
		{
			name: "Create a new adapter with options",
			input: struct {
				account   string
				key       string
				container string
				blob      string
				options   []Option
			}{
				account:   "account",
				key:       _testKey,
				container: "container",
				blob:      "blob",
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
					WithTimeout(time.Second * 20),
				},
			},
			want: &Adapter{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 20,
			},
		},
		{
			name: "Create a new adapter with invalid account",
			input: struct {
				account   string
				key       string
				container string
				blob      string
				options   []Option
			}{
				account:   "",
				key:       _testKey,
				container: "container",
				blob:      "blob",
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want:    nil,
			wantErr: ErrInvalidAccount,
		},
		{
			name: "Create a new adapter with invalid key",
			input: struct {
				account   string
				key       string
				container string
				blob      string
				options   []Option
			}{
				account:   "account",
				key:       "",
				container: "container",
				blob:      "blob",
				options: []Option{
					func(a *Adapter) {
						a.c = &mockBlobClient{}
					},
				},
			},
			want:    nil,
			wantErr: ErrInvalidKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := NewAdapterFromSharedKeyCredential(test.input.account, test.input.key, test.input.container, test.input.blob, test.input.options...)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Adapter{}), cmpopts.IgnoreUnexported(mockBlobClient{})); diff != "" {
				t.Errorf("NewAdapterFromSharedKeyCredential() unexpected result (-want +got):\n%s\n", diff)
			}

			if diff := cmp.Diff(test.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("NewAdapterFromSharedKeyCredential() unexpected error (-want +got):\n%s\n", diff)
			}
		})
	}
}

func TestAdapter_LoadPolicy(t *testing.T) {
	var tests = []struct {
		name    string
		input   func() *Adapter
		want    [][]string
		wantErr error
	}{
		{
			name: "Load policy",
			input: func() *Adapter {
				return &Adapter{
					c:         &mockBlobClient{},
					container: "container",
					blob:      "blob",
				}
			},
			want: [][]string{
				{"alice", "domain1", "data1", "read"},
			},
		},
		{
			name: "Load policy with error (container does not exist)",
			input: func() *Adapter {
				return &Adapter{
					c: &mockBlobClient{
						errDownload: &azcore.ResponseError{
							ErrorCode: string(bloberror.ContainerNotFound),
						},
					},
					container: "container",
					blob:      "blob",
				}
			},
			want:    nil,
			wantErr: ErrContainerDoesNotExist,
		},
		{
			name: "Load policy with error (blob does not exist)",
			input: func() *Adapter {
				return &Adapter{
					c: &mockBlobClient{
						errDownload: &azcore.ResponseError{
							ErrorCode: string(bloberror.BlobNotFound),
						},
					},
					container: "container",
					blob:      "blob",
				}
			},
			want:    nil,
			wantErr: ErrBlobDoesNotExist,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := test.input()
			e, gotErr := casbin.NewEnforcer("_examples/rbac_with_domains_model.conf", a)
			if gotErr == nil {
				got := e.GetPolicy()
				if diff := cmp.Diff(test.want, got); diff != "" {
					t.Errorf("LoadPolicy() unexpected result (-want +got):\n%s\n", diff)
				}
			}

			if diff := cmp.Diff(test.wantErr, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("LoadPolicy() unexpected error (-want +got):\n%s\n", diff)
			}
		})
	}
}

func TestAdapter_SavePolicy(t *testing.T) {
	var tests = []struct {
		name  string
		input struct {
			c         *mockBlobClient
			container string
			blob      string
		}
		want    []byte
		wantErr error
	}{
		{
			name: "Save policy",
			input: struct {
				c         *mockBlobClient
				container string
				blob      string
			}{
				c:         &mockBlobClient{},
				container: "container",
				blob:      "blob",
			},
			want: []byte(`p, alice, domain1, data1, read` + "\n" + `g, alice, admin, domain1`),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			a := &Adapter{
				c:         test.input.c,
				container: test.input.container,
				blob:      test.input.blob,
			}

			e, err := casbin.NewEnforcer("_examples/rbac_with_domains_model.conf", a)
			if err != nil {
				t.Errorf("error in test: %v\n", err)
			}

			_, _ = e.AddPolicy("alice", "domain1", "data1", "read")
			_, _ = e.AddGroupingPolicy("alice", "admin", "domain1")

			gotErr := e.SavePolicy()
			got := test.input.c.policies

			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("SavePolicy() unexpected result (-want +got):\n%s\n", diff)
			}

			if diff := cmp.Diff(nil, gotErr, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("SavePolicy() unexpected error (-want +got):\n%s\n", diff)
			}

		})
	}
}

type mockBlobClient struct {
	errCreate      error
	errDownload    error
	errUpload      error
	containerFound bool
	blobFound      bool
	policies       []byte
}

func (c mockBlobClient) NewListContainersPager(o *azblob.ListContainersOptions) *runtime.Pager[azblob.ListContainersResponse] {
	containers := []*service.ContainerItem{}
	if c.containerFound {
		containers = append(containers, &service.ContainerItem{
			Name: toPtr("container"),
		})
	}
	pager := runtime.NewPager(runtime.PagingHandler[azblob.ListContainersResponse]{
		More: func(page azblob.ListContainersResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, page *azblob.ListContainersResponse) (azblob.ListContainersResponse, error) {
			return azblob.ListContainersResponse{
				ListContainersSegmentResponse: azblob.ListContainersSegmentResponse{
					ContainerItems: containers,
				},
			}, nil
		},
	})
	return pager
}

func (c mockBlobClient) NewListBlobsFlatPager(containerName string, o *azblob.ListBlobsFlatOptions) *runtime.Pager[azblob.ListBlobsFlatResponse] {
	blobs := []*container.BlobItem{}
	if c.blobFound {
		blobs = append(blobs, &container.BlobItem{
			Name: toPtr("blob"),
		})
	}
	pager := runtime.NewPager(runtime.PagingHandler[azblob.ListBlobsFlatResponse]{
		More: func(page azblob.ListBlobsFlatResponse) bool {
			return false
		},
		Fetcher: func(ctx context.Context, page *azblob.ListBlobsFlatResponse) (azblob.ListBlobsFlatResponse, error) {
			return azblob.ListBlobsFlatResponse{
				ListBlobsFlatSegmentResponse: azblob.ListBlobsFlatSegmentResponse{
					Segment: &container.BlobFlatListSegment{
						BlobItems: blobs,
					},
				},
			}, nil
		},
	})
	return pager
}

func (c mockBlobClient) CreateContainer(ctx context.Context, containerName string, o *azblob.CreateContainerOptions) (azblob.CreateContainerResponse, error) {
	if c.errCreate != nil {
		return azblob.CreateContainerResponse{}, c.errCreate
	}
	return azblob.CreateContainerResponse{}, nil
}

func (c mockBlobClient) DownloadStream(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error) {
	if c.errDownload != nil {
		return azblob.DownloadStreamResponse{}, c.errDownload
	}
	return azblob.DownloadStreamResponse{
		DownloadResponse: blob.DownloadResponse{
			Body: io.NopCloser(bytes.NewReader([]byte(`p, alice, domain1, data1, read`))),
		},
	}, nil
}

func (c *mockBlobClient) UploadStream(ctx context.Context, containerName string, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error) {
	if c.errUpload != nil {
		return azblob.UploadStreamResponse{}, c.errUpload
	}
	b, _ := io.ReadAll(body)
	c.policies = b
	return azblob.UploadStreamResponse{}, nil
}

type mockCredential struct{}

func (c *mockCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{}, nil
}

var _testKey = base64.StdEncoding.EncodeToString([]byte("<accountKey>"))
