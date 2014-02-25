package config

type Main struct {
	Web         webInfo
	StoreES     storeElasticsearchInfo
	StoreInflux storeInfluxdbInfo
	Stores      []string
}

type webInfo struct {
	ListenAddr string
}

type storeElasticsearchInfo struct {
	Host       string
	Port       int
	MaxPending int
	CarbonPort int
}

type storeInfluxdbInfo struct {
	Host     string
	Username string
	Password string
	Database string
}
