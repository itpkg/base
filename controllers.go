package base

import (
	"github.com/itpkg/web"
)

func Show() {

}
func SignIn() {

}
func SignUp() {

}
func ForgotPassword() {

}

func ResetPassword() {

}

func UpdateProfile() {

}

func Unlock() {

}

func Confirm() {

}

func Sitemap() {

}
func Rss() {

}

func init() {
	en := web.NewEngine("")

	en.POST("/users/sign_in", nil, SignIn)
	en.POST("/users/sign_up", nil, SignUp)
	en.POST("/users/confirm", nil, Confirm)
	en.POST("/users/unlock", nil, Unlock)
	en.POST("/users/forgot_password", nil, ForgotPassword)
	en.POST("/users/reset_password", nil, ResetPassword)
	en.POST("/users/profile", nil, UpdateProfile)
	en.GET(`/users/(?P<id>\d+)`, nil, Show)

	en.GET("/rss.atom", nil, Rss)
	en.GET("/sitemap.xml.gz", nil, Sitemap)

	web.Mount(en)
}
