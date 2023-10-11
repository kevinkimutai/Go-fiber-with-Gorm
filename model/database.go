package model

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var err error

const ConnectionStr = "root:P@ssw0rd@tcp(127.0.0.1:3306)/mysqlgo?parseTime=true"

func InitMigration() {

	DB, err = gorm.Open(mysql.Open(ConnectionStr), &gorm.Config{})
	if err != nil {
		panic("failed to connect database: " + err.Error())
	}

	// Migrate the schema
	DB.AutoMigrate(&User{})
}
