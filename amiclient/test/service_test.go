package test_test

import (
	"bufio"
	"bytes"
	"context"
	"embed"
	"fmt"
	"io"
	"net"
	"strconv"

	_ "embed"

	"github.com/Arten331/telephony/amiclient"

	"github.com/Arten331/observability/logger"
)

//go:embed data/*
var fs embed.FS

func StartTestTCPServer(ctx context.Context, port int, enableLog bool) net.Listener {
	c, _ := net.Listen("tcp", net.JoinHostPort("", strconv.Itoa(port)))

	logger.S().Info("Started test TCP server %s", c.Addr())

	go func() {
		for {
			select {
			case <-ctx.Done():
				err := c.Close()
				if err != nil {
					logger.S().Infof("Unable stop TCP server: %s", err.Error())
				}

				logger.S().Infof("Stopped test TCP server %s", c.Addr())

				return
			default:
				conn, err := c.Accept()
				if err != nil || conn == nil {
					continue
				}

				go handleConnection(conn, enableLog)
			}
		}
	}()

	return c
}

func handleConnection(conn net.Conn, enableLog bool) {
	var (
		message amiclient.Message
		err     error
	)

	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	for {
		message, err = readMessage(rw.Reader)
		if err != nil {
			_, _ = rw.WriteString("failed to read input")

			_ = rw.Flush()

			return
		}

		answer := getAnswer(message)

		// For bench section
		_, testCnt := message["Count"]
		if message["Action"] == "GiveMeTest" && testCnt {
			_, ok := message["Count"]
			if ok {
				cnt, _ := strconv.Atoi(message["Count"])
				cnt /= 4
				cnt++

				for {
					cnt--

					_, _ = rw.WriteString(fmt.Sprintf("\n%s\n", answer))
					_ = rw.Flush()

					if cnt == 0 {
						return
					}
				}
			}
		}

		if enableLog {
			logger.S().Infof("received: \n%s", serializeMessage(message))
			logger.S().Infof("answer: \n%s", answer)
		}

		_, _ = rw.WriteString(fmt.Sprintf("\n%s\n", answer))
		_ = rw.Flush()
	}
}

func getAnswer(message amiclient.Message) []byte {
	var answer bytes.Buffer

	_, ok := message["Action"]

	if !ok {
		_, _ = answer.WriteString("Response: Error\nMessage: Missing action in request")

		return answer.Bytes()
	}

	switch message["Action"] {
	case "Login":
		if message["Username"] == "BadGuy" || message["Secret"] == "BadGuy" {
			_, _ = answer.WriteString("Response: Error\nMessage: Authentication failed")
		}

		_, _ = answer.WriteString("Response: Success\nMessage: Authentication accepted")
	case "GiveMeTest":
		file, err := fs.Open("data/test_messages.txt")
		if err != nil {
			panic(err)
		}

		msg, _ := io.ReadAll(file)

		_, _ = answer.Write(msg)
	default:
		_, _ = answer.WriteString("Response: Error\nMessage: Permission denied")
	}

	return answer.Bytes()
}

func serializeMessage(m map[string]string) []byte {
	var command bytes.Buffer

	for key := range m {
		command.WriteString(key)
		command.Write([]byte{':', ' '})
		command.WriteString(m[key])
		command.Write([]byte{'\n'})
	}

	return command.Bytes()
}

func readMessage(r *bufio.Reader) (amiclient.Message, error) {
	var (
		buf      bytes.Buffer
		isPrefix bool
		err      error
		line     []byte
	)

	for {
		line, isPrefix, err = r.ReadLine()
		if line == nil || len(line) <= 1 {
			break
		}

		buf.Write(line)

		if !isPrefix {
			buf.Write([]byte{'\n'})
		}
	}

	msg := amiclient.ParseMessage(buf)

	return msg, err
}
