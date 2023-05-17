// bootstraps
// @author LanguageY++2013 2023/5/17 10:51
// @company soulgame
package bootstraps

import (
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitExit(waitTime time.Duration) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-c

		if waitTime > 0 {
			time.Sleep(waitTime)
		}
		os.Exit(0)
	}()
}
