<template>
  <div class="chat-room">
    <div class="messages" ref="messages">
      <div v-for="msg in messages" :key="msg.id" 
           class="message" 
           :class="{ 'own-message': msg.user === currentUser }">
        <div class="message-content">
          <div class="message-header">
            <span class="username">{{ msg.user }}</span>
            <span class="time">{{ formatTime(msg.createdAt) }}</span>
          </div>
          <div class="text">{{ msg.text }}</div>
        </div>
      </div>
    </div>
    <div class="input-area">
      <input 
        v-model="newMessage" 
        @keyup.enter="sendMessage"
        placeholder="Type a message..."
        :disabled="!isAuthenticated"
      >
      <button @click="sendMessage" :disabled="!isAuthenticated">
        <i class="fas fa-paper-plane"></i>
      </button>
    </div>
  </div>
</template>

<script>
import { ApolloClient } from 'apollo-client'
import gql from 'graphql-tag'

const GET_MESSAGES = gql`
  query GetMessages {
    messages {
      id
      user
      text
      createdAt
    }
  }
`

const POST_MESSAGE = gql`
  mutation PostMessage($user: String!, $text: String!) {
    postMessage(user: $user, text: $text) {
      id
      user
      text
      createdAt
    }
  }
`

const MESSAGE_SUBSCRIPTION = gql`
  subscription OnMessagePosted($user: String!) {
    messagePosted(user: $user) {
      id
      user
      text
      createdAt
    }
  }
`

export default {
  name: 'ChatRoom',
  data() {
    return {
      messages: [],
      newMessage: '',
      isAuthenticated: false,
      currentUser: null,
      subscription: null,
      authCheckInterval: null
    }
  },
  methods: {
    formatTime(timestamp) {
      return new Date(timestamp).toLocaleTimeString()
    },
    async sendMessage() {
      if (!this.newMessage.trim() || !this.isAuthenticated) return

      try {
        await this.$apollo.mutate({
          mutation: POST_MESSAGE,
          variables: {
            user: this.currentUser,
            text: this.newMessage.trim()
          }
        })

        // Remove the local message addition since we'll get it from subscription
        this.newMessage = ''
      } catch (error) {
        console.error('Error sending message:', error)
      }
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const messages = this.$refs.messages
        if (messages) {
          messages.scrollTop = messages.scrollHeight
        }
      })
    },
    async fetchMessages() {
      try {
        const result = await this.$apollo.query({
          query: GET_MESSAGES,
          fetchPolicy: 'network-only' // Don't use cache
        })
        this.messages = result.data.messages
        this.scrollToBottom()
      } catch (error) {
        console.error('Error fetching messages:', error)
      }
    },
    subscribeToMessages() {
      console.log('[DEBUG] Starting subscription setup');
      console.log('[DEBUG] Current user:', this.currentUser);

      this.subscription = this.$apollo.subscribe({
        query: MESSAGE_SUBSCRIPTION,
        variables: {
          user: this.currentUser || ''
        }
      }).subscribe({
        next: ({ data }) => {
          console.log('[DEBUG] Received subscription data:', data);
          if (data && data.messagePosted) {
            const message = data.messagePosted;
            console.log('[DEBUG] Processing new message:', message);
            
            // Check if message already exists by ID
            const messageIndex = this.messages.findIndex(m => m.id === message.id);
            
            if (messageIndex === -1) {
              // Message doesn't exist, add it
              this.messages.push(message);
              this.scrollToBottom();
            }
          }
        },
        error: error => {
          console.error('[ERROR] Subscription error:', error);
          // Try to resubscribe after a delay
          setTimeout(() => {
            console.log('[DEBUG] Attempting to resubscribe...');
            if (this.subscription) {
              this.subscription.unsubscribe();
            }
            this.subscribeToMessages();
          }, 5000);
        },
        complete: () => {
          console.log('[DEBUG] Subscription completed');
        }
      });

      console.log('[DEBUG] Subscription setup completed');
    }
  },
  async mounted() {
    // Check authentication
    const checkAuth = () => {
      this.isAuthenticated = document.cookie.includes('user_id')
      if (this.isAuthenticated) {
        const userIdCookie = document.cookie
          .split('; ')
          .find(row => row.startsWith('user_id='))
        this.currentUser = userIdCookie ? userIdCookie.split('=')[1] : null
      } else {
        this.currentUser = null
      }
    }

    // Initial check
    checkAuth()

    // Set up an interval to check authentication status
    this.authCheckInterval = setInterval(checkAuth, 1000)

    // Fetch initial messages
    this.fetchMessages()

    // Subscribe to new messages
    if (this.isAuthenticated) {
      this.subscribeToMessages()
    }
  },
  beforeDestroy() {
    // Clean up subscription and interval
    if (this.subscription) {
      this.subscription.unsubscribe()
    }
    if (this.authCheckInterval) {
      clearInterval(this.authCheckInterval)
    }
  }
}
</script>

<style scoped>
.chat-room {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 150px);
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 2px 10px rgba(0,0,0,0.1);
  overflow: hidden;
}

.messages {
  flex-grow: 1;
  overflow-y: auto;
  padding: 1.5rem;
  background: #f8f9fa;
}

.messages::-webkit-scrollbar {
  width: 6px;
}

.messages::-webkit-scrollbar-thumb {
  background: #cbd5e0;
  border-radius: 3px;
}

.message {
  display: flex;
  margin-bottom: 1rem;
  align-items: start;
  animation: fadeIn 0.3s ease-in-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}

.message-content {
  max-width: 80%;
  background: white;
  padding: 0.8rem 1rem;
  border-radius: 12px;
  box-shadow: 0 1px 2px rgba(0,0,0,0.1);
}

.own-message {
  justify-content: flex-end;
}

.own-message .message-content {
  background: #0084ff;
  color: white;
}

.own-message .time {
  color: rgba(255,255,255,0.8);
}

.message-header {
  margin-bottom: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.username {
  font-weight: 600;
  font-size: 0.9rem;
}

.time {
  color: #666;
  font-size: 0.75rem;
}

.text {
  white-space: pre-wrap;
  line-height: 1.4;
}

.input-area {
  display: flex;
  padding: 1rem;
  background: white;
  border-top: 1px solid #edf2f7;
  gap: 0.5rem;
}

input {
  flex-grow: 1;
  padding: 0.8rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 24px;
  font-size: 0.95rem;
  transition: all 0.2s;
}

input:focus {
  outline: none;
  border-color: #0084ff;
  box-shadow: 0 0 0 2px rgba(0,132,255,0.2);
}

button {
  padding: 0.8rem;
  background: #0084ff;
  color: white;
  border: none;
  border-radius: 50%;
  cursor: pointer;
  width: 45px;
  height: 45px;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}

button:hover {
  background: #0073e6;
  transform: scale(1.05);
}

button:disabled {
  background: #cbd5e0;
  cursor: not-allowed;
  transform: none;
}
</style> 