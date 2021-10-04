package room

import (
	"fmt"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/database/cache"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/models"
	manager "github.com/AkinAD/grpc_chat_fromNodeToGo/models/Manager"
	"github.com/AkinAD/grpc_chat_fromNodeToGo/models/session"
	pb "github.com/AkinAD/grpc_chat_fromNodeToGo/proto"

	"github.com/google/uuid"
)

const welcomeMessage = "%s showed up on your radar!"

type Room struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	sessions    map[*session.Session]bool
	register   chan *session.Session
	unregister chan *session.Session
	broadcast  chan *pb.StreamMessage
	Private    bool `json:"private"`
	manager *manager.Manager
}

// NewRoom creates a new Room
func NewRoom(name string, private bool, m *manager.Manager) *Room {
	return &Room{
		ID:         uuid.New(),
		Name:       name,
		sessions:    make(map[*session.Session]bool),
		register:   make(chan *session.Session),
		unregister: make(chan *session.Session),
		broadcast:  make(chan *pb.StreamMessage),
		Private:    private,
		manager: m,
	}
}

// RunRoom runs our room, accepting various requests
func (room *Room) RunRoom(m *manager.Manager) {
	go room.subscribeToRoomMessages(m)
	logger := m.Logger

	for {
		select {

		case client := <-room.register:
			logger.Debug.Printf("Registering client session [%v] from room [%v]", client, room.Name)
			room.registerClientInRoom(m, client)
			_= client

		case client := <-room.unregister:
			logger.Debug.Printf("Unregistering client session [%v] from room [%v]", client, room.Name)
			room.unregisterClientInRoom(m, client)

		case message := <-room.broadcast:
			logger.Debug.Printf("Received message [%v] in room [%v]", message, room.Name)
			room.publishRoomMessage(m, models.DoProtoMarshal(message, m.Logger))
		}

	}
}

func (room *Room) registerClientInRoom(m *manager.Manager, client *session.Session) {
	logger := m.Logger

	if !room.Private {
		logger.Debug.Printf("Adding user [%v] to room [%v]", client.ID, room.Name)
		room.notifyClientJoined(m, client)
	}
	room.sessions[client] = true
}

func (room *Room) unregisterClientInRoom(m *manager.Manager,client *session.Session) {
	logger := m.Logger
	if _, ok := room.sessions[client]; ok {
		delete(room.sessions, client)
		logger.Debug.Printf("Removed user [%v] from room [%v]", client.ID, room.Name)
	}
}

func (room *Room) broadcastToSessionsInRoom(message []byte) {
	for client := range room.sessions {
		if client.ChatOnline == nil {
			room.manager.Logger.Error.Printf("Client [%v] CHAT session is uninitialilized, skipping ", client.ID)
			room.manager.Logger.Warning.Printf( client.ID)
			continue
		}
		client.ChatOnline.Send <- message
	}
	room.manager.Logger.Warning.Print(room.sessions)
}

func (room *Room) publishRoomMessage(m *manager.Manager, message []byte) {
	logger := m.Logger
	logger.Debug.Printf("Publishing mesage to  room [%v]",  room.Name)
	err :=  cache.ExistingRedis.Publish(m.Ctx, room.GetName(), message).Err()

	if err != nil {
		logger.Debug.Printf("Error publishing message to redis: %v",err)
	}
}


func (room *Room) subscribeToRoomMessages(m *manager.Manager) {
	logger := m.Logger
	logger.Debug.Printf("Subscribing to mesage updates in room [%v]", room.Name)
	pubsub := cache.ExistingRedis.Subscribe(m.Ctx, room.GetName())

	ch := pubsub.Channel()
	logger.Debug.Printf("Broadcasting existing messages to room [%v]",room.Name)
	for msg := range ch {
		room.broadcastToSessionsInRoom([]byte(msg.Payload))
	}
	logger.Debug.Printf("Finished broadcasting existing messages to room [%v]",room.Name)
}

// TODO: use later
func (room *Room) notifyClientJoined(m *manager.Manager, s *session.Session) {
	logger := m.Logger
	// Action:  SendMessageAction,
	// Target:  room,
	// Message: fmt.Sprintf(welcomeMessage, session.ChatOnline.GetUser().Name),

	user, err :=  cache.ExistingRedisClient.GetUser(s.ID)
	if err != nil {
		logger.Error.Printf("Error getting user to notify of new session: %v", err)
	}

	message := &pb.StreamMessage{
		UserId: user.Id, UserName: user.Name, UserAvatar: user.AvatarUrl, Message: fmt.Sprintf(welcomeMessage, user.Name),
	}
	
	room.publishRoomMessage(m, models.DoProtoMarshal(message, m.Logger))
}
func (r *Room) Register(m *manager.Manager, msg  *session.Session, userId string) {
	logger := m.Logger
	logger.Debug.Printf("User [%v] with session added to room [%v]",userId, r.Name )
	for client := range r.sessions {
		if client.ID == userId {
			if msg.ChatOnline != nil {
				client.ChatOnline = msg.ChatOnline
			} 
			r.register <- client
			return // i.e the code below will not execute at all 
		}
	} 
	r.register <- msg
}

func (r *Room) UnRegister(m *manager.Manager, msg *session.Session, userId string) {
	logger := m.Logger
	logger.Debug.Printf("User [%v] with session removed from room [%v]",userId, r.Name )
	for client := range r.sessions {
		if client.ID == userId {
			r.unregister <- client 
			return // i.e the code below will not execute at all 
		}
	} 
	r.unregister <- msg
}

func (r *Room) UpdateRegistration(m *manager.Manager, chat *session.ChatOnlineSession, list *session.ActiveUserListSession, userId string) {
	logger := m.Logger
	// logger.Debug.Printf("User [%v] with session removed from room [%v]",msg.ChatOnline.SID, r.Name )
	update := &session.Session{ID: userId}
	

	if chat != nil {
		var exists bool
		updateRequired := true
		// update if already known
		for client := range r.sessions {
			if client.ID == userId && client.ChatOnline == nil {
				logger.Debug.Printf("Updating chat session for user [%v].", userId)
				client.ChatOnline = chat
				exists = true
				break
			}
			if client.ID == userId && client.ChatOnline != nil {
				updateRequired = false
				break
			}
		}
		// create if dne
		if !exists && updateRequired {
			logger.Debug.Printf("Did not find existing  for user [%v]. Creating session and updating CHAT session.", userId)
			update.ChatOnline = chat
			r.Register(m, update, userId)
		}
	} 
	if list != nil {
		var exists bool
		updateRequired := true
		// update if already known
		for client := range r.sessions {
			if client.ID == userId && client.UserList == nil {
				logger.Debug.Printf("Updating list session for user [%v]. Updating.", userId)
				client.UserList = list
				exists = true
				break
			}
			if client.ID == userId && client.UserList != nil {
				updateRequired = false
				break
			}
		}
		// create if dne
		if  !exists && updateRequired {
			logger.Debug.Printf("Did not find session for user [%v]. Creating session and updating LIST session.", userId)
			update.UserList = list
			r.Register(m, update, userId)
		}
	}
}

func (r *Room) Broadcast(msg *pb.StreamMessage) {

	r.broadcast <- msg
}

func (room *Room) GetId() string {
	return room.ID.String()
}

func (room *Room) GetName() string {
	return room.Name
}

func (room *Room) GetPrivate() bool {
	return room.Private
}