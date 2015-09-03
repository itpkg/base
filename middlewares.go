package base

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func CurrentUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user", nil)
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
