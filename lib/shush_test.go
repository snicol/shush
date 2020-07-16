package shush_test

import (
	"context"
	"testing"
	"time"

	shush "shush/lib"
	"shush/lib/cache"
	"shush/lib/storage"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/keybase/go-keychain"
)

type Example struct {
	MySecret string `shush:"my-dev-env.my-secret"`
}

func TestUnmarshal(t *testing.T) {
	t.Skip("skipping integration test")

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	sess, err := session.NewSessionWithOptions(session.Options{
		Profile:                 "integration_profile",
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	})
	if err != nil {
		t.Fatal(err)
	}

	storageProvider := storage.NewPMSKMS(sess, "5ab872f6-b721-41f6-9c1a-9aa699212ea4")
	cacheProvider := cache.NewKeychain(keychain.SecClassGenericPassword, "example-app", "com.example-app.secrets")

	ssh := shush.NewSession(cacheProvider, storageProvider, shush.UpsertVersionReplaceDifferent)

	ex := Example{}

	err = ssh.UnmarshalContext(ctx, &ex)
	if err != nil {
		t.Fatal(err)
	}
}
