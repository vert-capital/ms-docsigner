package kafka_handlers

import (
	"app/entity"
	usecase_user "app/usecase/user"
	"encoding/json"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func CreateUser(msg kafka.Message, usecaseUser usecase_user.IUsecaseUser) error {

	var entityUser entity.EntityUser

	err := json.Unmarshal(msg.Value, &entityUser)

	if err != nil {
		return err
	}

	err = usecaseUser.Create(&entityUser)

	return err
}
