<template>
  <div class="chat-room">
    <div class="messages" ref="messages">
      <div v-for="msg in messages" :key="msg.id" class="message">
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
      <button @click="sendMessage" :disabled="!isAuthenticated">Send</button>
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
      subscription: null
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
    this.isAuthenticated = document.cookie.includes('user_id')
    if (this.isAuthenticated) {
      const userIdCookie = document.cookie
        .split('; ')
        .find(row => row.startsWith('user_id='))
      this.currentUser = userIdCookie ? userIdCookie.split('=')[1] : null
    }

    // Fetch initial messages
    await this.fetchMessages()

    // Subscribe to new messages
    this.subscribeToMessages()
  },
  beforeDestroy() {
    // Clean up subscription
    if (this.subscription) {
      this.subscription.unsubscribe()
    }
  }
}
</script>

<style scoped>
.chat-room {
  display: flex;
  flex-direction: column;
  height: 500px;
  border: 1px solid #ccc;
  border-radius: 4px;
}

.messages {
  flex-grow: 1;
  overflow-y: auto;
  padding: 1rem;
}

.message {
  display: flex;
  margin-bottom: 1rem;
  align-items: start;
}

.message-content {
  flex-grow: 1;
}

.message-header {
  margin-bottom: 0.25rem;
}

.username {
  font-weight: bold;
  margin-right: 0.5rem;
}

.time {
  color: #666;
  font-size: 0.8rem;
}

.text {
  white-space: pre-wrap;
}

.input-area {
  display: flex;
  padding: 1rem;
  border-top: 1px solid #ccc;
}

input {
  flex-grow: 1;
  margin-right: 1rem;
  padding: 0.5rem;
  border: 1px solid #ccc;
  border-radius: 4px;
}

button {
  padding: 0.5rem 1rem;
  background: #007bff;
  color: white;
  border: none;
  border-radius: 4px;
  cursor: pointer;
}

button:disabled {
  background: #ccc;
  cursor: not-allowed;
}
</style> 