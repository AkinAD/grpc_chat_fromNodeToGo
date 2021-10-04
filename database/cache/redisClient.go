package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/AkinAD/grpc_chat_fromNodeToGo/config"
	manager "github.com/AkinAD/grpc_chat_fromNodeToGo/models/Manager"
	pb "github.com/AkinAD/grpc_chat_fromNodeToGo/proto"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/go-redis/redis/v8"
)


type RedisClient struct {
	redisConfig config.RedisConfig
	Client      redis.Client
	Manager *manager.Manager
}

type RedisCache interface {
	Set(key string, value *[]byte)
	Get(key string) *[]byte
	AddUser(user *pb.User)
	ListUsers() []*pb.User
	UpdateUser(user *pb.User) error
	GetUser(id string) (*pb.User, error)

	AddMessageToChatroom(m *pb.StreamMessage) (error)
	ListMessagesInRoom() ([]*pb.StreamMessage)

	// GetActiveUserListSession(key string) *session.ActiveUserListSession
	// SetActiveUserListSession(key string, sesh *session.ActiveUserListSession)
	GetClient() *redis.Client
}

var redis_keys = map[string]string{
	"broadcastRoom": "room:0:messages",
	"users":         "users",
}

// advised not to use
var ExistingRedis *redis.Client
// advised not to use
var ExistingRedisClient *RedisClient

func NewRedisClient(ctx context.Context, host string, db int, exp time.Duration, logger manager.Logger) (*redis.Client, error){
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       db,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error.Printf("Error c reating redis client: %v", err)
		return nil, err
	}
	return client, nil
}

// Creates a new redis cache interface for communicating with a redis client
func NewRedisCache(m *manager.Manager, host string, db int, exp time.Duration) (RedisCache, error) {
	logger := m.Logger
	client := redis.NewClient(&redis.Options{
		Addr:     host,
		Password: "",
		DB:       db,
	})

	if err := client.Ping(m.Ctx).Err(); err != nil {
		logger.Error.Printf("Error reaching redis: %v", err)
		return nil, err
	}

	redisCache := &RedisClient{
		redisConfig: config.RedisConfig{Host: host, Db: db, Expires: exp},
		Manager: m,
		Client:      *client,
	}
	ExistingRedis = client
	ExistingRedisClient = redisCache

	return redisCache, nil
}

func (cache *RedisClient) Set(key string, value *[]byte) {
	client := cache.Client

	// if err != nil {
	// 	log.Errorf("Unable to marshal data for cahce set: %v", err)
	// }
	client.Set(cache.Manager.Ctx, key, value, cache.redisConfig.Expires)

}

func (cache *RedisClient) Get(key string) *[]byte {
	client := cache.Client

	val, err := client.Get(cache.Manager.Ctx, key).Result()
	if err != nil {
		cache.Manager.Logger.Error.Printf("Error fetching from cache: %v", err)
		return nil
	}
	res := []byte(val)
	return &res
}

// ================ converted from video tutorial ====================
func (cache *RedisClient) AddUser(user *pb.User) {
	json := cache.doProtoMarshal(user)
	cache.Client.RPush(cache.Manager.Ctx, redis_keys["users"], json)
}

func (cache *RedisClient) ListUsers() []*pb.User {
	// logger := cache.Logger
	fetched, err := cache.Client.LRange(cache.Manager.Ctx, redis_keys["users"], 0, -1).Result()
	if err != nil {
		cache.Manager.Logger.Error.Printf("Error reading users list from cache: %v", err)
	}
var users []*pb.User
	for _, item := range fetched {
		user := pb.User{}
		protojson.Unmarshal([]byte(item), &user)
		
		users = append(users, &user)
	}
	return users
}

func (cache *RedisClient) UpdateUser(user *pb.User) error {
	found := false
	for i, u := range cache.ListUsers() {
		if u.Id == user.Id {
			found = true
			json := cache.doProtoMarshal(user)
			cache.Client.LSet(cache.Manager.Ctx, redis_keys["users"], int64(i), json)
			break
		}
	}
	if !found {
		return errors.New("unable to update: user not found")
	}
	return nil
}

func (cache *RedisClient) GetUser(id string) (*pb.User, error) {
	for _, u := range cache.ListUsers() {
		if u.Id == id {
			return u, nil
		}
	}

	return nil, errors.New("unable to update: user not found")
}

func (cache *RedisClient) AddMessageToChatroom(m *pb.StreamMessage) (error) {
	json,err := json.Marshal(m)
	if err != nil {
		return  err
	}
	cache.Client.RPush(cache.Manager.Ctx, redis_keys["broadcastRoom"], json)
	return nil
}

func (cache *RedisClient) ListMessagesInRoom() ([]*pb.StreamMessage) {
	fetched, err := cache.Client.LRange(cache.Manager.Ctx, redis_keys["broadcastRoom"], 0, -1).Result()
	if err != nil {
		cache.Manager.Logger.Error.Printf("Error reading messages list from cache: %v", err)
	}

	var msgs []*pb.StreamMessage
	for _, item := range fetched {
		msg := pb.StreamMessage{}
		err:=protojson.Unmarshal([]byte(item), &msg)

		if err != nil {
			cache.Manager.Logger.Error.Printf("Error unmarshalling chats: %v",err)
			return nil
		}
		msgs = append(msgs, &msg)
	}
	return msgs
}


// ===================

// func (cache *RedisClient) GetActiveUserListSession(key string) *session.ActiveUserListSession {
// 	session := session.ActiveUserListSession{}
// 	fetched := cache.Get(key)
// 	err := json.Unmarshal(*fetched, &session)
// 	if err != nil {
// 		cache.Manager.Logger.Error.Panicf("Unable to unmarshal value: %v", err)
// 		return nil
// 	}

// 	return &session
// }

// func (cache *RedisClient) SetActiveUserListSession(key string, sesh *session.ActiveUserListSession) {
// 	json, err := json.Marshal(sesh)
// 	if err != nil {
// 		cache.Manager.Logger.Error.Panicf("Unable to Marshal value: %v", err)

// 	}
// 	cache.Set(key, &json)

// }

func (cache *RedisClient) GetClient() *redis.Client {
	return  &cache.Client
}


func (cache *RedisClient) doProtoMarshal(item protoreflect.ProtoMessage) []byte{
	m := protojson.MarshalOptions{
		Indent:          "  ",
		EmitUnpopulated: true,
	}

	json, err := m.Marshal(item)
	if err != nil {
		cache.Manager.Logger.Error.Panicf("ProtoJson unnable to Marshal value: %v", err)
	}
	return json
}

