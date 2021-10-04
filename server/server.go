package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/config"
	db "github.com/AkinAD/grpc_chat_fromNodeToGo/database"
	manager "github.com/AkinAD/grpc_chat_fromNodeToGo/models/Manager"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/models/room"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/models/session"
	pb "github.com/AkinAD/grpc_chat_fromNodeToGo/proto" // your module in go.mod and the proto must match
	"github.com/google/uuid"
	"google.golang.org/protobuf/encoding/protojson"
	emptypb "google.golang.org/protobuf/types/known/emptypb"

	"google.golang.org/grpc"
)

const (
	PORT       = ":8082"
	PROTO_FILE = "./proto/chat.proto"
	REDIS_PORT = ":6379"
	GENERAL_ROOM = "GENERAL_ROOM"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	database *db.Database
	rooms          map[*room.Room]bool
	userListSessions  map[string]*session.ActiveUserListSession
	ChatOnlineSessions  map[string]*session.ChatOnlineSession
	register       chan *session.Session
	unregister       chan *session.Session
	Manager manager.Manager
}

func (server *ChatServer) Run() error {
	logger := server.Manager.Logger
	lis, err := net.Listen("tcp", PORT)
	if err != nil {
		logger.Error.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()

	pb.RegisterChatServiceServer(s, server)
	logger.Info.Printf("server listening at %v", lis.Addr())
	go server.HandleSessionEvents()
	// run the general room 
	generalRoom := server.findRoomByName(GENERAL_ROOM)
	go generalRoom.RunRoom(&server.Manager)
	return s.Serve(lis)
}

func (server *ChatServer) HandleSessionEvents() {
	go server.listenPubSubChannel()
	logger := server.Manager.Logger
	for {
		select {

		case session := <-server.register:
			server.registerSession(session)

		case session := <-server.unregister:
			logger.Debug.Printf("=== client sent unregister request [%v]", session)
			server.unregisterSession(session)
		}

	}
}

func (s *ChatServer) unregisterSession(sesh *session.Session) {
	logger := s.Manager.Logger
	if sesh.ChatOnline != nil{
		user, err := s.database.RedisClient.GetUser(sesh.ID)
		if err != nil {
			logger.Error.Printf("Could not get user attached to session: %v", err)
		}
		logger.Info.Println("Chat stream closing...")
		user.Status = pb.Status_OFFLINE.String()
		s.database.RedisClient.UpdateUser(user)
		delete(s.userListSessions, user.Id)
	} else {
		user, err := s.database.RedisClient.GetUser(sesh.ChatOnline.UID)
		if err != nil {
			logger.Error.Printf("Could not get user attached to session: %v", err)
		}
		logger.Info.Println("User list stream closing...")
		user.Status = pb.Status_OFFLINE.String()
		s.database.RedisClient.UpdateUser(user)
		delete(s.ChatOnlineSessions, user.Id)
	}
		//TODO: emit user update when users join or leave
		// here 
	
}

func (s *ChatServer) registerSession(sesh *session.Session) {
	
}

func NewChatServer(m *manager.Manager) *ChatServer {
	logger := m.Logger
	cacheConfig := config.RedisConfig{Host: fmt.Sprintf("localhost%v", REDIS_PORT), Db: 0, Expires: 600}
	sqliteConfig := config.SQLiteConfig{DriverName: "sqlite3", DataSourceName: "./chatdb.db"}
	database, err := db.InitDatabases(*m, cacheConfig, sqliteConfig)
	if err != nil {
		logger.Error.Fatalf("Failed to connect to redis: %s", err.Error())
	}

	return &ChatServer{
		database: database,
		Manager: *m,
		rooms: make(map[*room.Room]bool),
		userListSessions: make(map[string]*session.ActiveUserListSession),
		register: make(chan *session.Session),
		unregister: make(chan *session.Session),
		ChatOnlineSessions: make(map[string]*session.ChatOnlineSession),
	}
}

func (s *ChatServer) ChatInitiate(ctx context.Context, in *pb.InitiateRequest) (*pb.InitiateResponse, error) {
	logger := s.Manager.Logger
	newID := uuid.New()

 	logger.Info.Printf("Received chat initiate request, returning id %s", newID.String())
	allUsers := s.database.RedisClient.ListUsers()
	var dbUser *pb.User
	for _, user := range allUsers {
		if  strings.EqualFold(user.Name, in.Name) {
			// Found!
			logger.Info.Printf("Found existing user with username %v", user.Name)
			dbUser = user
			break
		}
	}
	if dbUser == nil {
		logger.Info.Println("Could not find existing, creating new user ...")

		newUser := pb.User{Id: newID.String(), Status: pb.Status_ONLINE.String(), Name: in.GetName(), AvatarUrl: in.GetAvatarUrl()}
		
		s.database.RedisClient.AddUser(&newUser)
		dbUser = &newUser
		//TODO: add this user to general room 
		// do this in stream initiation (userlist and chat stream)
	} else {
		if dbUser.Status == pb.Status_ONLINE.String() {
			logger.Warning.Println("User exists and is online")
		}
		dbUser.Status = pb.Status_ONLINE.String();
		s.database.RedisClient.UpdateUser(dbUser) 
		return &pb.InitiateResponse{Id: dbUser.Id}, nil
	}


	return &pb.InitiateResponse{Id: newID.String()}, nil
}

func (s *ChatServer) SendMessage(ctx context.Context, in *pb.MessageRequest) (*emptypb.Empty, error) {
	logger := s.Manager.Logger
	msgNotSent := "message not sent for user"
	user, err := s.database.RedisClient.GetUser(in.Id)
	if err != nil {
		logger.Error.Printf("%v %v : %v",msgNotSent, in.Id, err)
		return &emptypb.Empty{}, err
	}

	msg := &pb.StreamMessage{
		UserId:user.Id , Message: in.GetMessage(), UserAvatar: user.AvatarUrl, UserName: user.Name,
	}
	err = s.database.RedisClient.AddMessageToChatroom(msg)
	rm := s.findRoomByName(GENERAL_ROOM)
	rm.Broadcast(msg)

	if err != nil {
		logger.Error.Printf("%v %v : %v",msgNotSent, in.Id, err)
		return &emptypb.Empty{}, err
	}
	return  &emptypb.Empty{}, nil
}


func (s *ChatServer) ChatStream(in *pb.StreamRequest, stream pb.ChatService_ChatStreamServer) error {
	logger := s.Manager.Logger
	errorWithChatStream := "Error with chat stream"
	user, err := s.database.RedisClient.GetUser(in.Id)
	if err != nil {
		logger.Error.Printf("%v %v : %v",errorWithChatStream, in.Id, err)
		return err
	}
	msgs := s.database.RedisClient.ListMessagesInRoom()
	logger.Info.Printf("Spamming %d previous messages to user %v", len(msgs), user.Id)
	for _, m := range msgs {
			err := stream.Send(m)
			if err != nil {
				logger.Error.Printf("Error distributing messages: %v",err)
			}
	}

	sesh, exists := s.ChatOnlineSessions[in.Id] 
	sessionForThisClient := session.NewChatOnlineSession(in.Id, stream)

	if !exists {
        s.ChatOnlineSessions[in.Id] = sessionForThisClient
    } else {
        sessionForThisClient = sesh
    }

	rm := s.findRoomByName(GENERAL_ROOM)
	logger.Debug.Print("Checking if user CHAT session is registered in room...")
	rm.UpdateRegistration(&s.Manager, sessionForThisClient, nil, in.Id)
	logger.Info.Print("Starting chat stream...")
	// go sessionForThisClient.WriteListen(s.register, s.unregister)

	go func () {
		// err := session.ContextError(sessionForThisClient.Stream.Context())
		// if err != nil {
		// 	logger.Debug.Printf("Context error. sigh %v", err)
		// 	sessionForThisClient.Err <- err
		// }
		for {
			sessionForThisClient.Err <- fmt.Errorf("ye")
			select {
				// case <-sessionForThisClient.Stream.Context().Done() :
				// 	sessionForThisClient.Info <- "UserList Stream returned done signal"
				// 	s.unregister <- &session.Session{ID: sessionForThisClient.UID, ChatOnline: sessionForThisClient}
					// return
				case res := <-sessionForThisClient.Send: // read from the 'send' channel and use the result you read
					sessionForThisClient.Info <- "Sending user list packets to stream..."
					m := &pb.StreamMessage{}
					// Unmarshal the data into the message
					err := protojson.Unmarshal([]byte(res), m)
					if err != nil {
						sessionForThisClient.Err <- fmt.Errorf("failed to unmarshal json read from channel: %v", err)
					}
					err1 := stream.Send(m) 
					if err1 != nil {
						sessionForThisClient.Err <- fmt.Errorf("====AKIN===== Error sending with session embedded stream: %v", err1)
					}

					//  (keep reading channel till empty)
					n := len(sessionForThisClient.Send)
					for i := 0; i < n; i++ {
						var ul *pb.StreamMessage
						err := protojson.Unmarshal(<-sessionForThisClient.Send, ul)
						if err != nil {
							sessionForThisClient.Err <- err
						}
						err = stream.Send(ul)
						if err != nil {
							sessionForThisClient.Err <- err
						}
					}
					sessionForThisClient.Info <- "Done sending user-list packets"
				default:
					sessionForThisClient.Info <- "ping"
			}
    	}}()
	go sessionForThisClient.HandleSessionErrors(&s.Manager)
	
	return nil
}

func (s *ChatServer) UserStream(in *pb.StreamRequest, stream pb.ChatService_UserStreamServer) error {
	logger := s.Manager.Logger
	errorWithChatStream := "Error with chat stream"
	user, err := s.database.RedisClient.GetUser(in.Id)
	if err != nil {
		logger.Error.Printf("%v %v : %v",errorWithChatStream, in.Id, err)
		return err
	}
	users := s.database.RedisClient.ListUsers()
	logger.Info.Printf("Spamming user list of size (%d) to user %v", len(users), user.Id)
	res:= pb.UserStreamResponse{Users: users}
	stream.Send(&res)
	//TODO: emit user update when users join or leave

	sesh, exists := s.userListSessions[in.Id] 
	sessionForThisClient := session.NewActiveUsersListSession(in.Id, stream)

	

	if !exists {
        s.userListSessions[in.Id] = sessionForThisClient
    } else {
        sessionForThisClient = sesh
    }
	
	rm := s.findRoomByName(GENERAL_ROOM)
	logger.Debug.Print("Checking if user LIST session is registered in room...")
	rm.UpdateRegistration(&s.Manager, nil, sessionForThisClient, in.Id)
		
	logger.Info.Print("Starting user list stream...")
	go sessionForThisClient.WriteListen(s.register, s.unregister)
	go sessionForThisClient.HandleSessionErrors(&s.Manager)
	//https://stackoverflow.com/questions/59726003/grpc-keep-a-reference-to-stream-to-send-data-to-multiple-clients
	//https://stackoverflow.com/questions/58614680/how-to-create-server-for-persistent-stream-aka-pubsub-in-golang-grpc

	// md, ok := metadata.FromIncomingContext(stream.Context())
	// 	if ok {
	// 		logger.Info.Printf("Metadata from connection: %v", md)
	// 	}
	// 	p, ok := peer.FromContext(stream.Context())
	// 	if ok {
	// 		logger.Info.Printf("Connected peers: %v", p)
	// 	}


	return nil
}

func (s *ChatServer) listenPubSubChannel() {
	redis := s.database.RedisClient.GetClient()
	messagesPubSub := redis.Subscribe(s.Manager.Ctx, GENERAL_ROOM)
	userListPubSub := redis.Subscribe(s.Manager.Ctx, GENERAL_ROOM)
	logger := s.Manager.Logger
	msgCh := messagesPubSub.Channel()
	ulCh := userListPubSub.Channel()

	for msg := range msgCh {

		var message *pb.StreamMessage
		if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
			log.Printf("Error on unmarshal JSON message %s", err)
			return
		}
		logger.Info.Printf("Message Received: %v ", message)
		// switch message.Action {
		// case UserJoinedAction:
		// 	server.handleUserJoined(message)
		// case UserLeftAction:
		// 	server.handleUserLeft(message)
		// case JoinRoomPrivateAction:
		// 	server.handleUserJoinPrivate(message)
		// }
	}
	for user := range ulCh {

		var message *pb.UserStreamResponse
		if err := json.Unmarshal([]byte(user.Payload), &message); err != nil {
			log.Printf("Error on unmarshal JSON message %s", err)
			return
		}
		logger.Info.Printf("Message Received: %v ", message)
	}
}

