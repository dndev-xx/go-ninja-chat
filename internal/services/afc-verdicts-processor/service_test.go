package afcverdictsprocessor_test

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"io"
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/golang/mock/gomock"
	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/suite"

	afcverdictsprocessor "github.com/dndev-xx/go-ninja-chat/internal/services/afc-verdicts-processor"
	afcverdictsprocessormocks "github.com/dndev-xx/go-ninja-chat/internal/services/afc-verdicts-processor/mocks"
	clientmessageblockedjob "github.com/dndev-xx/go-ninja-chat/internal/services/outbox/jobs/client-message-blocked"
	clientmessagesentjob "github.com/dndev-xx/go-ninja-chat/internal/services/outbox/jobs/client-message-sent"
	"github.com/dndev-xx/go-ninja-chat/internal/testingh"
	"github.com/dndev-xx/go-ninja-chat/internal/types"
)

const (
	backoffInitialInterval = 50 * time.Millisecond
	backoffMaxElapsedTime  = 500 * time.Millisecond
)

type ServiceSuite struct {
	testingh.ContextSuite

	SignPrivateKey string
	SignPubKey     string

	ctrl        *gomock.Controller
	outboxSvc   *afcverdictsprocessormocks.MockoutboxService
	msgRepo     *afcverdictsprocessormocks.MockmessagesRepository
	transactor  *afcverdictsprocessormocks.Mocktransactor
	consumer    *afcverdictsprocessormocks.MockKafkaReader
	dlqProducer *afcverdictsprocessormocks.MockKafkaDLQWriter

	privateKey *rsa.PrivateKey
	svc        *afcverdictsprocessor.Service
}

func TestServiceSuite_NoSignature(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(ServiceSuite))
}

func TestServiceSuite_SignedVerdicts(t *testing.T) {
	t.Parallel()
	suite.Run(t, &ServiceSuite{
		SignPrivateKey: privateKey,
		SignPubKey:     publicKey,
	})
}

func (s *ServiceSuite) SetupTest() {
	s.ContextSuite.SetupTest()

	s.ctrl = gomock.NewController(s.T())
	s.outboxSvc = afcverdictsprocessormocks.NewMockoutboxService(s.ctrl)
	s.msgRepo = afcverdictsprocessormocks.NewMockmessagesRepository(s.ctrl)

	s.transactor = afcverdictsprocessormocks.NewMocktransactor(s.ctrl)
	s.transactor.EXPECT().RunInTx(gomock.Any(), gomock.Any()).DoAndReturn(
		func(ctx context.Context, f func(ctx context.Context) error) error {
			return f(ctx)
		}).AnyTimes()

	s.consumer = afcverdictsprocessormocks.NewMockKafkaReader(s.ctrl)
	s.dlqProducer = afcverdictsprocessormocks.NewMockKafkaDLQWriter(s.ctrl)

	if k := s.SignPrivateKey; k != "" {
		var err error
		s.privateKey, err = jwt.ParseRSAPrivateKeyFromPEM([]byte(k))
		s.Require().NoError(err)
	}

	var err error
	s.svc, err = afcverdictsprocessor.New(afcverdictsprocessor.NewOptions(
		[]string{"test:9092"},
		1,
		"afcverdictsprocessor_test.ServiceSuite",
		"afc.unit-test.verdicts",
		func(brokers []string, groupID string, topic string) afcverdictsprocessor.KafkaReader {
			s.Equal([]string{"test:9092"}, brokers)
			s.Equal("afcverdictsprocessor_test.ServiceSuite", groupID)
			s.Equal("afc.unit-test.verdicts", topic)
			return s.consumer
		},
		s.dlqProducer,
		s.transactor,
		s.msgRepo,
		s.outboxSvc,
		afcverdictsprocessor.WithVerdictsSignKey(s.SignPubKey),
		afcverdictsprocessor.WithBackoffInitialInterval(backoffInitialInterval),
		afcverdictsprocessor.WithBackoffMaxElapsedTime(backoffMaxElapsedTime),
	))
	s.Require().NoError(err)

	// Always.
	s.consumer.EXPECT().Close().Return(nil)
	s.dlqProducer.EXPECT().Close().Return(nil)
}

