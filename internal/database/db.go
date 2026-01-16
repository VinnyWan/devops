// internal/database/db.go
package database

import (
	"context"
	"fmt"
	"time"

	"devops/common/config"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB

// Init 初始化数据库连接
func InitMysql() error {
	dbConfig := config.GetConfig()
	if dbConfig.MaxIdle <= 0 {
		dbConfig.MaxIdle = 10
	}
	if dbConfig.MaxOpen <= 0 {
		dbConfig.MaxOpen = 100
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local&timeout=5s&readTimeout=5s&writeTimeout=5s&sql_mode=''",
		dbConfig.Username,
		dbConfig.Password,
		dbConfig.Host,
		dbConfig.Port,
		dbConfig.Db,
		dbConfig.Charset)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return err
	}

	Db = db

	sqlDB, err := Db.DB()
	if err != nil {
		return err
	}

	sqlDB.SetMaxIdleConns(dbConfig.MaxIdle)
	sqlDB.SetMaxOpenConns(dbConfig.MaxOpen)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	sqlDB.SetConnMaxLifetime(2 * time.Hour)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return err
	}
	return nil
}
