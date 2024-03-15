# Chat App
The app is still in development but is functional in its current state

## Overview
This is a simple chat application written in Golang, designed for real-time communication. The application is orchestrated using Kubernetes for easy deployment, scaling, and management.

## Features

- Real-time chat functionality.
- Golang backend for efficiency and performance.
- Kubernetes orchestration for scalability and reliability.

## Requirements

- Locally hosted kubernetes cluster
- Helm installed

## Getting Started

```
git clone https://github.com/Ryan-Har/chat-app
cd chat app
helm install chat-app helm/ -f helm/values.yaml
```

The user interface can be accessed from http://localhost:30080 <br>
The app/admin interface can be accessed from http://localhost:30005 <br>
The lavinmq management interface can be accessed from http://localhost:30672/login