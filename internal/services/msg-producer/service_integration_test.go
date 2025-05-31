//go:build integration

package msgproducer_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"

	"github.com/dndev-xx/go-ninja-chat/internal/logger"
	msgproducer "github.com/dndev-xx/go-ninja-chat/internal/services/msg-producer"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

type ServiceIntegrationSuite struct {
	testingh.KafkaSuite

	messagesTopic    string
	messagesConsumer *kafka.Reader
}

func TestServiceIntegrationSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceIntegrationSuite))
}

func (s *ServiceIntegrationSuite) SetupTest() {
	s.KafkaSuite.SetupTest()

	s.messagesTopic = fmt.Sprintf("%s.%d", "chat.messages", time.Now().UnixMilli())
	s.RecreateTopics(s.messagesTopic)

	const testName = "msgproducer_test.ServiceIntegrationSuite"
	s.messagesConsumer = kafka.NewReader(kafka.ReaderConfig{
		Brokers:     s.KafkaBrokers(),
		GroupID:     testName,
		Topic:       s.messagesTopic,
		StartOffset: kafka.FirstOffset,
		Logger:      logger.NewKafkaAdapted().WithServiceName(testName),
		ErrorLogger: logger.NewKafkaAdapted().WithServiceName(testName).ForErrors(),
	})
}

func (s *ServiceIntegrationSuite) TearDownTest() {
	if s.messagesConsumer != nil {
		s.NoError(s.messagesConsumer.Close())
	}
	s.KafkaSuite.TearDownTest()
}

func (s *ServiceIntegrationSuite) TestPlainMessages() {
	// Arrange.
	svc, err := msgproducer.New(msgproducer.NewOptions(
		msgproducer.NewKafkaWriter(s.KafkaBrokers(), s.messagesTopic, 1),
	))
	s.Require().NoError(err)
	defer func() { s.Require().NoError(svc.Close()) }()

	// Action.
	for _, m := range msgs {
		err := svc.ProduceMessage(s.Ctx, m)
		s.Require().NoError(err, "msg = ", m)
	}

	// Assert.
	producedMsgs := s.consumeMessages(len(msgs))

	producedMsgsByKey := groupByKey(producedMsgs)
	s.Len(producedMsgsByKey, chatsNumber)
	s.Require().Len(producedMsgsByKey[chat1], 3)
	s.Require().Len(producedMsgsByKey[chat2], 2)
	s.Require().Len(producedMsgsByKey[chat3], 1)

	s.Run("messages from one chat are in same partition", func() {
		for _, chatMsgs := range producedMsgsByKey {
			for _, m := range chatMsgs {
				s.Equal(chatMsgs[0].Partition, m.Partition, "msg = %s", m)
			}
		}
	})

	s.Run("message key is chat id", func() {
		for chatID, chatMsgs := range producedMsgsByKey {
			for _, m := range chatMsgs {
				s.Equal(chatID, string(m.Key), "msg = %s", m)
			}
		}
	})

	s.Run("assert messages values and order", func() {
		expectedChatMsgs := map[string][]string{
			chat1: {
				fmt.Sprintf(`{"id":%q,"chatId":%q,"body":"chat 1, message 1","fromClient":true}`, msg11, chat1),
				fmt.Sprintf(`{"id":%q,"chatId":%q,"body":"chat 1, message 2","fromClient":false}`, msg12, chat1),
				fmt.Sprintf(`{"id":%q,"chatId":%q,"body":"chat 1, message 3","fromClient":true}`, msg13, chat1),
			},
			chat2: {
				fmt.Sprintf(`{"id":%q,"chatId":%q,"body":"chat 2, message 1","fromClient":true}`, msg21, chat2),
				fmt.Sprintf(`{"id":%q,"chatId":%q,"body":"chat 2, message 2","fromClient":false}`, msg22, chat2),
			},
			chat3: {
				fmt.Sprintf(`{"id":%q,"chatId":%q,"body":"chat 3, message 1","fromClient":true}`, msg31, chat3),
			},
		}

		for chatID, chatMsgs := range producedMsgsByKey {
			s.Require().Len(chatMsgs, len(expectedChatMsgs[chatID]))
			for i, m := range chatMsgs {
				s.JSONEq(expectedChatMsgs[chatID][i], string(m.Value), "chat = %s, msg #%d", chatID, i)
			}
		}
	})
}

