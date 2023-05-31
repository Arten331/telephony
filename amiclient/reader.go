package amiclient

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"strings"
	"time"
)

func (c *Client) runReader(ctx context.Context) {
	defer func() {
		_ = c.conn.Close()
	}()

	reader := bufio.NewReader(c.conn)

	for {
		select {
		case <-c.stopReader:
			return
		case <-ctx.Done():
			return
		default:
			if c.settings.ReadTimeOut != 0 {
				_ = c.conn.SetReadDeadline(time.Now().Add(c.settings.ReadTimeOut))
			}

			msg, err := ReadMessage(reader)

			c.metrics.StoreReceivedMessage()

			if err != nil {
				c.errChan <- err

				return
			}

			if len(msg) == 0 {
				continue
			}

			c.msgChan <- msg
		}
	}
}

func ReadMessage(r *bufio.Reader) (Message, error) {
	var (
		buf      bytes.Buffer
		isPrefix bool
		err      error
		line     []byte
	)

	for {
		line, isPrefix, err = r.ReadLine()
		if err == io.EOF /*|| isPrefix */ {
			continue
		}

		if err != nil {
			if strings.Contains(err.Error(), "i/o timeout") {
				err = nil
			} else {
				return nil, err
			}
		}

		if line == nil || len(line) <= 1 {
			break
		}

		buf.Write(line)

		if !isPrefix {
			buf.Write([]byte{'\n'})
		}

		//if r.Buffered() == 0 {
		//	break
		//}
	}

	msg := ParseMessage(buf)

	return msg, err
}
