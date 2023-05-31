package amiclient

import (
	"bytes"
	"io"
)

//nolint:gochecknoglobals // bytes constraint
var (
	actionDelimiterS = []byte{':', ' '}
	lineTerm         = []byte{'\n'}
	actionDelimiterB = []byte(": ")
)

type (
	ActionKey []byte
	Action    map[string]string
	Message   map[string]string
)

func (a Action) Serialize() []byte {
	var command bytes.Buffer

	for key := range a {
		command.WriteString(key)
		command.Write(actionDelimiterS)
		command.WriteString(a[key])
		command.Write(lineTerm)
	}

	command.Write(lineTerm)

	return command.Bytes()
}

func ParseAction(buf bytes.Buffer) Action {
	event := make(Action)

	for {
		line, err := buf.ReadBytes('\n')
		line = bytes.TrimSuffix(line, lineTerm)

		if (err != io.EOF && err != nil) || len(line) == 0 {
			break
		}

		vMap := bytes.Split(line, actionDelimiterB)

		event[string(vMap[0])] = string(vMap[1])
	}

	return event
}

func ParseMessage(buf bytes.Buffer) Message {
	event := make(Message)

	for {
		line, err := buf.ReadBytes('\n')
		line = bytes.TrimSuffix(line, lineTerm)

		if (err != io.EOF && err != nil) || len(line) == 0 {
			break
		}

		vMap := bytes.Split(line, actionDelimiterB)

		event[string(vMap[0])] = getMsgValue(vMap)
	}

	return event
}

func getMsgValue(vMap [][]byte) string {
	if len(vMap) == 1 {
		return ""
	}

	return string(vMap[1])
}
