package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/timurkash/pow/internal/pkg/config"
	"github.com/timurkash/pow/internal/pkg/pow"
	"github.com/timurkash/pow/internal/pkg/protocol"
	"io"
	"net"
	"strings"
	"time"
)

// Run - main function, launches client to connect and work with server on address
func Run(ctx context.Context, address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	fmt.Println("connected to", address)
	defer conn.Close()

	// client will send new request every 5 seconds endlessly
	for {
		message, err := HandleConnection(ctx, conn, conn)
		if err != nil {
			return err
		}
		fmt.Println("quote result:")
		fmt.Println("======================")
		fmt.Println(strings.ReplaceAll(message, "世界", "\n"))
		fmt.Println("======================")
		fmt.Println()
		time.Sleep(10 * time.Second)
	}
}

func HandleConnection(ctx context.Context, readerConn io.Reader, writerConn io.Writer) (string, error) {
	reader := bufio.NewReader(readerConn)
	err := sendMsg(protocol.Message{
		Header: protocol.RequestChallenge,
	}, writerConn)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}

	msgStr, err := readConnMsg(reader)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}
	msg, err := protocol.ParseMessage(msgStr)
	if err != nil {
		return "", fmt.Errorf("err parse msg: %w", err)
	}
	var hashCash pow.HashCashData
	err = json.Unmarshal([]byte(msg.Payload), &hashCash)
	if err != nil {
		return "", fmt.Errorf("err parse hashCash: %w", err)
	}
	fmt.Println("got hashCash:", hashCash)

	conf := ctx.Value("config").(*config.Config)
	hashCash, err = hashCash.ComputeHashCash(conf.HashCashMaxIterations)
	if err != nil {
		return "", fmt.Errorf("err compute hashCash: %w", err)
	}
	fmt.Println("hashCash computed:", hashCash)
	byteData, err := json.Marshal(hashCash)
	if err != nil {
		return "", fmt.Errorf("err marshal hashCash: %w", err)
	}

	err = sendMsg(protocol.Message{
		Header:  protocol.RequestResource,
		Payload: string(byteData),
	}, writerConn)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}
	fmt.Println("challenge sent to server")

	msgStr, err = readConnMsg(reader)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}
	msg, err = protocol.ParseMessage(msgStr)
	if err != nil {
		return "", fmt.Errorf("err parse msg: %w", err)
	}
	return msg.Payload, nil
}

// readConnMsg - read string message from connection
func readConnMsg(reader *bufio.Reader) (string, error) {
	return reader.ReadString('\n')
}

// sendMsg - send protocol message to connection
func sendMsg(msg protocol.Message, conn io.Writer) error {
	msgStr := fmt.Sprintf("%s\n", msg.Stringify())
	_, err := conn.Write([]byte(msgStr))
	return err
}
