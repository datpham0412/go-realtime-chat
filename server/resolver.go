package server

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/segmentio/ksuid"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	redis       *redis.Client
	mutex       sync.RWMutex
	subscribers map[string][]chan *Message
}

func NewResolver(redisClient *redis.Client) *Resolver {
	return &Resolver{
		redis:       redisClient,
		subscribers: make(map[string][]chan *Message),
	}
}

// Helper method to broadcast messages
func (r *Resolver) broadcastMessage(message *Message) {
	r.mutex.RLock()
	log.Printf("[DEBUG] Starting broadcast for message ID: %s", message.ID)
	log.Printf("[DEBUG] Current subscribers count: %d", len(r.subscribers))

	// Take a snapshot of channels
	var channels []chan *Message
	for _, userChannels := range r.subscribers {
		channels = append(channels, userChannels...)
	}
	r.mutex.RUnlock()

	log.Printf("[DEBUG] Broadcasting to %d channels", len(channels))

	// Broadcast to all channels
	for _, ch := range channels {
		select {
		case ch <- message:
			log.Printf("[DEBUG] Successfully sent message ID: %s to a channel", message.ID)
		default:
			log.Printf("[DEBUG] Failed to send message to channel (blocked or closed)")
		}
	}
	log.Printf("[DEBUG] Broadcast complete for message ID: %s", message.ID)
}

// Add a helper method for channel cleanup
func (r *Resolver) cleanupChannel(user string, ch chan *Message) {
	if channels, exists := r.subscribers[user]; exists {
		for i, c := range channels {
			if c == ch {
				r.subscribers[user] = append(channels[:i], channels[i+1:]...)
				if len(r.subscribers[user]) == 0 {
					delete(r.subscribers, user)
				}
				close(ch)
				log.Printf("[DEBUG] Cleaned up channel for user: %s", user)
				break
			}
		}
	}
}

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type queryResolver struct{ *Resolver }
type mutationResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

func (r *queryResolver) Messages(ctx context.Context) ([]*Message, error) {
	messages := []*Message{}

	// Clear any wrong type data (temporary fix)
	r.redis.Del(ctx, "messages")

	// Get message IDs from sorted set
	messageIDs, err := r.redis.ZRange(ctx, "messages", 0, -1).Result()
	if err != nil {
		if err != redis.Nil {
			return nil, err
		}
		// Return empty array if no messages
		return messages, nil
	}

	// Get each message
	for _, id := range messageIDs {
		messageJSON, err := r.redis.Get(ctx, "message:"+id).Result()
		if err != nil {
			if err != redis.Nil {
				continue
			}
			continue
		}

		var message Message
		if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
			continue
		}

		messages = append(messages, &message)
	}

	return messages, nil
}

func (r *queryResolver) Users(ctx context.Context) ([]string, error) {
	return []string{}, nil
}

func (r *queryResolver) Hello(ctx context.Context) (string, error) {
	return "Hello, World!", nil
}

func (r *mutationResolver) PostMessage(ctx context.Context, user string, text string) (*Message, error) {
	msg := &Message{
		ID:        ksuid.New().String(),
		User:      user,
		Text:      text,
		CreatedAt: Time{Time: time.Now()},
	}

	// Save to Redis
	messageJSON, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal message: %v", err)
		return nil, err
	}

	if err := r.redis.Set(ctx, "message:"+msg.ID, messageJSON, 0).Err(); err != nil {
		log.Printf("[ERROR] Failed to save message to Redis: %v", err)
		return nil, err
	}

	score := float64(msg.CreatedAt.Unix())
	if err := r.redis.ZAdd(ctx, "messages", &redis.Z{
		Score:  score,
		Member: msg.ID,
	}).Err(); err != nil {
		log.Printf("[ERROR] Failed to add message to sorted set: %v", err)
		return nil, err
	}

	log.Printf("[DEBUG] Message created - ID: %s, User: %s, Text: %s", msg.ID, msg.User, msg.Text)

	// Broadcast in the same goroutine
	r.broadcastMessage(msg)

	return msg, nil
}

func (r *subscriptionResolver) MessagePosted(ctx context.Context, user string) (<-chan *Message, error) {
	log.Printf("[DEBUG] New subscription request from user: %s", user)

	// Create buffered channel
	ch := make(chan *Message, 100)

	r.mutex.Lock()
	// Initialize subscribers map if needed
	if r.subscribers == nil {
		r.subscribers = make(map[string][]chan *Message)
	}

	// Add channel to subscribers
	r.subscribers["broadcast"] = append(r.subscribers["broadcast"], ch)
	currentCount := len(r.subscribers["broadcast"])
	r.mutex.Unlock()

	log.Printf("[DEBUG] Added subscription channel. Total channels now: %d", currentCount)

	// Handle cleanup when context is done
	go func() {
		<-ctx.Done()
		log.Printf("[DEBUG] Subscription context done for user: %s", user)

		r.mutex.Lock()
		defer r.mutex.Unlock()

		// Find and remove the channel
		channels := r.subscribers["broadcast"]
		for i, c := range channels {
			if c == ch {
				// Remove this channel
				r.subscribers["broadcast"] = append(channels[:i], channels[i+1:]...)
				close(ch)
				log.Printf("[DEBUG] Removed subscription channel. Remaining: %d", len(r.subscribers["broadcast"]))
				break
			}
		}
	}()

	return ch, nil
}

func (r *subscriptionResolver) UserJoined(ctx context.Context, user string) (<-chan string, error) {
	ch := make(chan string, 1)
	go func() {
		defer close(ch)
	}()
	return ch, nil
}
