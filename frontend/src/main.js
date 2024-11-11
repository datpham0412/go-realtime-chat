import Vue from 'vue';
import { ApolloClient } from 'apollo-client';
import { HttpLink } from 'apollo-link-http';
import { InMemoryCache } from 'apollo-cache-inmemory';
import VueApollo from 'vue-apollo';
import { split } from 'apollo-link';
import { WebSocketLink } from 'apollo-link-ws';
import { getMainDefinition } from 'apollo-utilities';
import 'bootstrap/scss/bootstrap.scss';

import router from './router';
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
    httpUri: '/graphql',
    wsUri: `wss://${window.location.host}/graphql`,
  },
};

// Get current environment configuration
const env = process.env.NODE_ENV || 'development';
const currentConfig = config[env];

// HTTP connection
const httpLink = new HttpLink({
  uri: currentConfig.httpUri,
});

// WebSocket connection
const wsLink = new WebSocketLink({
  uri: currentConfig.wsUri,
  options: {
    reconnect: true,
    timeout: 30000,
    connectionParams: {},
    connectionCallback: (error) => {
      if (error) {
        console.error('WebSocket connection error:', error);
      } else {
        console.log(`WebSocket connected to ${currentConfig.wsUri}`);
      }
    },
  },
});

// Log environment and connection details
console.log(`Running in ${env} mode`);
console.log(`HTTP endpoint: ${currentConfig.httpUri}`);
console.log(`WebSocket endpoint: ${currentConfig.wsUri}`);

const link = split(
  ({ query }) => {
    const { kind, operation } = getMainDefinition(query);
    return kind === 'OperationDefinition' && operation === 'subscription';
  },
  wsLink,
  httpLink,
);

const apolloClient = new ApolloClient({
  link: link,
  cache: new InMemoryCache(),
  defaultOptions: {
    watchQuery: {
      fetchPolicy: 'cache-and-network',
    },
  },
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
