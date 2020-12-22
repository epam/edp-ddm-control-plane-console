package test

import (
	"io/ioutil"
	"os"

	"github.com/astaxie/beego"
)

func InitBeego() (err error) {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	fs, err := ioutil.ReadDir(wd)
	if err != nil {
		return err
	}

	defer func() {
		if err == nil {
			beego.TestBeegoInit(".")
			beego.BConfig.WebConfig.EnableXSRF = false
			beego.BConfig.WebConfig.Session.SessionOn = false
		}
	}()

	for _, v := range fs {
		if v.Name() == "conf" {
			return
		}
	}

	err = os.Chdir("..")
	return
}