func (s *ServiceSuite) TearDownTest() {
	s.ContextSuite.TearDownTest()
	s.ctrl.Finish()
}

func (s *ServiceSuite) TestNotRetriableError_InvalidEncoding() {
	// Arrange.
	v := []byte(`{
  "chatId": "2d1bb2b4-1e11-11ed-9c9f-461e464ebed9",
  "messageId": "b611c338-1e11-11ed-b5ce-461e464ebed9",
  "status": "ok"`)
	msg := kafka.Message{Value: v}
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil)
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(kafka.Message{}, io.EOF).MaxTimes(1)
	s.consumer.EXPECT().CommitMessages(gomock.Any(), msg)
	s.dlqProducer.EXPECT().WriteMessages(gomock.Any(), kafkaMsgValueMatcher{v})

	// Action & assert.
	s.runProcessorFor(100 * time.Millisecond)
}

func (s *ServiceSuite) TestNotRetriableError_InvalidVerdict() {
	// Arrange.
	v := verdict{
		ChatID:    "2d1bb2b4-1e11-11ed-9c9f-461e464ebed9",
		MessageID: "b611c338-1e11-11ed-b5ce-461e464ebed9",
		Status:    "", // No required field.
	}
	data := []byte(s.encode(v))

	msg := kafka.Message{Value: data}
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil)
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(kafka.Message{}, io.EOF).MaxTimes(1)
	s.consumer.EXPECT().CommitMessages(gomock.Any(), msg)
	s.dlqProducer.EXPECT().WriteMessages(gomock.Any(), kafkaMsgValueMatcher{data})

	// Action & assert.
	s.runProcessorFor(100 * time.Millisecond)
}

func (s *ServiceSuite) TestOperationRetriedWithSuccess() {
	// Arrange.
	msgID := types.NewMessageID()
	v := verdict{
		ChatID:    "2d1bb2b4-1e11-11ed-9c9f-461e464ebed9",
		MessageID: msgID.String(),
		Status:    "ok",
	}
	data := []byte(s.encode(v))

	msg := kafka.Message{Value: data}
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil)
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(kafka.Message{}, io.EOF).MaxTimes(1)
	s.msgRepo.EXPECT().MarkAsVisibleForManager(gomock.Any(), msgID).Return(context.Canceled)
	s.msgRepo.EXPECT().MarkAsVisibleForManager(gomock.Any(), msgID).Return(context.Canceled)
	s.msgRepo.EXPECT().MarkAsVisibleForManager(gomock.Any(), msgID).Return(nil)
	s.outboxSvc.EXPECT().Put(gomock.Any(), clientmessagesentjob.Name, gomock.Any(), gomock.Any())
	s.consumer.EXPECT().CommitMessages(gomock.Any(), msg)

	// Action & assert.
	s.runProcessorFor(backoffMaxElapsedTime)
}

func (s *ServiceSuite) TestOperationBackoffExceeded() {
	// Arrange.
	msgID := types.NewMessageID()
	v := verdict{
		ChatID:    "2d1bb2b4-1e11-11ed-9c9f-461e464ebed9",
		MessageID: msgID.String(),
		Status:    "ok",
	}
	data := []byte(s.encode(v))

	msg := kafka.Message{Value: data}
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil)
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(kafka.Message{}, io.EOF).MaxTimes(1)
	s.msgRepo.EXPECT().MarkAsVisibleForManager(gomock.Any(), msgID).Return(context.Canceled).AnyTimes()
	s.consumer.EXPECT().CommitMessages(gomock.Any(), msg)
	s.dlqProducer.EXPECT().WriteMessages(gomock.Any(), kafkaMsgValueMatcher{data})

	// Action & assert.
	s.runProcessorFor(2 * backoffMaxElapsedTime)
}

