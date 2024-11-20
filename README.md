# 💬 Real-time Chat with GraphQL Subscriptions in Go

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/datpham0412/go-realtime-chat)](https://goreportcard.com/report/github.com/datpham0412/go-realtime-chat)
[![GitHub issues](https://img.shields.io/github/issues/datpham0412/go-realtime-chat)](https://github.com/datpham0412/go-realtime-chat/issues)
[![GitHub stars](https://img.shields.io/github/stars/datpham0412/go-realtime-chat)](https://github.com/datpham0412/go-realtime-chat/stargazers)

## 📋 Project Description

A modern real-time chat application built with Go and Vue.js, featuring GraphQL subscriptions for live updates. The application demonstrates the implementation of real-time features using GraphQL subscriptions, Redis for message persistence, and GitHub OAuth for authentication. Perfect for developers looking to understand how to build real-time applications with modern web technologies.

## 🛠 Technologies Used

<p align="left">
    <a href="https://go.dev/" target="_blank" rel="noreferrer">
        <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/go/go-original.svg" alt="go" width="40" height="40"/>
    </a>
    <a href="https://vuejs.org/" target="_blank" rel="noreferrer">
        <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/vuejs/vuejs-original.svg" alt="vue" width="40" height="40"/>
    </a>
    <a href="https://redis.io/" target="_blank" rel="noreferrer">
        <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/redis/redis-original.svg" alt="redis" width="40" height="40"/>
    </a>
    <a href="https://graphql.org/" target="_blank" rel="noreferrer">
        <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/graphql/graphql-plain.svg" alt="graphql" width="40" height="40"/>
    </a>
    <a href="https://www.docker.com/" target="_blank" rel="noreferrer">
        <img src="https://raw.githubusercontent.com/devicons/devicon/master/icons/docker/docker-original.svg" alt="docker" width="40" height="40"/>
    </a>
</p>

- **Go**: Backend server implementation
- **Vue.js**: Frontend framework
- **GraphQL**: API query language with subscription support
- **Redis**: Message persistence and real-time features
- **Docker**: Containerization and deployment
- **GitHub OAuth**: User authentication
- **WebSocket**: Real-time communication
- **gqlgen**: GraphQL server library for Go

## 📚 Features

- Real-time message updates using GraphQL subscriptions
- Persistent chat history with Redis
- GitHub OAuth authentication
- Modern, responsive UI with Vue.js
- Docker containerization for easy deployment
- GraphQL playground for API testing
- Message timestamps and user identification
- Cross-platform compatibility

## 🚀 Installation and Running the Project

### Prerequisites

- Docker and Docker Compose
- Go 1.22 or later
- Node.js 18 or later
- GitHub OAuth credentials

### Development Setup

1. **Clone the Repository**:

```bash
git clone https://github.com/datpham0412/go-realtime-chat.git
cd go-realtime-chat
```

2. **Environment Configuration**:
   Create a `.env` file with:

```env
REDIS_URL=redis://redis:6379
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
```

3. **Start Development Services**:

```bash
docker-compose up -d
```

4. **Run Frontend Development Server**:

```bash
cd frontend
npm install
npm run start
```

### Production Deployment with Fly.io

1. **Install Fly CLI**:
   Follow instructions at [fly.io/docs/hands-on/install-flyctl](https://fly.io/docs/hands-on/install-flyctl/)

2. **Login to Fly.io**:

```bash
fly auth login
```

3. **Deploy the Application**:

```bash
fly deploy
```

4. **Set Environment Variables**:

```bash
fly secrets set GITHUB_CLIENT_ID=your_client_id
fly secrets set GITHUB_CLIENT_SECRET=your_client_secret
```

## 📷 Screenshots

[Add your application screenshots here]

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 📞 Contact

For any inquiries, please open an issue in the GitHub repository.

Made with ❤️ by [Your Name](https://github.com/datpham0412)
