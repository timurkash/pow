package server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/timurkash/pow/internal/pkg/config"
	"github.com/timurkash/pow/internal/pkg/pow"
	"github.com/timurkash/pow/internal/pkg/protocol"
	"log"
	"math/rand"
	"net"
	"os/exec"
	"strconv"
	"time"
)

var ErrQuit = errors.New("client requests to close connection")

// Clock  - interface for easier mock time.Now in tests
type Clock interface {
	Now() time.Time
}

type Cache interface {
	// Add - add rand value with expiration (in seconds) to cache
	Add(int, int64) error
	// Get - check existence of int key in cache
	Get(int) (bool, error)
	// Delete - delete key from cache
	Delete(int)
}

func Exec(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	return stdout, nil
}

// Run - main function, launches server to listen on given address and handle new connections
func Run(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	// Close the listener when the application closes.
	defer listener.Close()
	log.Println("listening", listener.Addr())
	for {
		// Listen for an incoming connection.
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accept connection: %w", err)
		}
		// Handle connections in a new goroutine.
		go handleConnection(ctx, conn)
	}
}

func handleConnection(ctx context.Context, conn net.Conn) {
	log.Println("new client:", conn.RemoteAddr())
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		req, err := reader.ReadString('\n')
		if err != nil {
			log.Println("err read connection:", err)
			return
		}
		msg, err := ProcessRequest(ctx, req, conn.RemoteAddr().String())
		if err != nil {
			log.Println("err process request:", err)
			return
		}
		if msg != nil {
			err := sendMsg(*msg, conn)
			if err != nil {
				log.Println("err send message:", err)
			}
		}
	}
}

func ProcessRequest(ctx context.Context,
	msgStr, clientInfo string,
) (*protocol.Message, error) {
	msg, err := protocol.ParseMessage(msgStr)
	if err != nil {
		return nil, err
	}
	switch msg.Header {
	case protocol.Quit:
		return nil, ErrQuit
	case protocol.RequestChallenge:
		log.Printf("client %s requests challenge\n", clientInfo)
		conf := ctx.Value("config").(*config.Config)
		clock := ctx.Value("clock").(Clock)
		cache := ctx.Value("cache").(Cache)
		date := clock.Now()

		randValue := rand.Intn(100000)
		err := cache.Add(randValue, conf.HashCashDuration)
		if err != nil {
			return nil, fmt.Errorf("err add rand to cache: %w", err)
		}

		hashCash := pow.HashCashData{
			Version:    1,
			ZerosCount: conf.HashCashZerosCount,
			Date:       date.Unix(),
			Resource:   clientInfo,
			Rand:       base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", randValue))),
			Counter:    0,
		}
		hashCashMarshaled, err := json.Marshal(hashCash)
		if err != nil {
			return nil, fmt.Errorf("err marshal hashCash: %v", err)
		}
		msg := protocol.Message{
			Header:  protocol.ResponseChallenge,
			Payload: string(hashCashMarshaled),
		}
		return &msg, nil
	case protocol.RequestResource:
		phrase, err := Exec("fortune")
		if err != nil {
			return nil, err
		}
		phrase = bytes.Trim(phrase, "\n")
		phrase = bytes.ReplaceAll(phrase, []byte("\n"), []byte("世界"))
		log.Println(string(phrase))
		log.Printf("client %s requests resource with payload %s\n", clientInfo, msg.Payload)
		// parse client's solution
		var hashCash pow.HashCashData
		if err = json.Unmarshal([]byte(msg.Payload), &hashCash); err != nil {
			return nil, fmt.Errorf("err unmarshal hashCash: %w", err)
		}
		if hashCash.Resource != clientInfo {
			return nil, fmt.Errorf("invalid hashCash resource")
		}
		conf := ctx.Value("config").(*config.Config)
		clock := ctx.Value("clock").(Clock)
		cache := ctx.Value("cache").(Cache)

		randValueBytes, err := base64.StdEncoding.DecodeString(hashCash.Rand)
		if err != nil {
			return nil, fmt.Errorf("err decode rand: %w", err)
		}
		randValue, err := strconv.Atoi(string(randValueBytes))
		if err != nil {
			return nil, fmt.Errorf("err decode rand: %w", err)
		}

		exists, err := cache.Get(randValue)
		if err != nil {
			return nil, fmt.Errorf("err get rand from cache: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("challenge expired or not sent")
		}

		// sent solution should not be outdated
		if clock.Now().Unix()-hashCash.Date > conf.HashCashDuration {
			return nil, fmt.Errorf("challenge expired")
		}
		//to prevent indefinite computing on server if client sent hashCash with 0 counter
		maxIter := hashCash.Counter
		if maxIter == 0 {
			maxIter = 1
		}
		_, err = hashCash.ComputeHashCash(maxIter)
		if err != nil {
			return nil, fmt.Errorf("invalid hashCash")
		}
		//get random quote
		log.Printf("client %s succesfully computed hashCash %s\n", clientInfo, msg.Payload)
		msg := protocol.Message{
			Header:  protocol.ResponseResource,
			Payload: string(phrase),
		}
		cache.Delete(randValue)
		return &msg, nil
	default:
		return nil, fmt.Errorf("unknown header")
	}
}

func sendMsg(msg protocol.Message, conn net.Conn) error {
	msgStr := fmt.Sprintf("%s\n", msg.Stringify())
	_, err := conn.Write([]byte(msgStr))
	return err
}
