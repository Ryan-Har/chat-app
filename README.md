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
helm pull oci://registry-1.docker.io/pandects/chat-app-helm --untar
helm install chat-app chat-app-helm/ -f chat-app-helm/values.yaml
```

The user interface can be accessed from http://localhost:30080 <br>
The app/admin interface can be accessed from http://localhost:30005 <br>
The lavinmq management interface can be accessed from http://localhost:30672/login