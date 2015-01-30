package controllers

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (c *MainController) Get() {
	room := c.Input().Get("chat")
	gameHost, _ := beego.Config("String", "gameHost", "")
	webHost, _ := beego.Config("String", "webHost", "")

	if room == "" {
		c.Data["Room"] = GetGuid(time.Now().UnixNano())
	} else {
		c.Data["Room"] = room
	}

	c.Data["gameHost"] = gameHost
	c.Data["webHost"] = webHost

	c.TplNames = "index.tpl"
}

func GetGuid(id int64) string {
	idStr := strconv.Itoa(int(id))
	return GetMd5(idStr + strconv.Itoa(time.Now().Nanosecond()))
}
func GetMd5(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}
