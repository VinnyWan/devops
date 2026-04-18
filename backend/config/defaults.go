package config

import "github.com/spf13/viper"

// SetDefaults 设置配置默认值
// 确保即使无配置文件，核心服务也能以默认参数启动
// 注意：所有键名必须与 config.yaml 和代码中 config.Cfg.GetXxx() 调用保持一致
func SetDefaults(v *viper.Viper) {
	// Server 默认配置
	v.SetDefault("server.port", 8000)
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.enableSwagger", true)

	// 日志默认配置（与 config.yaml 中 log.* 对齐）
	v.SetDefault("log.level", "info")
	v.SetDefault("log.output", "console")
	v.SetDefault("log.filePath", "./logs/app.log")
	v.SetDefault("log.enableCaller", true)
	v.SetDefault("log.enableStacktrace", true)

	// 数据库默认配置（与 bootstrap/db.go 中 db.* 对齐）
	v.SetDefault("db.dialects", "mysql")
	v.SetDefault("db.host", "127.0.0.1")
	v.SetDefault("db.port", 3306)
	v.SetDefault("db.db", "devops")
	v.SetDefault("db.username", "root")
	v.SetDefault("db.password", "")
	v.SetDefault("db.charset", "utf8mb4")
	v.SetDefault("db.maxIdle", 10)
	v.SetDefault("db.maxOpen", 100)

	// Redis 默认配置
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.password", "")
	v.SetDefault("redis.db", 0)

	// Session 默认配置
	v.SetDefault("session.expire", 7200)
	v.SetDefault("auth.enable_external", false)

	// 加密默认配置
	v.SetDefault("crypto.secret", "")

	// Terminal 默认配置
	v.SetDefault("terminal.recording_dir", "./data/recordings/cmdb-terminal")
	v.SetDefault("terminal.max_session_duration", 86400)
	v.SetDefault("terminal.idle_timeout", 300)
	v.SetDefault("terminal.known_hosts_path", "")

	// LDAP 默认配置（默认关闭）
	v.SetDefault("ldap.enable", false)
	v.SetDefault("ldap.host", "ldap.example.com")
	v.SetDefault("ldap.port", 389)
	v.SetDefault("ldap.base_dn", "dc=example,dc=com")
	v.SetDefault("ldap.bind_dn", "cn=admin,dc=example,dc=com")
	v.SetDefault("ldap.bind_password", "")
	v.SetDefault("ldap.user_filter", "(&(uid=%s)(objectClass=person))")
	v.SetDefault("ldap.attributes.username", "uid")
	v.SetDefault("ldap.attributes.email", "mail")
	v.SetDefault("ldap.attributes.nickname", "cn")

	// OIDC 默认配置（默认关闭）
	v.SetDefault("oidc.enable", false)
	v.SetDefault("oidc.provider", "")
	v.SetDefault("oidc.client_id", "")
	v.SetDefault("oidc.client_secret", "")
	v.SetDefault("oidc.redirect_url", "")
	v.SetDefault("oidc.scopes", []string{"openid", "profile", "email"})

	// CORS 默认配置
	v.SetDefault("cors.allow_origins", []string{"http://localhost:3000"})
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allow_headers", []string{"Content-Type", "Authorization", "X-Token"})
	v.SetDefault("cors.expose_headers", []string{"Set-Cookie"})

	// Nacos 默认配置（默认关闭）
	v.SetDefault("nacos.enable", false)
	v.SetDefault("nacos.host", "127.0.0.1")
	v.SetDefault("nacos.port", 8848)
	v.SetDefault("nacos.namespace", "public")
	v.SetDefault("nacos.data_id", "devops-platform")
	v.SetDefault("nacos.group", "DEFAULT_GROUP")
	v.SetDefault("nacos.timeout_ms", 5000)
	v.SetDefault("nacos.server_addrs", "")
	v.SetDefault("nacos.username", "")
	v.SetDefault("nacos.password", "")
	v.SetDefault("nacos.register_enable", false)
	v.SetDefault("nacos.service_name", "devops-backend")
	v.SetDefault("nacos.service_ip", "")
	v.SetDefault("nacos.service_port", 0)
	v.SetDefault("nacos.service_group", "DEFAULT_GROUP")
	v.SetDefault("nacos.service_cluster", "DEFAULT")
	v.SetDefault("nacos.service_weight", 1)
	v.SetDefault("nacos.service_ephemeral", true)
	v.SetDefault("nacos.service_metadata", map[string]string{
		"protocol": "http",
		"version":  "v1",
	})

	// Cloud 云账号同步默认配置
	v.SetDefault("cloud.sync_concurrency", 5)
	v.SetDefault("cloud.sync_timeout", 300)
	v.SetDefault("cloud.default_regions", "ap-guangzhou,ap-shanghai,ap-beijing")
}
