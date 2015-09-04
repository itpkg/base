package base

import (
	"fmt"
	"log/syslog"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetCurrentUser(helper *Helper, cfg *Configuration, log *syslog.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {

		if token, err := helper.TokenParse(c.Request); err == nil {
			db := Db(c)
			var user User
			if db.Model(User{}).Where("uid = ?", token["user"]).First(&user).RecordNotFound() {
				c.Set("user", nil)
			} else {
				c.Set("user", &user)
			}
		} else {
			log.Err(fmt.Sprintf("parse token: %v", err))
			c.Set("user", nil)
		}

		c.Next()

		if err := helper.TokenTtl(c.Request, cfg.Http.Expire); err != nil {
			log.Err(fmt.Sprintf("parse token: %v", err))
		}
	}
}

func SetLocale() gin.HandlerFunc {
	return func(c *gin.Context) {
		locale := c.DefaultQuery("locale", "en_US")
		c.Set("locale", locale)
		c.Next()
	}
}

func SetTransactions(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		tx := db.Begin()
		c.Set("db", tx)
		c.Next()
		tx.Commit()
	}
}
