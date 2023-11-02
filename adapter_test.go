package blobadapter

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
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

type mockBlobClient struct{}

func (c mockBlobClient) CreateContainer(ctx context.Context, containerName string, o *azblob.CreateContainerOptions) (azblob.CreateContainerResponse, error) {
	return azblob.CreateContainerResponse{}, nil
}

func (c mockBlobClient) DownloadStream(ctx context.Context, containerName string, blobName string, o *azblob.DownloadStreamOptions) (azblob.DownloadStreamResponse, error) {
	return azblob.DownloadStreamResponse{}, nil
}

func (c mockBlobClient) UploadStream(ctx context.Context, containerName string, blobName string, body io.Reader, o *azblob.UploadStreamOptions) (azblob.UploadStreamResponse, error) {
	return azblob.UploadStreamResponse{}, nil
}

type mockCredential struct{}

func (c *mockCredential) GetToken(ctx context.Context, options policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return azcore.AccessToken{}, nil
}
