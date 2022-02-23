package cluster

const configFile = "hcloudtalosconfig"

type Config struct {
	ClusterName string       `yaml:"clusterName"`
	Hcloud      ConfigHcloud `yaml:"hcloud"`
}

type ConfigHcloud struct {
	Location    string `yaml:"location"`
	NetworkZone string `yaml:"networkZone"`
	Token       string `yaml:"token"`
}
