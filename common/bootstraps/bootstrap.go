// bootstraps
// @author LanguageY++2013 2023/5/8 23:38
// @company soulgame
package bootstraps

import (
	"time"
)

func init() {
	InitLogger()

	InitEtcd()


	WaitExit(time.Second)
}