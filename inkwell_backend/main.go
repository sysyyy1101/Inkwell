package main

import (
	"inkwell_backend/dao/mysql"
	"inkwell_backend/dao/redis"
	"inkwell_backend/logger"
	"inkwell_backend/pkg/jwt"
	"inkwell_backend/pkg/snowflake"
	"inkwell_backend/routers"
	"inkwell_backend/settings"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// Swagger 文档的通用信息(只用于生成文档, 不影响程序运行)。
// @title Inkwell 社区 API
// @version 0.1.1
// @description 基于 gin + MySQL + Redis + JWT 的社区后端接口文档。
// @description 所有接口返回统一结构: {"code":1000,"message":"success","data":{}}。业务失败时 HTTP 状态码仍然是 200(只有访问不存在的路由才返回 404), 需要根据 code 判断结果, code 取值见 controller/code.go。
// @description 需要登录的接口要在请求头带上 Authorization: Bearer <accessToken>; access token 失效(1006)时可以调用 /api/v1/refresh_token 刷新后重放请求。
// @description 所有 64 位 ID(用户/帖子/评论)在 JSON 中都使用字符串传输, 用数字传会丢精度。
// @BasePath /api/v1
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description 取值格式 "Bearer <accessToken>"
func main() {
	confFile := flag.String("conf", "./conf/config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	if err := settings.Init(*confFile); err != nil {
		fmt.Printf("load config failed, err:%v\n", err)
		os.Exit(1)
	}
	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		fmt.Printf("init logger failed, err:%v\n", err)
		os.Exit(1)
	}
	// JWT 密钥: dev 模式可以用内置开发密钥, 其它模式必须显式配置
	if err := jwt.Init(settings.Conf.JWTSecret, settings.Conf.Mode); err != nil {
		zap.L().Error("init jwt failed", zap.Error(err))
		fmt.Printf("init jwt failed, err:%v\n", err)
		os.Exit(1)
	}
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		fmt.Printf("init mysql failed, err:%v\n", err)
		os.Exit(1)
	}
	defer mysql.Close() // 程序退出关闭数据库连接
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		fmt.Printf("init redis failed, err:%v\n", err)
		os.Exit(1)
	}
	defer redis.Close()
	// 雪花算法的机器 ID 来自配置(多实例部署时每个实例必须不同)
	if err := snowflake.Init(settings.Conf.MachineID); err != nil {
		fmt.Printf("init snowflake failed, err:%v\n", err)
		os.Exit(1)
	}
	zap.L().Info("init snowflake success", zap.Uint16("machine_id", settings.Conf.MachineID))
	// 注册路由
	r := routers.SetupRouter()
	addr := fmt.Sprintf(":%d", settings.Conf.Port)
	if err := r.Run(addr); err != nil {
		fmt.Printf("run server failed, err:%v\n", err)
		os.Exit(1)
	}
}
