package setting

import (
	"blockchain-event-plugin/logger"
	"github.com/spf13/viper"
)

// init
func init() {
	// 初始化配置文件
	if err := Config(); err != nil {
		logger.Error("Load configuration failed, ", err)
	}
}

// GetString 获取字符串类型的配置
func GetString(params string) string {
	return viper.GetString(params)
}

// GetInt 获取INT类型的配置
func GetInt(params string) int {
	return viper.GetInt(params)
}

// GetBool 获取布尔类型的配置
func GetBool(params string) bool {
	return viper.GetBool(params)
}

// Config viper解析配置文件
func Config() error {
	viper.AddConfigPath("./config")
	viper.SetConfigName("conf")
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	return nil
}
