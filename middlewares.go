package base

import (
	"fmt"
	"log/syslog"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func SetCurrentUser(helper *Helper, cfg *Configuration, log *syslog.Writer) gin.HandlerFunc {
	return func(c *gin.Context) {

		bearer := c.Request.Header.Get("Authorization")
		pre := "Bearer "

		if strings.HasPrefix(bearer, pre) {
			ticket := bearer[len(pre):len(bearer)]
			if token, err := helper.TokenParse(ticket); err == nil {
				db := Db(c)
				var user User
				if db.Model(User{}).Where("uid = ?", token["user"]).First(&user).RecordNotFound() {
					c.Set("user", nil)
				} else {
					helper.TokenTtl(user.Uid, cfg.Http.Expire)
					c.Set("user", &user)
				}
			} else {
				log.Err(fmt.Sprintf("parse token: %v [%s]", err, ticket))
				c.Set("user", nil)
			}
		} else {
			c.Set("user", nil)

		}
		c.Next()
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