func (s *ServiceSuite) TestProcessMessagesWithoutErrors() {
	// Arrange.
	const n = 10
	verdicts := make([]verdict, n)
	for i := 0; i < n; i++ {
		s := "ok"
		if i%2 == 0 {
			s = "suspicious"
		}
		verdicts[i] = verdict{
			ChatID:    types.NewChatID().String(),
			MessageID: types.NewMessageID().String(),
			Status:    s,
		}
	}

	for _, v := range verdicts {
		data := []byte(s.encode(v))

		msg := kafka.Message{Value: data}
		s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(msg, nil)
		if v.Status == "ok" {
			s.msgRepo.EXPECT().MarkAsVisibleForManager(gomock.Any(), types.MustParse[types.MessageID](v.MessageID)).Return(nil)
			s.outboxSvc.EXPECT().Put(gomock.Any(), clientmessagesentjob.Name, gomock.Any(), gomock.Any())
		} else {
			s.msgRepo.EXPECT().BlockMessage(gomock.Any(), types.MustParse[types.MessageID](v.MessageID))
			s.outboxSvc.EXPECT().Put(gomock.Any(), clientmessageblockedjob.Name, gomock.Any(), gomock.Any())
		}
		s.consumer.EXPECT().CommitMessages(gomock.Any(), msg)
	}
	s.consumer.EXPECT().FetchMessage(gomock.Any()).Return(kafka.Message{}, io.EOF).MaxTimes(1)

	// Action & assert.
	s.runProcessorFor(100 * time.Millisecond)
}

func (s *ServiceSuite) runProcessorFor(timeout time.Duration) {
	s.T().Helper()

	cancel, errCh := s.runProcessor()
	defer cancel()

	time.Sleep(timeout)
	cancel()
	s.NoError(<-errCh) // No error expected because of graceful shutdown via cancel ctx.
}

func (s *ServiceSuite) runProcessor() (context.CancelFunc, <-chan error) {
	s.T().Helper()

	ctx, cancel := context.WithCancel(s.Ctx)

	errCh := make(chan error)
	go func() { errCh <- s.svc.Run(ctx) }()

	return cancel, errCh
}

func (s *ServiceSuite) encode(v verdict) string {
	s.T().Helper()

	if s.privateKey != nil {
		result, err := jwt.NewWithClaims(jwt.SigningMethodRS256, v).SignedString(s.privateKey)
		s.Require().NoError(err)
		return result
	}

	result, err := json.Marshal(v)
	s.Require().NoError(err)
	return string(result)
}

type verdict struct {
	ChatID    string `json:"chatId"`
	MessageID string `json:"messageId"`
	Status    string `json:"status"`
}

func (v verdict) Valid() error { return nil }

var _ gomock.Matcher = kafkaMsgValueMatcher{}

type kafkaMsgValueMatcher struct {
	v []byte
}

func (km kafkaMsgValueMatcher) Matches(x any) bool {
	v, ok := x.(kafka.Message)
	if !ok {
		return false
	}
	return string(v.Value) == string(km.v)
}

func (km kafkaMsgValueMatcher) String() string {
	return string(km.v)
}

