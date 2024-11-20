<template>
  <div class="github-login">
    <div v-if="!isAuthenticated">
      <button @click="login" class="login-button">
        <i class="fab fa-github"></i> Sign in with GitHub
      </button>
    </div>
    <div v-else class="user-profile">
      <span class="username">Signed in with GitHub</span>
      <button @click="logout" class="logout-button">Sign Out</button>
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
      const baseURL = process.env.NODE_ENV === 'production'
        ? `${window.location.protocol}//${window.location.host}`
        : 'http://localhost:8080';
        
      window.location.href = `${baseURL}/auth/github`;
    },
    logout() {
      document.cookie = 'user_id=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/;';
      this.isAuthenticated = false;
      window.location.reload();
    }
  },
  mounted() {
    this.isAuthenticated = document.cookie.includes('user_id')
  }
}
</script>

<style scoped>
.github-login {
  padding: 0.5rem;
}

.login-button {
  background: #24292e;
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 24px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.95rem;
  transition: all 0.2s;
}

.login-button:hover {
  background: #2c3238;
  transform: translateY(-1px);
}

.user-profile {
  display: flex;
  align-items: center;
  gap: 1rem;
}

.username {
  font-weight: 500;
  color: #1a1a1a;
}

.logout-button {
  background: #f0f2f5;
  color: #1a1a1a;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 24px;
  cursor: pointer;
  font-size: 0.9rem;
  transition: all 0.2s;
}

.logout-button:hover {
  background: #e4e6eb;
}
</style> 