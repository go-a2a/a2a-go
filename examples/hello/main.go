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

func init() {
	// Enable the use of the random pool for UUID generation.
	uuid.EnableRandPool()
}

func main() {
	ctx := context.Background()

	agentCard := makeAgentCard()
	opts := &server.ServerOptions{
		SendMessageHandler: func(ctx context.Context, ss *transport.ServerSession, params *a2a.MessageSendParams) (a2a.MessageOrTask, error) {
			return &a2a.Message{
				Kind:      a2a.MessageEventKind,
				MessageID: params.Message.ContextID,
				Parts: []a2a.Part{
					a2a.NewTextPart("hello"),
					a2a.NewTextPart("world"),
				},
				Role: a2a.RoleAgent,
			}, nil
		},
		SendStreamMessageHandler: func(ctx context.Context, ss *transport.ServerSession, params *a2a.MessageSendParams) (a2a.SendStreamingMessageResponse, error) {
			return &a2a.Message{
				Kind:      a2a.MessageEventKind,
				MessageID: params.Message.ContextID,
				Parts: []a2a.Part{
					a2a.NewTextPart("hello"),
					a2a.NewTextPart("stream"),
				},
				Role: a2a.RoleAgent,
			}, nil
		},
		GetTaskHandler: func(ctx context.Context, ss *transport.ServerSession, params *a2a.TaskQueryParams) (*a2a.Task, error) {
			return a2a.NewTask(&a2a.Message{
				Kind:      a2a.MessageEventKind,
				MessageID: uuid.NewString(),
				Parts: []a2a.Part{
					a2a.NewTextPart("hello"),
					a2a.NewTextPart("get task"),
				},
				Role: a2a.RoleAgent,
			})
		},
	}
	server := server.NewServer(agentCard, opts)

	handler := transport.NewServerHandler(func(req *http.Request) transport.Server {
		return server
	}, nil)
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
	client := client.NewClient(nil)
	cs, err := client.Connect(ctx, tp)
	if err != nil {
		log.Fatal(err)
	}
	defer cs.Close()

	// ac, err := cs.GetAgentCard(ctx, "http://"+l.Addr().String())
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// acData, err := jsonv1.MarshalIndent(ac, "", "  ")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("agentCard: %s\n", string(acData))

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

	res2, err := cs.SendStreamMessage(ctx, &a2a.MessageSendParams{
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
	data2, err := jsonv1.MarshalIndent(res2, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data2: %s\n", string(data2))

	res3, err := cs.GetTask(ctx, &a2a.TaskQueryParams{
		ID: uuid.NewString(),
	})
	if err != nil {
		log.Fatal(err)
	}
	data3, err := jsonv1.MarshalIndent(res3, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("data3: %s\n", string(data3))

	// First, close the client connection to signal shutdown intent
	if err := cs.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Client close error: %v\n", err)
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func makeAgentCard() *a2a.AgentCard {
	return &a2a.AgentCard{
		ProtocolVersion:    "0.2.6",
		Name:               "GeoSpatial Route Planner Agent",
		Description:        "Provides advanced route planning, traffic analysis, and custom map generation services. This agent can calculate optimal routes, estimate travel times considering real-time traffic, and create personalized maps with points of interest.",
		URL:                "http://127.0.0.1/a2a/v1",
		PreferredTransport: a2a.TransportProtocolJSONRPC,
		Provider: &a2a.AgentProvider{
			Organization: "Example Geo Services Inc.",
			URL:          "https://www.examplegeoservices.com",
		},
		IconURL:          "https://georoute-agent.example.com/icon.png",
		Version:          "0.0.0",
		DocumentationURL: "https://docs.examplegeoservices.com/georoute-agent/api",
		Capabilities: &a2a.AgentCapabilities{
			Streaming:              true,
			PushNotifications:      true,
			StateTransitionHistory: false,
		},
		SecuritySchemes: map[string]a2a.SecurityScheme{
			"google": &a2a.OpenIDConnectSecurityScheme{
				Type:             a2a.OpenIDConnectSecuritySchemeType,
				OpenIDConnectURL: "https://accounts.google.com/.well-known/openid-configuration",
			},
		},
		Security: []map[string]a2a.SecurityScopes{
			{
				"google": a2a.NewSecurityScopes("openid", "profile", "email"),
			},
		},
		DefaultInputModes:  []string{"application/json", "text/plain"},
		DefaultOutputModes: []string{"application/json", "image/png"},
		Skills: []*a2a.AgentSkill{
			{
				ID:          "route-optimizer-traffic",
				Name:        "Traffic-Aware Route Optimizer",
				Description: "Calculates the optimal driving route between two or more locations, taking into account real-time traffic conditions, road closures, and user preferences (e.g., avoid tolls, prefer highways).",
				Tags:        []string{"maps", "routing", "navigation", "directions", "traffic"},
				Examples:    []string{"Plan a route from '1600 Amphitheatre Parkway, Mountain View, CA' to 'San Francisco International Airport' avoiding tolls.", "{\"origin\": {\"lat\": 37.422, \"lng\": -122.084}, \"destination\": {\"lat\": 37.7749, \"lng\": -122.4194}, \"preferences\": [\"avoid_ferries\"]}"},
				InputModes:  []string{"application/json", "text/plain"},
				OutputModes: []string{"application/json", "application/vnd.geo+json", "text/html"},
			},
			{
				ID:          "custom-map-generator",
				Name:        "Personalized Map Generator",
				Description: "Creates custom map images or interactive map views based on user-defined points of interest, routes, and style preferences. Can overlay data layers.",
				Tags:        []string{"maps", "customization", "visualization", "cartography"},
				Examples:    []string{"Generate a map of my upcoming road trip with all planned stops highlighted.", "Show me a map visualizing all coffee shops within a 1-mile radius of my current location."},
				InputModes:  []string{"application/json"},
				OutputModes: []string{"image/png", "image/jpeg", "application/json", "text/html"},
			},
		},
		SupportsAuthenticatedExtendedCard: true,
		Signatures: &a2a.AgentCardSignature{
			Protected: "eyJhbGciOiJFUzI1NiIsInR5cCI6IkpPU0UiLCJraWQiOiJrZXktMSIsImprdSI6Imh0dHBzOi8vZXhhbXBsZS5jb20vYWdlbnQvandrcy5qc29uIn0",
			Signature: "QFdkNLNszlGj3z3u0YQGt_T9LixY3qtdQpZmsTdDHDe3fXV9y9-B3m2-XgCpzuhiLt8E0tV6HXoZKHv4GtHgKQ",
		},
	}
}
