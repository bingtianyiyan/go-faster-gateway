package database

type Database struct {
	DbAlisName      string // 数据库昵称如果有多库则有用
	Driver          string
	Source          string
	ConnMaxIdleTime int
	ConnMaxLifeTime int
	MaxIdleCons     int
	MaxOpenCons     int
	Registers       []DBResolverConfig
}

type DBResolverConfig struct {
	Sources  []string
	Replicas []string
	Policy   string
	Tables   []string
}

var (
	DatabaseConfig = new(Database)
	//DatabasesConfig = make(map[string]*Database)
)
