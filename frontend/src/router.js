import Vue from 'vue';
import Router from 'vue-router';
import Home from '@/components/Home.vue';
import Login from '@/components/Login.vue';
import ChatRoom from '@/components/ChatRoom.vue';

Vue.use(Router);

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'home',
      component: ChatRoom,
      meta: { requiresAuth: true }
    },
    {
      path: '/login',
      name: 'login',
      component: Login
    }
  ]
});