func (s *ServiceIntegrationSuite) TestEncryptedMessages() {
	// Arrange.
	svc, err := msgproducer.New(msgproducer.NewOptions(
		msgproducer.NewKafkaWriter(s.KafkaBrokers(), s.messagesTopic, 1),
		msgproducer.WithEncryptKey("68566D597133743677397A2443264629"),
		msgproducer.WithNonceFactory(func(size int) ([]byte, error) {
			return bytes.Repeat([]byte{'1'}, size), nil
		}),
	))
	s.Require().NoError(err)
	defer func() { s.Require().NoError(svc.Close()) }()

	// Action.
	for _, m := range msgs {
		err := svc.ProduceMessage(s.Ctx, m)
		s.Require().NoError(err, "msg = ", m)
	}

	// Assert.
	producedMsgs := s.consumeMessages(len(msgs))

	producedMsgsByKey := groupByKey(producedMsgs)
	s.Len(producedMsgsByKey, chatsNumber)
	s.Require().Len(producedMsgsByKey[chat1], 3)
	s.Require().Len(producedMsgsByKey[chat2], 2)
	s.Require().Len(producedMsgsByKey[chat3], 1)

	s.Run("messages from one chat are in same partition", func() {
		for _, chatMsgs := range producedMsgsByKey {
			for _, m := range chatMsgs {
				s.Equal(chatMsgs[0].Partition, m.Partition, "msg = %s", m)
			}
		}
	})

	s.Run("message key is chat id", func() {
		for chatID, chatMsgs := range producedMsgsByKey {
			for _, m := range chatMsgs {
				s.Equal(chatID, string(m.Key), "msg = %s", m)
			}
		}
	})

	s.Run("assert messages values and order", func() {
		expectedChatMsgs := map[string][]string{
			chat1: {
				"3131313131313131313131317234451261f7b6d088f83e40f4192f5e4390f73013204bb5d664f96485e5fa3647b6c5c45e6cdb8fa8c9b39a4806a3dddbd14950c1409a45584e7b615026525078aaae1f61fb513a5fb260a3abd4217f1e7d13ff6a84dfe5031e815b6f3584d1c2bd37f9c70ad4ee049e06b147861b71c74aa5029b328aa5e7adf6561e9853694a329012b35cb6a0edde299c9e9aaf1ad3dba5d88d5db228e35f",   //nolint:lll
				"3131313131313131313131317234451261f7b6d786a13a47f14b785e4390f73113204bb5d664f93781e9fa3647b6c5c45e6cdb8fa8c9b39a4806a3dddbd14950c1409a45584e7b615026525078aaae1f61fb513a5fb260a3abd4217f1e7d13ff6a84dfe5031e815b6f3584d1c2bd37f9c70ad4ee049e06b147861b71c74aa5029b318aa5e7adf6561e9853694a329012b34ea5b9fbc6a77e8a986ccbd361fe78d806a2b5c8366a", //nolint:lll
				"3131313131313131313131317234451261f7b6d586ff6e15a14b7a5e4390f73113204bb5d664a33987eafa3647b6c5c45e6cdb8fa8c9b39a4806a3dddbd14950c1409a45584e7b615026525078aaae1f61fb513a5fb260a3abd4217f1e7d13ff6a84dfe5031e815b6f3584d1c2bd37f9c70ad4ee049e06b147861b71c74aa5029b308aa5e7adf6561e9853694a329012b35cb6a0edde70173140eb48ecaef2207c27f1ac3553",   //nolint:lll
			},
			chat2: {
				"3131313131313131313131317234451261f7b68687a83a15f54a7a5e4391a16713204bb5d664a260d2befa3647b6c5c45e6cdb8fa8c9b39a4806a3dddbd14950c1409a450d147c655377505078aaae1f61fb513a5fb238f3a8d0217f1e7d13ff6a84dfe5031e815b6f3584d1c2bd37f9c70ad4ee049e05b147861b71c74aa5029b328aa5e7adf6561e9853694a329012b35cb6a0edde4c98bcdfee9a02f37226affd75f50b84",   //nolint:lll
				"3131313131313131313131317234451261f7b6d3d0fb6e10a94f7c5e4391a13613204bb5d664fa31d2e5fa3647b6c5c45e6cdb8fa8c9b39a4806a3dddbd14950c1409a450d147c655377505078aaae1f61fb513a5fb238f3a8d0217f1e7d13ff6a84dfe5031e815b6f3584d1c2bd37f9c70ad4ee049e05b147861b71c74aa5029b318aa5e7adf6561e9853694a329012b34ea5b9fbc6a7ffc511d84310df726032cda90e98a6ad", //nolint:lll
			},
			chat3: {
				"3131313131313131313131317234451261f7b68585f86b12f3482f5e4391a16713204bb5d664f93181bafa3647b6c5c45e6cdb8fa8c9b39a4806a3dddbd14950c1409a445715796c007d075078aaae1f61fb513a5fb261abae84217f1e7d13ff6a84dfe5031e815b6f3584d1c2bd37f9c70ad4ee049e04b147861b71c74aa5029b328aa5e7adf6561e9853694a329012b35cb6a0eddea5dddf50c8cf2db87902fb33c56471a4", //nolint:lll
			},
		}

		for chatID, chatMsgs := range producedMsgsByKey {
			s.Require().Len(chatMsgs, len(expectedChatMsgs[chatID]))
			for i, m := range chatMsgs {
				s.Equal(expectedChatMsgs[chatID][i], hex.EncodeToString(m.Value),
					"chat = %s, msg #%d", chatID, i,
				)
			}
		}
	})
}

