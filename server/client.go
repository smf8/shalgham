package server

import (
	"bufio"
	"encoding/json"
	"net"

	"github.com/sirupsen/logrus"
	"github.com/smf8/shalgham/common"
)

type Client struct {
	Conn           net.Conn
	OutputBuffer   *bufio.Writer
	ReadBuffer     *bufio.Reader
	SendQueue      chan common.Msg
	RecvQueue      chan common.Msg
	DisconnectChan chan *Client
}

func (c *Client) ReadMessage() {
	for {
		msg := common.Msg{}

		decoder := json.NewDecoder(c.ReadBuffer)
		if err := decoder.Decode(&msg); err != nil {
			logrus.Errorf("failed to read and decode message: %s", err)
			c.Conn.Close()
			c.DisconnectChan <- c

			return
		}

		c.RecvQueue <- msg
	}
}

//SendMessage is for sending messages from channel to socket output buffer.
//Run it only 1 time in a separate goroutine.
func (c *Client) SendMessage() {
	for {
		m := <-c.SendQueue
		//TODO: implement message segmentation

		if _, err := c.OutputBuffer.Write(m.ToJSON()); err != nil {
			logrus.Errorf("failed to send message: %s\n", err)
		}

		if err := c.OutputBuffer.Flush(); err != nil {
			logrus.Errorf("failed to flush message: %s\n", err)
			c.Conn.Close()
			c.DisconnectChan <- c

			return
		}
	}
}
