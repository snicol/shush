package commands

import (
	"fmt"
	"log"
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
	conf    *Config

	rootCmd = &cobra.Command{
		Use:   "shush",
		Short: "Shush is an AWS secret manager tool. See `shush --help`.",
		Run:   func(cmd *cobra.Command, args []string) {},
	}
)

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "Config profile to use")

	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(setCmd)
	rootCmd.AddCommand(syncCmd)
}

func initConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("$HOME/.shush/")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error config file: %s", err)
	}

	err = viper.UnmarshalKey(profile, &conf)
	if err != nil {
		log.Fatalf("unmarshal config file failed: %s", err)
	}

	if conf == nil {
		log.Fatalf("no profile %s specified in config", profile)
	}

	storageProvider := getStorageProvider(conf.Storage)
	cacheProvider := getCacheProvider(conf.Cache)

	sess = shush.NewSession(storageProvider, cacheProvider, shush.UpsertVersionReplaceNewer)
}

func Execute() {
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
			log.Fatal(err)
		}

		return storage.NewPMSKMS(sess, storageConf.Config.KeyID)

	case "jsongit":
		return storage.NewJSONGit(storageConf.Config.Path, storageConf.Config.Filename,
			storageConf.Config.RemoteName, storageConf.Config.Indent)
	default:
		log.Fatal(fmt.Sprintf("unknown storage provider type %s", storageConf.Type))
		return nil
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
		log.Fatalf("unknown cache provider type %s", cacheConf.Type)
		return nil
	}
}