func (s *ServiceIntegrationSuite) consumeMessages(n int) []kafka.Message {
	s.T().Helper()

	result := make([]kafka.Message, 0, n)
	for i := 0; i < n; i++ {
		func() {
			ctx, cancel := context.WithTimeout(s.Ctx, 3*time.Second)
			defer cancel()

			msg, err := s.messagesConsumer.ReadMessage(ctx)
			s.Require().NoError(err, "i = ", i)
			result = append(result, msg)
		}()
	}
	return result
}

func groupByKey(msgs []kafka.Message) map[string][]kafka.Message {
	result := make(map[string][]kafka.Message)
	for _, m := range msgs {
		result[string(m.Key)] = append(result[string(m.Key)], m)
	}
	return result
}

// Fixtures.

const chatsNumber = 3

var (
	chat1 = "86ba45bc-84fd-11ed-9104-461e464ebed8"
	chat2 = "8c8f063a-84fd-11ed-aa30-461e464ebed8"
	chat3 = "999c9e96-84fd-11ed-895d-461e464ebed8"

	msg11 = "79a3cdc6-84fd-11ed-bea9-461e464ebed8"
	msg12 = "0787da1a-84fe-11ed-b6e5-461e464ebed8"
	msg13 = "27fc611c-84fe-11ed-88c6-461e464ebed8"
	msg21 = "a6176e0c-8503-11ed-9a6b-461e464ebed8"
	msg22 = "4abc395e-850b-11ed-a069-461e464ebed8"
	msg31 = "b4af1c26-8503-11ed-b0ef-461e464ebed8"
)

var msgs = []msgproducer.Message{
	{
		ID:         types.MustParse[types.MessageID](msg11),
		ChatID:     types.MustParse[types.ChatID](chat1),
		Body:       "chat 1, message 1",
		FromClient: true,
	},
	{
		ID:         types.MustParse[types.MessageID](msg12),
		ChatID:     types.MustParse[types.ChatID](chat1),
		Body:       "chat 1, message 2",
		FromClient: false,
	},
	{
		ID:         types.MustParse[types.MessageID](msg13),
		ChatID:     types.MustParse[types.ChatID](chat1),
		Body:       "chat 1, message 3",
		FromClient: true,
	},
	// ---
	{
		ID:         types.MustParse[types.MessageID](msg21),
		ChatID:     types.MustParse[types.ChatID](chat2),
		Body:       "chat 2, message 1",
		FromClient: true,
	},
	{
		ID:         types.MustParse[types.MessageID](msg22),
		ChatID:     types.MustParse[types.ChatID](chat2),
		Body:       "chat 2, message 2",
		FromClient: false,
	},
	// ---
	{
		ID:         types.MustParse[types.MessageID](msg31),
		ChatID:     types.MustParse[types.ChatID](chat3),
		Body:       "chat 3, message 1",
		FromClient: true,
	},
}
