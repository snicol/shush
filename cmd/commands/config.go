package commands

type Config struct {
	Storage StorageConfig `mapstructure:"storage"`
	Cache   *CacheConfig  `mapstructure:"cache"`
	Sync    *SyncConfig   `mapstructure:"sync"`
}

type StorageConfig struct {
	Type   string `mapstructure:"type"`
	Config struct {
		PMSKMSConfig  `mapstructure:",squash"`
		JSONGitConfig `mapstructure:",squash"`
	} `mapstructure:"config"`
}

type CacheConfig struct {
	Type   string `mapstructure:"type"`
	Config struct {
		KeychainConfig `mapstructure:",squash"`
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

type JSONGitConfig struct {
	Path       string `mapstruture:"path"`
	Filename   string `mapstructure:"filename"`
	RemoteName string `mapstructure:"remoteName"`
	Indent     string `mapstructure:"indent"`
}

type KeychainConfig struct {
	SecClass    string `mapstructure:"secClass"`
	Service     string `mapstructure:"service"`
	AccessGroup string `mapstructure:"accessGroup"`
}
