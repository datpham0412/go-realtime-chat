<template>
  <div class="github-login">
    <div v-if="!isAuthenticated">
      <button @click="login" class="login-button">
        <i class="fab fa-github"></i> Sign in with GitHub
      </button>
    </div>
    <div v-else class="user-profile">
      <span class="username">Signed in with GitHub</span>
    </div>
  </div>
</template>

<script>
export default {
  name: 'GitHubLogin',
  data() {
    return {
      isAuthenticated: false
    }
  },
  methods: {
    login() {
      // Get base URL based on environment
      const baseURL = process.env.NODE_ENV === 'production'
        ? `${window.location.protocol}//${window.location.host}`
        : 'http://localhost:8080';
        
      window.location.href = `${baseURL}/auth/github`;
    }
  },
  mounted() {
    // Check if we have a user_id cookie
    this.isAuthenticated = document.cookie.includes('user_id')
  }
}
</script>

<style scoped>
.github-login {
  padding: 10px;
}

.login-button {
  background: #24292e;
  color: white;
  border: none;
  padding: 10px 20px;
  border-radius: 5px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 10px;
}

.username {
  font-weight: bold;
}
</style> 