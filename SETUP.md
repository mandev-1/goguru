# Camagru Setup Guide

## Prerequisites

- Go 1.21 or later
- Docker and Docker Compose (optional, for containerized deployment)

## Quick Start

### Option 1: Run with Docker Compose (Recommended)

```bash
docker-compose up --build
```

The application will be available at `http://localhost:8080`

MailHog (email testing) will be available at `http://localhost:8025`

### Option 2: Run Locally

1. Install dependencies:
```bash
go mod download
```

2. Build the application:
```bash
go build -o camagru .
```

3. Run the server:
```bash
./camagru
```

The application will be available at `http://localhost:8080`

## Environment Variables

The application supports environment variables via a `.env` file or system environment variables.

### Setup .env File

1. Copy the example file:
```bash
cp env.example .env
```

2. Edit `.env` with your configuration:
```bash
# Server Configuration
PORT=8080

# SMTP Configuration (for email sending)
SMTP_HOST=mailhog
SMTP_PORT=1025
SMTP_USER=
SMTP_PASS=
FROM_EMAIL=noreply@camagru.local
```

**Important**: The `.env` file is git-ignored. Never commit credentials to the repository.

### Available Variables

- `PORT` - Server port (default: 8080)
- `SMTP_HOST` - SMTP server host (default: mailhog)
- `SMTP_PORT` - SMTP server port (default: 1025)
- `SMTP_USER` - SMTP username (optional, for authenticated SMTP)
- `SMTP_PASS` - SMTP password (optional, for authenticated SMTP)
- `FROM_EMAIL` - From email address (default: noreply@camagru.local)

### Docker Compose

When using Docker Compose, environment variables are automatically loaded from `.env` file. System environment variables take precedence over `.env` file values.

## Features

- User registration with email verification
- Login/logout
- Password reset via email
- Image editor with webcam support
- Image superimposition (server-side processing)
- Public gallery with pagination
- Like and comment on images
- Email notifications for comments
- User profile management

## Database

The application uses SQLite database (`camagru.db`) which is created automatically on first run.

## Email Testing

When running with Docker Compose, MailHog is included for email testing. All emails will be captured and can be viewed at `http://localhost:8025`.

Without SMTP configuration, emails are logged to the console for development purposes.

