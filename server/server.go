package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/cache"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/db"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/models"
	pb "github.com/AkinAD/grpc_chat_fromNodeToGo/proto" // your module in go.mod and the proto must match
	"google.golang.org/grpc"
)

const (
	PORT       = ":8082"
	PROTO_FILE = "./proto/chat.proto"
	REDIS_PORT = ":6379"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	database *db.Database
}

func (server *ChatServer) Run() error {
	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterChatServiceServer(s, server)
	log.Printf("server listening at %v", lis.Addr())
	return s.Serve(lis)
}

func NewChatServer(ctx context.Context, logger models.Logger) *ChatServer {
	cacheConfig := cache.RedisConfig{Host: fmt.Sprintf("localhost%v", REDIS_PORT), Db: 0, Expires: 600}
	database, err := db.InitDatabases(ctx, cacheConfig, logger)
	if err != nil {
		log.Fatalf("Failed to connect to redis: %s", err.Error())
	}

	return &ChatServer{
		database: database,
	}
}

func (s *ChatServer) ChatInitiate(ctx context.Context, in *pb.InitiateRequest) (*pb.InitiateResponse, error) {
	//  sessionName := in.Name
	//  avatar := in.AvatarUrl
	newID := int32(rand.Intn(10000))
	log.Printf("Received chat initiate request, returning id %d", newID)
	return &pb.InitiateResponse{Id: newID}, nil
}

func (s *ChatServer) ChatStream(in *pb.StreamRequest, idk pb.ChatService_ChatStreamServer) error {
	return nil
}

func main() {
	Logger := models.NewLogger()
	chatServer := NewChatServer(context.Background(), *Logger)
	if err := chatServer.Run(); err != nil {
		log.Fatalf("failed to serve : %v", err)
	}
}

// npx create-react-app grpc-chat-app --template cra-template-pwa-typescript
