<template>
  <div id="app">
    <header>
      <GitHubLogin />
    </header>
    <main>
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
#app {
  max-width: 800px;
  margin: 0 auto;
  padding: 2rem;
}

header {
  margin-bottom: 2rem;
}
</style>