const (
	//nolint:gosec // not real key
	privateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAzf10MG/4YiDJ7M94FaVIL7sZ1z/fJKyTEm3fbJ4PgownCTv3
o3adWZYhNRdGwu/YOhKak2uSOxQUj15QwaCFjmlVCwKuaJeXbI5BNHct46Kzo0pj
aX5SiY1RhCPxiZtfGk/OaRXbiyU+yHNffY7TTvpAyLoFNTgn7OiiYPWPSCOmZ2zQ
L+1judRIyjP1Z1aIwenmD+LoyPZ+RQ9TrdZXKHi5DxgdV/f660smWHICiMBEAJ5a
kcu/uemvJbmBCJkJPoeQWz39x3t1OrMWE0G/Ocs09tUDUzdxXNes+RDLyx+b0J0O
zUIq/+m3rWJRpe+6ErWhGvj7mBHlm8aQirBerwIDAQABAoIBAH+zkjV5JP4In8ZM
tICOz9qvXozADyFYT3EMZoea0bi4FHc4EwTmwxPH69xTCs5NDLqrz+J2vNgdUcWz
zdLMJiAskslZpzA2Umy9IBVbkTpfIoin1EuRQa/+yTtnYRVTGjlgonEpWMrBk1OH
mvpm8f8zS7hlAleE8dOAQTJk6afpPTyNvj1baN9okdpNZ7+5pK9Ij+YcS+aOWLix
A+vsIm5b0W6eXXnJLZzXNr2N2O9P/iEIdOs0+cvP58rkNQ/d4flZ6AYnUCgHHei4
gZxCWZgMHXzdY/t/pFM+l3G1QJzlGM6L8sIcXToTYmE1xJEf4PCV0ILt/cHUFkqU
HGeExnECgYEA6SpWbNgClEkl2NG9qmNCsXNOVWKj7mRylPJsKVLMZq208SekmTMj
qMqeN0wnwhyhmM//nYYu6dxTxhJBvXxYJyMz0G27p/5HaZC7WAl5XYM6QPPmsky5
T4h11J2X8TLBIjrE3wQm5d/EL9i3UqZtffTfFRwv6r4r0dGJdYZ2WokCgYEA4inO
iKAVd1ERIIRL9pBb0fVCVJSlX86NR3VRvB16fyrCrFjx1IE8CaeEKcYmnE7Abe0J
/jSM5OHKSULGbl2DeofhT2FhgV+hM/wKd3G3dVaHiuMWO9lCwnwelbXq2Rt0hhN0
b1YVHkI8rWMC1RDvK8Z9cExLz9VH+VJq+41TwXcCgYEAhf4cmIQyR0EaDNXLp0VP
qGZZF9yN1IvJBSujWMQKTt94YjWj855d2bxG3ARZvMVzYDv363CXOTGyutr3CIuS
pTsnpZnKA6qvI01XPCqFomWtbnI7my9YNwp2nG7MSIIgVylqxba/G89SEST7hPW7
amz0Xk9Kgh4zVGqUEgPps/ECgYEAnhR6uCSs3Gldf0z5i637gBXd9yCvNvg45+mo
58PzC0/oIm9JGS/7twPP7SMDed3RwwQcKAKzOIhZzDtQV3Qlok+3vLRkYvlkw+E3
r6VchjelJf70W4DQmQAIoLw3GumF2PFgQTH6MNw7bTX3lNXxVre2lfe+RdbeJ/bj
sFBoaqECgYEAzK91/ea5p5Hlt5yCQLeLDKSf2ohmYspkqk0HTi8iGfji2Zo99Iir
1rFR0Oe3otPG40HXhKDi2YdhNy/D4ypaVDkr94awTBYY8zlmgAPhf/oZu48tkxCh
qIanZhvea4LFXIctQKhXDCH0qwTkR9adILLKgLBS/dTrzWG2JHBE1B8=
-----END RSA PRIVATE KEY-----`

	publicKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzf10MG/4YiDJ7M94FaVI
L7sZ1z/fJKyTEm3fbJ4PgownCTv3o3adWZYhNRdGwu/YOhKak2uSOxQUj15QwaCF
jmlVCwKuaJeXbI5BNHct46Kzo0pjaX5SiY1RhCPxiZtfGk/OaRXbiyU+yHNffY7T
TvpAyLoFNTgn7OiiYPWPSCOmZ2zQL+1judRIyjP1Z1aIwenmD+LoyPZ+RQ9TrdZX
KHi5DxgdV/f660smWHICiMBEAJ5akcu/uemvJbmBCJkJPoeQWz39x3t1OrMWE0G/
Ocs09tUDUzdxXNes+RDLyx+b0J0OzUIq/+m3rWJRpe+6ErWhGvj7mBHlm8aQirBe
rwIDAQAB
-----END PUBLIC KEY-----`
)
