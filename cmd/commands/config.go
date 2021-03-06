package commands

type Config struct {
	Storage StorageConfig `mapstructure:"storage"`
	Cache   *CacheConfig  `mapstructure:"cache"`
	Sync    *SyncConfig   `mapstructure:"sync"`
}

type StorageConfig struct {
	Type   string `mapstructure:"type"`
	Config struct {
		PMSKMSConfig `mapstructure:",squash"`
	} `mapstructure:"config"`
}

type CacheConfig struct {
	Type   string `mapstructure:"type"`
	Config struct {
		KeybaseConfig `mapstructure:",squash"`
	} `mapstructure:"config"`
}

type SyncConfig struct {
	Prefixes []string `mapstructure:"prefixes"`
}

type PMSKMSConfig struct {
	KeyID      string `mapstructure:"keyId"`
	AWSProfile string `mapstructure:"awsProfile"`
	AWSRegion  string `mapstructure:"awsRegion"`
}

type KeybaseConfig struct {
	SecClass    string `mapstructure:"secClass"`
	Service     string `mapstructure:"service"`
	AccessGroup string `mapstructure:"accessGroup"`
}
