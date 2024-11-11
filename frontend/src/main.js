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

const httpLink = new HttpLink({
  uri: process.env.NODE_ENV === 'production' ? '/graphql' : 'http://localhost:8080/graphql',
});

// Improved WebSocket configuration
const wsLink = new WebSocketLink({
  uri: process.env.NODE_ENV === 'production' ? `wss://${window.location.host}/graphql` : 'ws://localhost:8080/graphql',
  options: {
    reconnect: true,
    reconnectionAttempts: 5,
    timeout: 30000,
    connectionParams: {
      // Add any auth tokens if needed
    },
    connectionCallback: (error) => {
      if (error) {
        console.error('WebSocket connection error:', error);
      } else {
        console.log('WebSocket connected successfully');
      }
    },
    inactivityTimeout: 30000,
    lazy: false, // Connect immediately instead of waiting for first subscription
  },
});

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
  cache: new InMemoryCache({
    addTypename: true,
    typePolicies: {
      Query: {
        fields: {
          messages: {
            merge(existing = [], incoming) {
              return [...incoming];
            },
          },
        },
      },
    },
  }),
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
