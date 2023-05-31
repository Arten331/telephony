package test_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"telephony/amiclient"

	"github.com/Arten331/observability/logger"
)

const testTCPPort = 40000

func init() {
	logger.MustSetupGlobal(
		logger.WithConfiguration(logger.CoreOptions{
			OutputPath: "/dev/null",
			Level:      logger.KeyLevelDebug,
			Encoding:   logger.EncodingConsole,
		}),
	)
}

func TestClient(t *testing.T) {
	var err error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s := StartTestTCPServer(ctx, testTCPPort, true)
	defer func() { _ = s.Close() }()

	<-time.After(time.Second * 2)

	client := amiclient.New(&amiclient.Settings{
		Host:              "",
		Port:              testTCPPort,
		Username:          "test",
		Password:          "test",
		ConnectionTimeout: 30 * time.Second,
	})

	err = client.Connect(ctx, true)
	if err != nil {
		t.Fatalf("Unable connect to test tcp server, %s", err.Error())
	}

	t.Log("Connection successful")

	commandTestData := make(amiclient.Action)
	commandTestData["Action"] = "GiveMeTest"

	err = client.SendCommand(commandTestData)
	if err != nil {
		t.Fatalf(err.Error())
	}

	wg := sync.WaitGroup{}
	waitCh := make(chan struct{})

	wg.Add(4)

	ctx, cancel = context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	msgs := make([]amiclient.Message, 0, 4)

	go func(t *testing.T) {
		for {
			select {
			case err := <-client.ErrChan():
				t.Errorf("Error while read test messages, %s", err.Error())
			case msg := <-client.MsgChan():
				msgs = append(msgs, msg)
				wg.Done()
				logger.S().Infof("Receive message: %v", msg)
			}
		}
	}(t)

	go func() {
		wg.Wait()
		close(waitCh)
	}()

	fmt.Println(msgs)

	select {
	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			t.Errorf("4 test messages not come back, %s", ctx.Err())
			wg.Add(4)
		}
	case <-waitCh:
	}
}
