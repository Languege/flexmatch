// kafka
// @author LanguageY++2013 2023/5/11 09:42
// @company soulgame
package kafka

import (
	"time"
	"os"
	"os/signal"
	"syscall"
	"github.com/Shopify/sarama"
	"github.com/Languege/flexmatch/common/logger"
)

func NewAsyncProducer(brokerList []string)(producer sarama.AsyncProducer) {
	var(
		err error
	)

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err = sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		logger.Panicf("[kafka]Failed to open Kafka producer: %s", err)
	}

	go func() {
		for err := range producer.Errors() {
			//错误记录
			logger.Warnf("[kafka]kafka producer err %s", err)
		}
	}()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigterm

		if producer != nil {
			err = producer.Close()
			logger.Infof("[kafka]closing kafka producer err %v", err)
		}

		time.Sleep(time.Second *3)
		os.Exit(0)
	}()

	return
}

