package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	jsonv1 "github.com/go-json-experiment/json/v1"
	"github.com/google/uuid"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/go-a2a/a2a-go"
	"github.com/go-a2a/a2a-go/client"
	"github.com/go-a2a/a2a-go/server"
	"github.com/go-a2a/a2a-go/transport"
)

func main() {
	ctx := context.Background()

	opts := &server.ServerOptions{
		SendMessageHandler: func(ctx context.Context, ss *transport.ServerSession, params *a2a.MessageSendParams) (a2a.MessageOrTask, error) {
			return &a2a.Message{
				Kind:      a2a.MessageEventKind,
				MessageID: params.Message.ContextID,
				Parts:     []a2a.Part{a2a.NewTextPart("hi")},
				Role:      a2a.RoleAgent,
			}, nil
		},
		GetTaskHandler: func(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskQueryParams) (*a2a.Task, error) {
			return a2a.NewTask(&a2a.Message{
				Kind:      a2a.MessageEventKind,
				MessageID: uuid.NewString(),
				Parts:     []a2a.Part{a2a.NewTextPart("hi")},
				Role:      a2a.RoleAgent,
			})
		},
	}
	server := server.NewServer(opts)

	handler := transport.NewSSEHandler(func(req *http.Request) transport.Server {
		url := req.URL.Path
		switch url {
		case "/":
			return server
		default:
			return nil
		}
	})
	h := h2c.NewHandler(handler, &http2.Server{})
	srv := &http.Server{
		Handler: h,
	}

	l, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		if err := srv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	tp := transport.NewSSEClientTransport("http://"+l.Addr().String(), nil)
	client := client.NewClient()
	cs, err := client.Connect(ctx, tp)
	if err != nil {
		log.Fatal(err)
	}
	defer cs.Close()

	res, err := cs.SendMessage(ctx, &a2a.MessageSendParams{
		Message: &a2a.Message{
			Kind:      a2a.MessageEventKind,
			ContextID: uuid.NewString(),
			MessageID: uuid.NewString(),
			Parts:     []a2a.Part{a2a.NewTextPart("test")},
			Role:      a2a.RoleUser,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	data, err := jsonv1.MarshalIndent(res, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data: %s\n", string(data))

	res2, err := cs.GetTask(ctx, &a2a.TaskQueryParams{
		ID: uuid.NewString(),
	})
	if err != nil {
		log.Fatal(err)
	}
	data2, err := jsonv1.MarshalIndent(res2, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data2: %s\n", string(data2))

	// First, close the client connection to signal shutdown intent
	if err := cs.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Client close error: %v\n", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
