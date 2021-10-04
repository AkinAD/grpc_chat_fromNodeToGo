package pubsub

// import (
// 	"context"

// 	"github.com/AkinAD/grpc_chat_fromNodeTosonfigche"
// 	"github.com/AkinAD/grpc_chat_fromNodeTodatabase/Gachefig"
// 	"github.com/AkinAD/grpc_chat_fromNodeToGo/models"
// 	pb "github.com/AkinAD/grpc_chat_fromNodeToGo/proto"
// 	"google.golang.org/protobuf/encoding/protojson"
// 	"google.golang.org/protobuf/reflect/protoreflect"
// )

// type PubSub interface {
// 	Publish(topic string)
// 	Subscribe(topic string)
// 	PubMainChatRoomUpdate(msg *pb.StreamMessage) error
// }

// type RedisPubSub struct {
// 	Publisher cache.RedisClient
// 	Subscriber cache.RedisClient
// 	ctx context.Context
// }

// const (
// 	MAIN_ROOM = "MAIN_ROOM"
// 	USER_CHANGE = "USER_CHANGE"
// )

// func NewRedisPubSub(config config.RedisConfig, manager manager.Manager) PubSub{
// 	ctx := context.Background()
// 	pubClient, err := cache.NewRedisClient(ctx, config.Host, config.Db,config.Expires, logger)
// 	if err != nil {
// 		logger.Error.Fatalf("Failed to initialize publisher: %v", err)
// 	}
// 	publisher := cache.RedisClient{
// 		Client: *pubClient,
// 		Manager: manager
// 	}
	
// 	subClient, err := cache.NewRedisClient(ctx, config.Host, config.Db,config.Expires, logger)
// 	if err != nil {
// 		logger.Error.Fatalf("Failed to initialize subscriber: %v", err)
// 	}
// 	subscriber := cache.RedisClient{
// 		Client: *subClient,
// 		Logger: logger,
// 	}

// 	return &RedisPubSub{
// 		Publisher: publisher,
// 		Subscriber: subscriber,
// 		ctx: ctx,
// 	}
// }

// func (s *RedisPubSub) Publish(topic string) {

// }
// func (s *RedisPubSub) Subscribe(topic string) {

// }



// func (s *RedisPubSub) PubMainChatRoomUpdate(msg *pb.StreamMessage) error {
// 	// s.Publisher.
// 	json, err := s.doProtoMarshal(msg)
// 	if err != nil {
// 		s.Publisher.Logger.Error.Printf("Failed to marshal json in publisher: %v", err)
// 		return err
// 	}
// 	s.Publisher.Client.Publish(s.ctx, MAIN_ROOM, json)
// 	return nil
// }


// func(s *RedisPubSub) SubMainChatRoomUpdate()(*pb.StreamMessage){
	
// 	// Subscribe to the Topic given
// 	topic := s.Subscriber.Client.Subscribe(s.ctx, MAIN_ROOM)
// 	// Get the Channel to use
// 	channel := topic.Channel()
// 	// Itterate any messages sent on the channel
// 	for msg := range channel {
// 		m := &pb.StreamMessage{}
// 		// Unmarshal the data into the message
// 		err := protojson.Unmarshal([]byte(msg.Payload), m)
// 		if err != nil {
// 			s.Publisher.Logger.Error.Printf("Failed to unmarshal json in subscriber: %v", err)
// 		}
// 	}
// }



// func (p *RedisPubSub) doProtoMarshal(item protoreflect.ProtoMessage) ([]byte, error){
// 	m := protojson.MarshalOptions{
// 		Indent:          "  ",
// 		EmitUnpopulated: true,
// 	}
	
// 	return  m.Marshal(item)
// }