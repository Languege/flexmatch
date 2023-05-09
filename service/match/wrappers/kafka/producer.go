// kafka_wrapper
// @author LanguageY++2013 2023/5/6 15:34
// @company soulgame
package kafka_wrapper

import (
	"github.com/Shopify/sarama"
	"github.com/spf13/viper"
	"time"
	"log"
	"os"
	"os/signal"
	"syscall"
	"github.com/Languege/flexmatch/common/logger"
	"strings"
	"github.com/rcrowley/go-metrics"
)

func NewAsyncProducer()(producer sarama.AsyncProducer) {
	if !viper.IsSet("kafka.bootstrapServers") {
		log.Panicf("kafka.bootstrapServers not setting")
	}
	brokerList := viper.GetStringSlice("kafka.bootstrapServers")
	var(
		err error
	)

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err = sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		log.Panicf("[kafka]Failed to open Kafka producer: %s", err)
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

//NewProducer	同步生产者
func NewProducer(params... interface{})(producer sarama.SyncProducer) {
	if !viper.IsSet("kafka.bootstrapServers") {
		log.Panicf("kafka.bootstrapServers not setting")
	}
	brokerList := viper.GetStringSlice("kafka.bootstrapServers")
	var(
		err error
	)
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Return.Successes = true
	config.MetricRegistry = metrics.DefaultRegistry

	producer, err = sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Panicf("[kafka]Failed to open Kafka producer: %s", err)
	}

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	go func() {
		<-sigterm
		logger.Infof("[kafka]closing kafka producer")
		if producer != nil {
			err = producer.Close()
			logger.Infof("[kafka]closing kafka producer err %v", err)
		}

		time.Sleep(time.Second *3)
		os.Exit(0)
	}()

	return
}

func MessageAdaptor(message *sarama.ProducerMessage) {
	if prefix := viper.GetString("kafka.prefix"); prefix != "" {
		message.Topic = prefix + message.Topic
	}

	if viper.GetBool("kafka.toLower") {
		message.Topic = strings.ToLower(message.Topic)
	}
}