import Vue from 'vue';
import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { InMemoryCache } from 'apollo-cache-inmemory';
import VueApollo from 'vue-apollo';
import { split } from 'apollo-link';
import { WebSocketLink } from 'apollo-link-ws';
import { getMainDefinition } from 'apollo-utilities';
import 'bootstrap/scss/bootstrap.scss';

import router from './router-guard';
import App from './App.vue';
import { AuthPlugin } from './auth';

Vue.config.productionTip = false;

// Configuration based on environment
const config = {
  development: {
    httpUri: 'http://localhost:8080/graphql',
    wsUri: 'ws://localhost:8080/graphql',
  },
  production: {
    httpUri: `${window.location.protocol}//${window.location.host}/graphql`,
    wsUri: `wss://${window.location.host}/graphql`,
  },
};

// Get current environment configuration
const env = process.env.NODE_ENV || 'development';
const currentConfig = config[env];

// HTTP connection
const httpLink = new HttpLink({
  uri: currentConfig.httpUri,
  credentials: 'include'
});

// WebSocket connection
const wsLink = new WebSocketLink({
  uri: currentConfig.wsUri,
  options: {
    reconnect: true,
    connectionParams: () => {
      console.log('Setting up WebSocket connection params');
      return {};
    },
    timeout: 30000,
    reconnectionAttempts: 5,
    lazy: false,
    inactivityTimeout: 30000,
    onError: (error) => {
      console.error('WebSocket error:', error);
    },
    connectionCallback: (error) => {
      if (error) {
        console.error('WebSocket connection error:', error);
      } else {
        console.log('WebSocket connected successfully');
      }
    }
  }
});

// Log environment and connection details
console.log(`Running in ${env} mode`);
console.log(`HTTP endpoint: ${currentConfig.httpUri}`);
console.log(`WebSocket endpoint: ${currentConfig.wsUri}`);

const link = split(
  ({ query }) => {
    try {
      const definition = getMainDefinition(query);
      const isSubscription = 
        definition.kind === 'OperationDefinition' &&
        definition.operation === 'subscription';
      console.log('Operation type:', definition.operation, 'Using WebSocket:', isSubscription);
      return isSubscription;
    } catch (error) {
      console.error('Error in split:', error);
      return false;
    }
  },
  wsLink,
  httpLink
);

const apolloClient = new ApolloClient({
  link,
  cache: new InMemoryCache()
});

const apolloProvider = new VueApollo({
  defaultClient: apolloClient,
  errorHandler(error) {
    console.error('Apollo error:', error);
  },
});

Vue.use(VueApollo);
Vue.use(AuthPlugin);

const vm = new Vue({
  router,
  provide: apolloProvider.provide(),
  render: (h) => h(App),
});
vm.$mount('#app');
