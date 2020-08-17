package shush_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	shush "shush/lib"
	"shush/lib/cache"
	"shush/lib/storage"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/keybase/go-keychain"
)

var (
	ctx    context.Context
	cancel context.CancelFunc
	sess   *session.Session
)

func init() {
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)

	var err error
	sess, err = session.NewSessionWithOptions(session.Options{
		Profile:                 "integration_profile",
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		SharedConfigState:       session.SharedConfigEnable,
	})
	if err != nil {
		log.Fatal(err)
	}
}

type Example struct {
	MySecret string `shush:"my-dev-env.my-secret"`
}

func TestUnmarshal(t *testing.T) {
	t.Skip("skipping integration test")

	storageProvider := storage.NewPMSKMS(sess, "5ab872f6-b721-41f6-9c1a-9aa699212ea4")
	cacheProvider := cache.NewKeychain(keychain.SecClassGenericPassword, "example-app", "com.example-app.secrets")

	ssh := shush.NewSession(storageProvider, cacheProvider, shush.UpsertVersionReplaceDifferent)

	ex := Example{}

	err := ssh.UnmarshalContext(ctx, &ex)
	if err != nil {
		t.Fatal(err)
	}
}

func TestEnv(t *testing.T) {
	t.Skip("skipping integration test")

	k := "MY_SECRET"
	os.Setenv(k, "shush://dev.my-secret")

	storageProvider := storage.NewPMSKMS(sess, "5ab872f6-b721-41f6-9c1a-9aa699212ea4")
	cacheProvider := cache.NewKeychain(keychain.SecClassGenericPassword, "example-app", "com.example-app.secrets")

	ssh := shush.NewSession(storageProvider, cacheProvider, shush.UpsertVersionReplaceDifferent)

	_, err := ssh.GetenvContext(ctx, k)
	if err != nil {
		t.Fatal(err)
	}
}
