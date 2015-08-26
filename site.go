package base

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/op/go-logging"
)

type SiteEngine struct {
	Cfg    *Configuration  `inject:""`
	Db     *gorm.DB        `inject:""`
	Router *gin.Engine     `inject:""`
	Logger *logging.Logger `inject:""`
}

func (p *SiteEngine) Job() {

}
func (p *SiteEngine) Mount() {

}
func (p *SiteEngine) Migrate() {

}
func (p *SiteEngine) Info() (name string, version string, desc string) {
	return "site", "v20150826", "Site engine"
}

//-----------------------------------------------------------------------------

func init() {
	Register(map[string]interface{}{"engine.site": SiteEngine{}})
}
