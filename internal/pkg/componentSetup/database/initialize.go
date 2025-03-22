package database

import (
	"go-faster-gateway/pkg/database"
	"go-faster-gateway/pkg/log"
	slogger "go-faster-gateway/pkg/log/logger"
	"time"

	"github.com/acmestack/gorm-plus/gplus"
	"gorm.io/gorm"
	sgorm "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// SetupDb 配置数据库
func SetupDb(dbConfig *database.Database) {
	setupSimpleDatabase(dbConfig)
}

func setupSimpleDatabase(c *database.Database) {
	//log.Infof("%s => %s", host, utils.Green(c.Source))
	registers := make([]database.ResolverConfigure, len(c.Registers))
	for i := range c.Registers {
		registers[i] = database.NewResolverConfigure(
			c.Registers[i].Sources,
			c.Registers[i].Replicas,
			c.Registers[i].Policy,
			c.Registers[i].Tables)
	}
	resolverConfig := database.NewConfigure(c.Source, c.MaxIdleCons, c.MaxOpenCons, c.ConnMaxIdleTime, c.ConnMaxLifeTime, registers)
	db, err := resolverConfig.Init(&gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "", // 表前缀
			SingularTable: true,
		},
		Logger: database.New(
			sgorm.Config{
				SlowThreshold: time.Second,
				Colorful:      true,
				LogLevel: sgorm.LogLevel(
					slogger.DefaultLogger.Options().Level.LevelForGorm()),
			},
		),
	}, database.Opens[c.Driver])

	if err != nil {
		log.Log.WithError(err).Fatal(c.Driver + " connect error :")
	} else {
		log.Log.Info(c.Driver + " connect success !")
	}
	// 初始化
	gplus.Init(db)
}
