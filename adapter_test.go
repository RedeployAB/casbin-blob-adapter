package blobadapter

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
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
			},
			want: &Adapter{
				c:         &azblob.Client{},
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
					WithTimeout(time.Second * 20),
				},
			},
			want: &Adapter{
				c:         &azblob.Client{},
				container: "container",
				blob:      "blob",
				timeout:   time.Second * 20,
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
			},
			want:    nil,
			wantErr: ErrInvalidBlob,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := NewAdapter(test.input.account, test.input.container, test.input.blob, test.input.cred, test.input.options...)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Adapter{}), cmpopts.IgnoreUnexported(azblob.Client{})); diff != "" {
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
				connectionString: "DefaultEndpointsProtocol=https;AccountName=<accountName>;AccountKey=PGFjY291bnRLZXk+;EndpointSuffix=core.windows.net",
				container:        "container",
				blob:             "blob",
			},
			want: &Adapter{
				c:         &azblob.Client{},
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
				connectionString: "DefaultEndpointsProtocol=https;AccountName=<accountName>;AccountKey=PGFjY291bnRLZXk+;EndpointSuffix=core.windows.net",
				container:        "container",
				blob:             "blob",
				options: []Option{
					WithTimeout(time.Second * 20),
				},
			},
			want: &Adapter{
				c:         &azblob.Client{},
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
			},
			want:    nil,
			wantErr: ErrInvalidConnectionString,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := NewAdapterFromConnectionString(test.input.connectionString, test.input.container, test.input.blob, test.input.options...)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Adapter{}), cmpopts.IgnoreUnexported(azblob.Client{})); diff != "" {
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
				key:       "PGFjY291bnRLZXk+",
				container: "container",
				blob:      "blob",
			},
			want: &Adapter{
				c:         &azblob.Client{},
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
				key:       "PGFjY291bnRLZXk+",
				container: "container",
				blob:      "blob",
				options: []Option{
					WithTimeout(time.Second * 20),
				},
			},
			want: &Adapter{
				c:         &azblob.Client{},
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
				key:       "PGFjY291bnRLZXk+",
				container: "container",
				blob:      "blob",
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
			},
			want:    nil,
			wantErr: ErrInvalidKey,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, gotErr := NewAdapterFromSharedKeyCredential(test.input.account, test.input.key, test.input.container, test.input.blob, test.input.options...)

			if diff := cmp.Diff(test.want, got, cmp.AllowUnexported(Adapter{}), cmpopts.IgnoreUnexported(azblob.Client{})); diff != "" {
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
			name: "Load policy with error (container does not exist) first call",
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
			want: nil,
		},
		{
			name: "Load policy with error (container does not exist) second call",
			input: func() *Adapter {
				return &Adapter{
					c: &mockBlobClient{
						errDownload: &azcore.ResponseError{
							ErrorCode: string(bloberror.ContainerNotFound),
						},
					},
					container: "container",
					blob:      "blob",
					initiated: true,
				}
			},
			want:    nil,
			wantErr: ErrContainerDoesNotExist,
		},
		{
			name: "Load policy with error (blob does not exist) first call",
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
			want: nil,
		},
		{
			name: "Load policy with error (blob does not exist) second call",
			input: func() *Adapter {
				return &Adapter{
					c: &mockBlobClient{
						errDownload: &azcore.ResponseError{
							ErrorCode: string(bloberror.BlobNotFound),
						},
					},
					container: "container",
					blob:      "blob",
					initiated: true,
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
			/* if err != nil {
				t.Errorf("error in test: %v\n", err)
			} */

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

func TestWriteRule(t *testing.T) {
	var tests = []struct {
		name string
	}{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}
}

type mockBlobClient struct {
	errCreate   error
	errDownload error
	errUpload   error
	policies    []byte
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
