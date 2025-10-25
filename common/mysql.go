package common

import (
	config2 "git.imooc.com/zhanshen1614/product/internal/config"
	"github.com/micro/go-micro/v2/config"
)

// 从Consul获取MySQL配置
func GetMySqlFromConsul(configInfo config.Config, path ...string) (*config2.MySqlConfig, error) {
	mysqlConfig := &config2.MySqlConfig{}
	err := configInfo.Get(path...).Scan(mysqlConfig)
	if err != nil {
		return nil, err
	}
	return mysqlConfig, nil
}
