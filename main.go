package main

import (
	_ "ddm-admin-console/routers"
	_ "ddm-admin-console/templatefunction"

	"github.com/astaxie/beego"
)

func main() {
	beego.Run()
}
