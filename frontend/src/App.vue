<template>
  <div id="app">
    <header class="app-header">
      <div class="header-content">
        <h1>Chat App</h1>
        <GitHubLogin />
      </div>
    </header>
    <main class="app-main">
      <ChatRoom />
    </main>
  </div>
</template>

<script>
import GitHubLogin from './components/GitHubLogin.vue'
import ChatRoom from './components/ChatRoom.vue'

export default {
  name: 'App',
  components: {
    GitHubLogin,
    ChatRoom
  }
}
</script>

<style>
body {
  margin: 0;
  padding: 0;
  background: #f0f2f5;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
}

#app {
  max-width: 1200px;
  margin: 0 auto;
  padding: 1rem;
}

.app-header {
  background: white;
  padding: 1rem;
  border-radius: 12px;
  margin-bottom: 1rem;
  box-shadow: 0 1px 3px rgba(0,0,0,0.1);
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-content h1 {
  font-size: 1.5rem;
  color: #1a1a1a;
  margin: 0;
}

.app-main {
  background: transparent;
}
</style>
