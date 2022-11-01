package helpers

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"os"
	"time"
)

func NewAzureCredential(options *azidentity.DefaultAzureCredentialOptions) (azcore.TokenCredential, error) {
	if os.Getenv("AZURE_FEDERATED_TOKEN_FILE") != "" {
		return newWorkloadIdentityCredential(options)
	}
	return azidentity.NewDefaultAzureCredential(options)
}

type workloadIdentityCredential struct {
	assertion, file string
	cred            *azidentity.ClientAssertionCredential
	lastRead        time.Time
}

func newWorkloadIdentityCredential(options *azidentity.DefaultAzureCredentialOptions) (*workloadIdentityCredential, error) {
	clientID := os.Getenv("AZURE_CLIENT_ID")
	file := os.Getenv("AZURE_FEDERATED_TOKEN_FILE")

	w := &workloadIdentityCredential{file: file}
	cred, err := azidentity.NewClientAssertionCredential(options.TenantID, clientID, w.getAssertion, &azidentity.ClientAssertionCredentialOptions{ClientOptions: options.ClientOptions})
	if err != nil {
		return nil, err
	}
	w.cred = cred
	return w, nil
}

func (w *workloadIdentityCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	return w.cred.GetToken(ctx, opts)
}

func (w *workloadIdentityCredential) getAssertion(context.Context) (string, error) {
	if now := time.Now(); w.lastRead.Add(5 * time.Minute).Before(now) {
		content, err := os.ReadFile(w.file)
		if err != nil {
			return "", err
		}
		w.assertion = string(content)
		w.lastRead = now
	}
	return w.assertion, nil
}
