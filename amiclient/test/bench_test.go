package test_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"telephony/amiclient"
)

// go test -test.bench=BenchmarkMessages -test.count=2 -test.benchtime=100000x -test.benchmem -test.cpu=1,2,4,6,12
func BenchmarkMessages(b *testing.B) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cnt := b.N

	s := StartTestTCPServer(ctx, testTCPPort, false)
	defer func() { _ = s.Close() }()

	commandTestData := make(amiclient.Action)
	commandTestData["Action"] = "GiveMeTest"
	commandTestData["Count"] = strconv.Itoa(cnt)

	client := amiclient.New(&amiclient.Settings{
		Host:              "",
		Port:              testTCPPort,
		Username:          "test",
		Password:          "test",
		ConnectionTimeout: 30 * time.Second,
		Disabled:          false,
		ReadTimeOut:       time.Second * 10,
	})

	err := client.Connect(ctx, true)
	if err != nil {
		b.Fatalf(err.Error())
	}

	err = client.SendCommand(commandTestData)
	if err != nil {
		b.Fatalf(err.Error())
	}

	wg := sync.WaitGroup{}
	waitCh := make(chan struct{})

	wg.Add(cnt)

	go func() {
		wg.Wait()
		close(waitCh)
	}()

	<-time.After(1 * time.Second)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		select {
		case err := <-client.ErrChan():
			b.Fatalf("Error while read test messages, %s", err.Error())
		case msg := <-client.MsgChan():
			valid := checkMessage(msg)
			if !valid {
				b.Fatalf("Wrong received message, %+v", msg)
			}

			wg.Done()
		}
	}

	<-waitCh
}

//nolint:gocognit
func checkMessage(msg amiclient.Message) bool {
	isOK := true

	_, isEvent := msg["Event"]

	if isEvent { //nolint:nestif
		switch msg["Event"] {
		case "FullyBooted":
			if msg["Privilege"] != "system,all" || msg["Status"] != "Fully Booted" {
				isOK = false

				break
			}
		case "PeerStatus":
			if msg["Address"] == "192.168.111.68:5060" {
				if msg["Privilege"] != "system,all" ||
					msg["ChannelType"] != "SIP" ||
					msg["Peer"] != "SIP/pbx_sbc2_test" ||
					msg["PeerStatus"] != "Registered" ||
					msg["Address"] != "192.168.111.68:5060" {
					isOK = false

					break
				}
			} else {
				if msg["Privilege"] != "system,all" ||
					msg["ChannelType"] != "SIP" ||
					msg["Peer"] != "SIP/sbc_incoming_test" ||
					msg["PeerStatus"] != "Registered" ||
					msg["Address"] != "192.168.111.4:5060" {
					isOK = false

					break
				}
			}
		}
	}

	_, isResponse := msg["Response"]
	if isResponse {
		if msg["Response"] != "Success" ||
			msg["AMIversion"] != "2.10.5" ||
			msg["AsteriskVersion"] != "13.31.0" ||
			msg["SystemName"] != "" ||
			msg["CoreMaxCalls"] != "0" ||
			msg["CoreMaxLoadAvg"] != "0.000000" ||
			msg["CoreRunUser"] != "" ||
			msg["CoreRunGroup"] != "" ||
			msg["CoreMaxFilehandles"] != "0" ||
			msg["CoreRealTimeEnabled"] != "Yes" ||
			msg["CoreCDRenabled"] != "Yes" ||
			msg["CoreHTTPenabled"] != "No" {
			isOK = false
		}
	}

	return isOK
}

/*Event: FullyBooted
Privilege: system,all
Status: Fully Booted

Event: PeerStatus
Privilege: system,all
ChannelType: SIP
Peer: SIP/pbx_sbc2_test
PeerStatus: Registered
Address: 192.168.111.68:5060

Event: PeerStatus
Privilege: system,all
ChannelType: SIP
Peer: SIP/sbc_incoming_test
PeerStatus: Registered
Address: 192.168.111.4:5060

Response: Success
AMIversion: 2.10.5
AsteriskVersion: 13.31.0
SystemName:
CoreMaxCalls: 0
CoreMaxLoadAvg: 0.000000
CoreRunUser:
CoreRunGroup:
CoreMaxFilehandles: 0
CoreRealTimeEnabled: Yes
CoreCDRenabled: Yes
CoreHTTPenabled: No*/
