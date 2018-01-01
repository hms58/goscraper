// 2017/12/29 21:15:58 Fri
package goscraper

import (
	"github.com/astaxie/beego/logs"
)

var (
	l = logs.NewLogger(10000)
)

func init() {
	l.SetLogger("console", "")
	l.EnableFuncCallDepth(true)
	l.SetLevel(logs.LevelDebug)
}
