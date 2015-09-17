package base

import (
	"github.com/itpkg/web"
)

func GetShow() {

}
func PostSignIn() {

}
func PostSignUp() {

}
func PostForgotPassword() {

}

func PostResetPassword() {

}

func PostUpdateProfile() {

}

func PostUnlock() {

}

func PostConfirm() {

}

func GetUnlock() {

}

func GetConfirm() {

}

func GetSitemap() {

}
func GetRss() {

}

func init() {
	en := web.NewEngine("")

	en.POST("/users/sign_in", nil, PostSignIn)
	en.POST("/users/sign_up", nil, PostSignUp)
	en.POST("/users/confirm", nil, PostConfirm)
	en.GET("/users/confirm", nil, GetConfirm)
	en.POST("/users/unlock", nil, PostUnlock)
	en.GET("/users/unlock", nil, GetUnlock)
	en.POST("/users/forgot_password", nil, PostForgotPassword)
	en.POST("/users/reset_password", nil, PostResetPassword)
	en.POST("/users/profile", nil, PostUpdateProfile)
	en.GET(`/users/(?P<id>\d+)`, nil, GetShow)

	en.GET("/rss.atom", nil, GetRss)
	en.GET("/sitemap.xml.gz", nil, GetSitemap)

	web.Mount(en)
}
