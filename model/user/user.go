package user

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const ConnectionStr = "root:admin@tcp(127.0.0.1:3306)/mysqlgo"

type User struct {
	gorm.Model
	Fname string `json:"fName"`
	Lname string `json:"lName"`
	Email string `json:"email"`
}

func InitMigration() {

	db, err := gorm.Open(mysql.Open(ConnectionStr), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&User{})
}
