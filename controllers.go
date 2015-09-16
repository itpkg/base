package base

import (
	"github.com/itpkg/web"
)

type UserController struct {

}

func (p *UserController) SignIn(){

}
func(p *UserController)  SignUp(){

}
func(p *UserController)  ForgotPassword(){

}

func(p *UserController)  ResetPassword(){

}

func(p *UserController)  UpdateProfile(){

}

func(p *UserController)  Unlock(){

}

func(p *UserController)  Confirm(){

}

type SiteController struct {

}
func(p *SiteController)  Sitemap(){

}
func (p *SiteController) Rss(){

}

func init(){
	en:=web.NewEngine("")

	uc := &UserController{}
	en.POST("/users/sign_in",nil, uc.SignIn)
	en.POST("/users/sign_up",nil, uc.SignUp)
	en.POST("/users/confirm",nil, uc.Confirm)
	en.POST("/users/unlock",nil, uc.Unlock)
	en.POST("/users/forgot_password",nil, uc.ForgotPassword)
	en.POST("/users/reset_password",nil, uc.ResetPassword)
	en.POST("/users/profile",nil, uc.UpdateProfile)

	sc := &SiteController{}
	en.GET("/rss.atom",nil, sc.Rss)
	en.GET("/sitemap.xml.gz", nil, sc.Sitemap)

	web.Mount(en)
}
