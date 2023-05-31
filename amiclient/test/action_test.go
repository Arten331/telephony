package test_test

import (
	"bytes"
	"reflect"
	"testing"

	"telephony/amiclient"
)

type ParseActionTC struct {
	name           string
	input          []byte
	expectedResult amiclient.Message
}

type SerialiseActionTC struct {
	name           string
	input          amiclient.Action
	expectedResult amiclient.Action
}

func TestAction_Serialize(t *testing.T) {
	testCases := []SerialiseActionTC{
		{
			name: "login response success",
			input: amiclient.Action{
				"Response": "Success",
				"Message":  "Authentication accepted",
			},
			expectedResult: amiclient.Action{
				"Response": "Success",
				"Message":  "Authentication accepted",
			},
		},
		{
			name: "hangup normal clearing",
			input: amiclient.Action{
				"Message":           "Hangup",
				"Privilege":         "call,all",
				"Channel":           "Local/979144181775@phonenumber-checker-00000147;2",
				"ChannelState":      "4",
				"ChannelStateDesc":  "Ring",
				"CallerIDNum":       "<unknown>",
				"CallerIDName":      "<unknown>",
				"ConnectedLineNum":  "<unknown>",
				"ConnectedLineName": "<unknown>",
				"Language":          "en",
				"AccountCode":       "phonenumber-checker",
				"Context":           "phonenumber-checker",
				"Exten":             "h",
				"Priority":          "1",
				"Uniqueid":          "1651218111.2444",
				"Linkedid":          "1651218111.2443",
				"cause":             "16",
				"cause-txt":         "Normal Clearing",
			},
			expectedResult: amiclient.Action{
				"Message":           "Hangup",
				"Privilege":         "call,all",
				"Channel":           "Local/979144181775@phonenumber-checker-00000147;2",
				"ChannelState":      "4",
				"ChannelStateDesc":  "Ring",
				"CallerIDNum":       "<unknown>",
				"CallerIDName":      "<unknown>",
				"ConnectedLineNum":  "<unknown>",
				"ConnectedLineName": "<unknown>",
				"Language":          "en",
				"AccountCode":       "phonenumber-checker",
				"Context":           "phonenumber-checker",
				"Exten":             "h",
				"Priority":          "1",
				"Uniqueid":          "1651218111.2444",
				"Linkedid":          "1651218111.2443",
				"cause":             "16",
				"cause-txt":         "Normal Clearing",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			resBytes := tc.input.Serialize()

			buf.Write(resBytes)

			res := amiclient.ParseAction(buf)

			if !reflect.DeepEqual(res, tc.expectedResult) {
				t.Error("Wrong expected result", "result:", res, "expected", tc.expectedResult)
			}
		})
	}
}

func TestParseEvent(t *testing.T) {
	testCases := []ParseActionTC{
		{
			name:  "login response success",
			input: []byte("Response: Success\nMessage: Authentication accepted"),
			expectedResult: amiclient.Message{
				"Response": "Success",
				"Message":  "Authentication accepted",
			},
		},
		{
			name:  "hangup normal clearing",
			input: []byte("Message: Hangup\nPrivilege: call,all\nChannel: Local/979144181775@phonenumber-checker-00000147;2\nChannelState: 4\nChannelStateDesc: Ring\nCallerIDNum: <unknown>\nCallerIDName: <unknown>\nConnectedLineNum: <unknown>\nConnectedLineName: <unknown>\nLanguage: en\nAccountCode: phonenumber-checker\nContext: phonenumber-checker\nExten: h\nPriority: 1\nUniqueid: 1651218111.2444\nLinkedid: 1651218111.2443\ncause: 16\ncause-txt: Normal Clearing"),
			expectedResult: amiclient.Message{
				"Message":           "Hangup",
				"Privilege":         "call,all",
				"Channel":           "Local/979144181775@phonenumber-checker-00000147;2",
				"ChannelState":      "4",
				"ChannelStateDesc":  "Ring",
				"CallerIDNum":       "<unknown>",
				"CallerIDName":      "<unknown>",
				"ConnectedLineNum":  "<unknown>",
				"ConnectedLineName": "<unknown>",
				"Language":          "en",
				"AccountCode":       "phonenumber-checker",
				"Context":           "phonenumber-checker",
				"Exten":             "h",
				"Priority":          "1",
				"Uniqueid":          "1651218111.2444",
				"Linkedid":          "1651218111.2443",
				"cause":             "16",
				"cause-txt":         "Normal Clearing",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var buf bytes.Buffer

			buf.Write(tc.input)

			res := amiclient.ParseMessage(buf)

			if !reflect.DeepEqual(res, tc.expectedResult) {
				t.Error("Wrong expected result", "result:", res, "expected", tc.expectedResult)
			}
		})
	}
}
