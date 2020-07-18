# shush

Shush is a secret manager which allows writing and syncronisation between
providers. It is similar to rclone in that it is designed to work with a number
of backends. The project is split into two distinct areas, `storage` providers
and `cache` providers. It also provides Go helpers to load in secrets to a
given struct using struct tags, or via specific ENV variables.

## Supported providers

Currently the only supported storage backend is AWS SSM Parameter Store w/ KMS.
The only supported cache store is the local keychain.

## Creating new providers

New providers and contributions to the existing providers are greatly
appreciated. There is a standard interface which should cover most use cases and
should be relatively simple to add new providers.

## CLI

### Config

The config file should live at `~/.shush/config.yml`. Here is a basic example
for AWS PMS+KMS and Keychain cache:

```yaml
default:
  storage:
    type: pmskms
    config:
      keyId: 0a000000-0000-0000-0000-000000000000 # your KMS key ID
      awsProfile: my_aws_profile
      awsRegion: eu-west-1
  cache:
    type: keychain
    config:
      securityClass: internet
      service: my-app
      accessGroup: com.my-app.secrets
```

You can specify additional profiles and use them at runtime using
`-p`/`--profile`. For example `-p prod`.

### Usage

To set a secret:

	shush set <key> <value>

To get a secret:

	shush get <key>

## Programmatic examples

### Structs

Go example for AWS PMS KMS and Keychain, loading secrets into structs:

```go
type MyConfig struct {
	SomeSecret string `shush:"my-dev-env.some-secret"`
}

func main() {
	// create storage and cache providers
	storageProvider := storage.NewPMSKMS(awsSession, "your-kms-key-id-here")
	cacheProvider := cache.NewKeychain(keychain.SecClassGenericPassword, "example-app", "com.example-app.secrets")

	// create a new shush session
	ssh := shush.NewSession(storageProvider, cacheProvider, shush.UpsertVersionReplaceDifferent)

	// unmarshal your config
	conf := MyConfig{}
	err = ssh.UnmarshalContext(ctx, &ex)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(conf.SomeSecret) // plaintext value
}
```

### Environment variables

Shush provides a way of using an environment variable to describe which key to
fetch to circumvent specifying secrets directly on an environment variable in an
unsafe manner.

For example:

	export MY_SECRET=shush://dev.my-secret

Then to retrieve:

```go
func main() {
	// create storage and cache providers
	storageProvider := storage.NewPMSKMS(awsSession, "your-kms-key-id-here")
	cacheProvider := cache.NewKeychain(keychain.SecClassGenericPassword, "example-app", "com.example-app.secrets")

	// create a new shush session
	ssh := shush.NewSession(storageProvider, cacheProvider, shush.UpsertVersionReplaceDifferent)

	mySecret, err := ssh.GetenvContext(ctx, "MY_SECRET")
	if err != nil {
		t.Fatal(err)
	}

	log.Println(mySecret) // plaintext value
}
```
