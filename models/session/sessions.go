package session

import (
	"context"
	"fmt"

	manager "github.com/AkinAD/grpc_chat_fromNodeToGo/models/Manager"
	pb "github.com/AkinAD/grpc_chat_fromNodeToGo/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

type ChatOnlineSession struct {
	UID string `json:"uid" binding:"required"` // client unique identifier
	SID uuid.UUID `json:"sid" binding:"required"` // session_id
	Stream pb.ChatService_ChatStreamServer
	Err   chan error
	Info chan string
	Send  chan []byte
	Rooms map[string]string // to be updated
	Active bool
}
type ActiveUserListSession struct {
	UID string `json:"uid" binding:"required"` // client unique identifier
	SID uuid.UUID `json:"sid" binding:"required"` // session_id
	Stream pb.ChatService_UserStreamServer
	Err   chan error
	Info chan string
	Send  chan []byte
	Rooms map[string]string // to be updated
	Active bool
}

type SessionEvent int

const (
	SESSION_UNREGISTER SessionEvent =  iota
	SESSION_REGISTER
	
)

// var (
// 	newline = []byte{'\n'}
// )

type Session struct {
	ID string
	ChatOnline *ChatOnlineSession
	UserList *ActiveUserListSession
}


func NewActiveUsersListSession(uid string, stream pb.ChatService_UserStreamServer) *ActiveUserListSession{
	return &ActiveUserListSession{
		UID: uid,
		SID: uuid.New(),
		Stream: stream,
		Err:  make(chan error),
		Send: make(chan []byte, 256),
	}
}

func NewChatOnlineSession(uid string, stream pb.ChatService_ChatStreamServer  ) *ChatOnlineSession{
	return &ChatOnlineSession{
		UID: uid,
		SID: uuid.New(),
		Stream: stream,
		Err:  make(chan error),
		Send: make(chan []byte, 256),
	}
}

//TODO: implement me 
// func (c *ChatOnlineSession) GetUser() pb.User {

// 	return pb.User{}
// } 
// func (u *ActiveUserListSession) GetUser() pb.User {
// 	return pb.User{}
// } 

func (c *ChatOnlineSession) WriteListen(register chan *Session, unregister chan *Session) {
	
	for {
		err := ContextError(c.Stream.Context())
		if err != nil {
			c.Err <- err
		}

		select {
	
		case <-c.Stream.Context().Done() :
			c.Info <- "Chat Stream returned done signal"
			unregister <- &Session{ID: c.UID, ChatOnline: c}
			return
		case message, ok := <-c.Send:
			// c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			c.Info <- "Sending chat packets to stream..."
			if !ok {
				// The WsServer closed the channel.
				c.Stream.Send(&pb.StreamMessage{UserId: c.UID, Message: "Session closed"})   
				return
			}
			var sm *pb.StreamMessage
			err := protojson.Unmarshal(message, sm)
			if err != nil {
				return
			}
			err = c.Stream.Send(sm)
			if err != nil {
				c.Err <- err
			}

			// Attach queued chat messages to the current message. (keep reading till empty)
			n := len(c.Send)
			for i := 0; i < n; i++ {
				var sm_b *pb.StreamMessage

				err := protojson.Unmarshal(<-c.Send, sm_b)
				if err != nil {
					c.Err <- err
				}
				err = c.Stream.Send(sm_b)
				if err != nil {
					c.Err <- err
				}

			}
			c.Info <- "Done sending chat packets"
		}
	}
}

func (u *ActiveUserListSession) WriteListen(register chan *Session, unregister chan *Session) {
	// When a user disconnects from chat
	// don't know if below is useful yet
	
	// err := ContextError(u.Stream.Context())
	// 	if err != nil {
	// 		u.Err <- err
	// 	}
		for {
			select {
				// case <-u.Stream.Context().Done() :
				// 	u.Info <- "UserList Stream returned done signal"
				// 	unregister <- &Session{ID: u.UID, UserList: u}
				// 	return
				case res := <-u.Send: // read from the 'send' channel and use the result you read
					u.Info <- "Sending user list packets to stream..."
					m := &pb.UserStreamResponse{}
					// Unmarshal the data into the message
					err := protojson.Unmarshal([]byte(res), m)
					if err != nil {
						u.Err <- fmt.Errorf("failed to unmarshal json read from channel: %v", err)
					}
					err1 := u.Stream.Send(m) 
					if err1 != nil {
						u.Err <- fmt.Errorf("====AKIN===== Error sending with session embedded stream: %v", err1)
					}

					//  (keep reading channel till empty)
					n := len(u.Send)
					for i := 0; i < n; i++ {
						var ul *pb.UserStreamResponse
						err := protojson.Unmarshal(<-u.Send, ul)
						if err != nil {
							u.Err <- err
						}
						err = u.Stream.Send(ul)
						if err != nil {
							u.Err <- err
						}
					}
					u.Info <- "Done sending user-list packets"

			}
    	}
	
}

func (u *ActiveUserListSession) HandleSessionErrors(manager *manager.Manager) {
	for {
		select {
		case errIn := <-u.Err:
			manager.Logger.Error.Print(errIn) 
		case infoIn := <-u.Info:
			manager.Logger.Info.Print(infoIn) 
		}
	}

}

func (c *ChatOnlineSession) HandleSessionErrors(manager *manager.Manager) {
	for {
		select {
		case errIn := <-c.Err:
			manager.Logger.Error.Print(errIn) 
		case infoIn := <-c.Info:
			manager.Logger.Info.Print(infoIn) 
		
		}
	}

}

func ContextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		return status.Error(codes.Canceled, "context is canceled")
	case context.DeadlineExceeded:
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}