func (server *ChatServer) findRoomByName(name string) *room.Room {
	var foundRoom *room.Room
	for room := range server.rooms {
		if room.GetName() == name {
			foundRoom = room
			break
		}
	}

	if foundRoom == nil {
		// Try to run the room from the repository, if it is found.
		foundRoom = server.runRoomFromRepository(name)
	}

	return foundRoom
}

func (s *ChatServer) runRoomFromRepository(name string) *room.Room {
	logger := s.Manager.Logger
	var rm *room.Room
	dbRoom := s.database.SQLite.RoomRepository().FindRoomByName(name)
	if dbRoom != nil {
		
		rm = room.NewRoom(dbRoom.GetName(), dbRoom.GetPrivate(), &s.Manager)
		rm.ID, _ = uuid.Parse(dbRoom.GetId())

		go rm.RunRoom(&s.Manager)
		s.rooms[rm] = true
	} else if name == GENERAL_ROOM {
		logger.Info.Println("Creating General room")
		rm = s.createRoom(name, false)
	}

	return rm
}


func (s *ChatServer) createRoom(name string, private bool) *room.Room {
	rm := room.NewRoom(name, private, &s.Manager)
	s.database.SQLite.RoomRepository().AddRoom(*rm)

	go rm.RunRoom(&s.Manager)
	s.rooms[rm] = true

	return rm
}

func main() {
	manager := manager.NewManager()
	chatServer := NewChatServer(&manager)
	// chatServer.database.RedisClient.AddUser(pb.User{Name: "Akin",AvatarUrl: "https://filmschoolrejects.com/wp-content/uploads/2018/10/avatar-last-airbender-episodes-ranked.jpg"})
	// chatServer.database.RedisClient.AddUser(pb.User{Name: "Someone Else",AvatarUrl: "https://i.ytimg.com/vi/aEvItEpMly8/maxresdefault.jpg"})
	// allUsers := chatServer.database.RedisClient.ListUsers()

	// for _, some := range allUsers {
	// 	Logger.Info.Printf("A user: %v, %v, %v \n", some.Name, some.AvatarUrl, some.Status)
	// }
	if err := chatServer.Run(); err != nil {
		log.Fatalf("failed to serve : %v", err)
	}
}

// npx create-react-app grpc-chat-app --template cra-template-pwa-typescript
