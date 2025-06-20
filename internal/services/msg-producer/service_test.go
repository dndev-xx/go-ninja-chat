package msgproducer_test

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	msgproducer "github.com/dndev-xx/go-ninja-chat/internal/services/msg-producer"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

func TestService_ProduceMessage(t *testing.T) {
	const messagesCount = 10

	cases := []struct {
		name string
		key  string
	}{
		{
			name: "plain",
			key:  "",
		},
		{
			name: "encrypted",
			key:  "24432646294A404E635266546A576E5A",
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange.
			writer := new(kafkaWriterMock)
			s, err := msgproducer.New(msgproducer.NewOptions(writer, msgproducer.WithEncryptKey(tt.key)))
			require.NoError(t, err)
			defer func() {
				require.NoError(t, s.Close())
				assert.True(t, writer.closed)
			}()

			msgs := make([]msgproducer.Message, 0, messagesCount)
			for i := 0; i < messagesCount; i++ {
				msgs = append(msgs, msgproducer.Message{
					ID:         types.NewMessageID(),
					ChatID:     types.NewChatID(),
					Body:       fmt.Sprintf("Message %d", i),
					FromClient: i%3 != 0,
				})
			}

			// Action.
			for i := 0; i < messagesCount; i++ {
				err = s.ProduceMessage(context.Background(), msgs[i])
				require.NoError(t, err, "i=%d", i)
			}
			require.Len(t, writer.msgs, messagesCount)

			// Assert.
			produced := make([]msgproducer.Message, 0, messagesCount)
			for _, m := range writer.msgs {
				data := m.Value
				if tt.key != "" {
					data = requireMsgDecrypt(t, tt.key, data)
				}

				msg := requireMsgUnmarshal(t, data)
				assert.Equal(t, []byte(msg.ChatID.String()), m.Key)

				produced = append(produced, msg)
			}
			assert.Equal(t, msgs, produced)
		})
	}
}

func requireMsgDecrypt(t *testing.T, keyStr string, data []byte) []byte {
	t.Helper()

	key, err := hex.DecodeString(keyStr)
	require.NoError(t, err)

	blockCipher, err := aes.NewCipher(key)
	require.NoError(t, err)

	aead, err := cipher.NewGCM(blockCipher)
	require.NoError(t, err)

	raw, ns := data, aead.NonceSize()
	nonce, encrypted := raw[:ns], raw[ns:]

	decrypted, err := aead.Open(nil, nonce, encrypted, nil)
	require.NoError(t, err)

	return decrypted
}

func requireMsgUnmarshal(t *testing.T, data []byte) msgproducer.Message {
	t.Helper()

	var rcvMsg struct {
		ID         string `json:"id"`
		ChatID     string `json:"chatId"`
		Body       string `json:"body"`
		FromClient bool   `json:"fromClient"`
	}
	require.NoError(t, json.Unmarshal(data, &rcvMsg))

	return msgproducer.Message{
		ID:         types.MustParse[types.MessageID](rcvMsg.ID),
		ChatID:     types.MustParse[types.ChatID](rcvMsg.ChatID),
		Body:       rcvMsg.Body,
		FromClient: rcvMsg.FromClient,
	}
}

var _ msgproducer.KafkaWriter = (*kafkaWriterMock)(nil)

type kafkaWriterMock struct {
	msgs   []kafka.Message
	closed bool
}

func (m *kafkaWriterMock) Close() error {
	m.closed = true
	return nil
}

func (m *kafkaWriterMock) WriteMessages(_ context.Context, msgs ...kafka.Message) error {
	m.msgs = append(m.msgs, msgs...)
	return nil
}
