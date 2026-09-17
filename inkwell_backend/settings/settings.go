package settings

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var Conf = new(AppConfig)

type AppConfig struct {
	Mode string `mapstructure:"mode"`
	Port int    `mapstructure:"port"`
	// MachineID 雪花算法使用的机器 ID。多实例部署时每个实例必须不同, 否则会生成重复的业务 ID
	MachineID uint16 `mapstructure:"machine_id"`
	// JWTSecret JWT 签名密钥, 非 dev 模式下必须显式配置(配置文件或环境变量)
	JWTSecret             string `mapstructure:"jwt_secret"`
	*LogConfig            `mapstructure:"log"`
	*MySQLConfig          `mapstructure:"mysql"`
	*RedisConfig          `mapstructure:"redis"`
	*LoginRateLimitConfig `mapstructure:"login_rate_limit"`
}

// 环境变量名: 部署时优先用环境变量覆盖配置文件里的敏感/部署相关字段
const (
	// JWTSecretEnv JWT 签名密钥对应的环境变量
	JWTSecretEnv = "INKWELL_JWT_SECRET"
	// MachineIDEnv 雪花算法机器 ID 对应的环境变量
	MachineIDEnv = "INKWELL_MACHINE_ID"
)

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DB           string `mapstructure:"db"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host         string `mapstructure:"host"`
	Password     string `mapstructure:"password"`
	Port         int    `mapstructure:"port"`
	DB           int    `mapstructure:"db"`
	PoolSize     int    `mapstructure:"pool_size"`
	MinIdleConns int    `mapstructure:"min_idle_conns"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// LoginRateLimitConfig 登录接口的令牌桶限流参数。
// 整个进程只有一个桶(不是每个 IP 一个桶): rate 是每秒补充的令牌数, burst 是桶容量。
type LoginRateLimitConfig struct {
	Rate  float64 `mapstructure:"rate"`
	Burst int     `mapstructure:"burst"`
}

// Init 读取配置文件并监听文件变化。
// confFile 为空时使用默认的 ./conf/config.yaml。
func Init(confFile string) error {
	if confFile != "" {
		viper.SetConfigFile(confFile)
	} else {
		viper.SetConfigFile("./conf/config.yaml")
	}

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("ReadInConfig failed, err: %v", err)
	}
	if err := viper.Unmarshal(Conf); err != nil {
		return fmt.Errorf("unmarshal to Conf failed, err:%v", err)
	}
	applyEnvOverrides()
	if err := validate(); err != nil {
		return err
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("夭寿啦~配置文件被人修改啦...")
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("配置文件解析失败, 继续使用旧配置, err:%v\n", err)
			return
		}
		// 环境变量的优先级比配置文件高, 重新读配置后要再叠加一次
		applyEnvOverrides()
	})
	return nil
}

// applyEnvOverrides 用环境变量覆盖配置文件里的值(环境变量优先)。
func applyEnvOverrides() {
	if secret := strings.TrimSpace(os.Getenv(JWTSecretEnv)); secret != "" {
		Conf.JWTSecret = secret
	}
	if raw := strings.TrimSpace(os.Getenv(MachineIDEnv)); raw != "" {
		id, err := strconv.ParseUint(raw, 10, 16)
		if err != nil {
			fmt.Printf("环境变量 %s=%q 不是合法的机器 ID(0-65535), 继续使用配置文件里的值\n", MachineIDEnv, raw)
			return
		}
		Conf.MachineID = uint16(id)
	}
}

// validate 校验启动必需的配置项, 避免把有问题的配置带到线上。
func validate() error {
	// 机器 ID 必须显式配置: 用固定默认值会让多个实例生成相同的 ID(唯一索引冲突 / 数据串号)
	if Conf.MachineID < 1 || Conf.MachineID > 65535 {
		return fmt.Errorf("machine_id 必须是 1-65535(当前 %d), 多实例部署时每个实例要配置不同的值, "+
			"也可以用环境变量 %s 覆盖", Conf.MachineID, MachineIDEnv)
	}
	return nil
}
