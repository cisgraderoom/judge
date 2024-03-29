package conn

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Connection() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", os.Getenv("MYSQL_ROOT_USERNAME"), os.Getenv("MYSQL_ROOT_PASSWORD"), os.Getenv("MYSQL_CONTAINER"), os.Getenv("MYSQL_DATABASE"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err.Error())
		return nil
	}
	return db
}
