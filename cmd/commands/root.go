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

	profile string
	conf    Config

	rootCmd = &cobra.Command{
		Use:   "shush",
		Short: "Shush is an AWS secret manager tool. See `shush --help`.",
	}
)

func Execute() {
	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "Config profile to use")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.shush/")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	err = viper.UnmarshalKey(profile, &conf)
	if err != nil {
		panic(fmt.Errorf("unmarshal config file failed: %s", err))
	}

	storageProvider := getStorageProvider(conf.Storage)
	cacheProvider := getCacheProvider(conf.Cache)

	sess = shush.NewSession(storageProvider, cacheProvider, shush.UpsertVersionReplaceNewer)

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(syncCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func getStorageProvider(storageConf StorageConfig) storage.Provider {
	switch storageConf.Type {
	case "pmskms":
		sess, err := session.NewSessionWithOptions(session.Options{
			Profile:                 storageConf.Config.AWSProfile,
			AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
			SharedConfigState:       session.SharedConfigEnable,
		})
		if err != nil {
			panic(err)
		}

		return storage.NewPMSKMS(sess, storageConf.Config.KeyID)
	default:
		panic(fmt.Sprintf("unknown storage provider type %s", storageConf.Type))
	}
}

func getCacheProvider(cacheConf *CacheConfig) cache.Provider {
	if cacheConf == nil {
		return nil
	}

	switch cacheConf.Type {
	case "keychain":
		secClass := keychain.SecClassGenericPassword
		if cacheConf.Config.SecClass == "internet" {
			secClass = keychain.SecClassInternetPassword
		}

		return cache.NewKeychain(secClass, cacheConf.Config.Service, cacheConf.Config.AccessGroup)
	default:
		panic(fmt.Errorf("unknown cache provider type %s", cacheConf.Type))
	}
}
