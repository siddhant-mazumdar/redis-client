package config

import (
	"errors"
)

type messagingQueueConfig struct {
	Topic     string
	Partition int32
}

func (c config) GetTopic(topic string) (messagingQueueConfig, error) {
	for _, topicConfig := range c.MessagingQueueTopics {
		if topicConfig.Topic == topic {
			return topicConfig, nil
		}
	}
	return messagingQueueConfig{}, errors.New("")

}
