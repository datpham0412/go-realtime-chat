<template>
  <div>
    <app-message v-for="message of messages"
                 :key="message.id"
                 :message="message">
    </app-message>
  </div>
</template>

<script>
import gql from 'graphql-tag';
import Message from '@/components/Message';

const MESSAGE_SUBSCRIPTION = gql`
  subscription OnMessagePosted($user: String!) {
    messagePosted(user: $user) {
      id
      user
      text
      createdAt
    }
  }
`;

const GET_MESSAGES = gql`
  query GetMessages {
    messages {
      id
      user
      text
      createdAt
    }
  }
`;

export default {
  components: {
    'app-message': Message,
  },
  data() {
    return {
      messages: [],
      subscriptionObserver: null,
    };
  },
  apollo: {
    messages: {
      query: GET_MESSAGES,
    },
  },
  methods: {
    setupSubscription() {
        console.log('[DEBUG] Setting up subscription...');
        const user = this.$currentUser();
        
        if (this.subscriptionObserver) {
            console.log('[DEBUG] Cleaning up existing subscription');
            this.subscriptionObserver.unsubscribe();
        }

        console.log('[DEBUG] Creating new subscription for user:', user);
        
        // Create the subscription
        this.subscriptionObserver = this.$apollo.subscribe({
            query: MESSAGE_SUBSCRIPTION,
            variables: {
                user: user,
            },
        }).subscribe({
            next: ({ data }) => {
                console.log('[DEBUG] Received subscription data:', data);
                if (data && data.messagePosted) {
                    const newMessage = data.messagePosted;
                    console.log('[DEBUG] New message received:', newMessage);
                    
                    // Update messages array
                    this.messages = [newMessage, ...this.messages];
                    console.log('[DEBUG] Messages updated, new count:', this.messages.length);
                }
            },
            error: (error) => {
                console.error('[ERROR] Subscription error:', error);
                // Attempt to reconnect
                setTimeout(() => {
                    console.log('[DEBUG] Attempting to reconnect subscription...');
                    this.setupSubscription();
                }, 3000);
            },
        });
    },
  },
  created() {
    console.log('[DEBUG] MessageList component created');
    this.setupSubscription();
  },
  beforeDestroy() {
    console.log('[DEBUG] MessageList component being destroyed');
    if (this.subscriptionObserver) {
        this.subscriptionObserver.unsubscribe();
    }
  },
};
</script>
