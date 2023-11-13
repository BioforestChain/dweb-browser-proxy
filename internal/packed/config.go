// Package config 负责配置信息
package packed

import (
	"context"
	"fmt"
	"proxyServer/internal/helper/config"
)

type Config struct {
	Ctx context.Context
}

func (s *Config) GetDbNames(filename string) []string {
	DbNames := make([]string, 0)
	path := fmt.Sprintf("%sconfig/%s/%s.yml", ProjectPath(), DevEnv, filename)
	//jwtExpire, _ := g.Cfg().Get(s.Ctx, "redis")
	DBConfigs, err := config.GetConfig(path)
	configList, err := DBConfigs.Map(filename)
	if err == nil {
		for DBName, _ := range configList {
			DbNames = append(DbNames, DBName)
		}
	}

	return DbNames
}
