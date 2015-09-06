package base

import (
	"fmt"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
)

func AuthRequired(name string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if user := CurrentUser(c); user != nil && user.Is(name) {
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
		}

	}
}

func SetCurrentUser(helper *Helper, cfg *Configuration, logger *logging.Logger) gin.HandlerFunc {
	loge := func(err error) {
		switch err {
		case jwt.ErrNoTokenInRequest:
		case jwt.ErrInvalidKey:
		default:
			logger.Error(fmt.Sprintf("parse token - %v", err))
		}
	}
	return func(c *gin.Context) {

		if token, err := helper.TokenParse(c.Request); err == nil {
			db := Db(c)
			var user User
			if db.Model(User{}).Where("uid = ?", token["user"]).First(&user).RecordNotFound() {
				c.Set("user", nil)
			} else {
				db.Model(&user).Related(&user.Permissions)
				for i, _ := range user.Permissions {
					db.Model(&user.Permissions[i]).Related(&user.Permissions[i].Role)
				}

				//				for _, pm := range user.Permissions{
				//					logger.Debug(fmt.Sprintf("PERMISSION %d=>%s %v %v", pm.RoleID,pm.Role.Name, pm.StartUp, pm.ShutDown))
				//				}

				c.Set("user", &user)
			}
		} else {
			loge(err)
			c.Set("user", nil)
		}

		c.Next()

		if err := helper.TokenTtl(c.Request, cfg.Http.Expire); err != nil {
			loge(err)
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
