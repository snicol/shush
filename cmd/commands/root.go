package commands

import (
	"fmt"
	"os"

	shush "shush/lib"
	"shush/lib/cache"
	"shush/lib/storage"

	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/keybase/go-keychain"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	sess *shush.Session

	rootCmd = &cobra.Command{
		Use:   "shush",
		Short: "Shush is an AWS secret manager tool. See `shush --help`.",
	}
)

func Execute() {
	var profile string
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "Config profile to use")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.shush/")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	var conf Config
	err = viper.UnmarshalKey(profile, &conf)
	if err != nil {
		panic(fmt.Errorf("unmarshal config file failed: %s", err))
	}

	var storageProvider storage.Provider

	switch conf.Storage.Type {
	case "pmskms":
		sess, err := session.NewSessionWithOptions(session.Options{
			Profile:                 conf.Storage.Config.AWSProfile,
			AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
			SharedConfigState:       session.SharedConfigEnable,
		})
		if err != nil {
			panic(err)
		}

		storageProvider = storage.NewPMSKMS(sess, conf.Storage.Config.KeyID)
	default:
		panic(fmt.Sprintf("unknown storage provider type %s", conf.Storage.Type))
	}

	var cacheProvider cache.Provider

	switch conf.Cache.Type {
	case "keychain":
		secClass := keychain.SecClassGenericPassword
		if conf.Cache.Config.SecClass == "internet" {
			secClass = keychain.SecClassInternetPassword
		}

		cacheProvider = cache.NewKeychain(secClass, conf.Cache.Config.Service, conf.Cache.Config.AccessGroup)
	default:
		panic(fmt.Errorf("unknown cache provider type %s", conf.Cache.Type))
	}

	sess = shush.NewSession(cacheProvider, storageProvider, shush.UpsertVersionReplaceNewer)

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
