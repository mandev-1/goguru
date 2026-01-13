# 📷 Camagru

A web application for creating and sharing edited photos using your webcam and superposable images. Built with Go on the server-side and vanilla HTML, CSS, and JavaScript on the client-side.

## ✨ Features

### User Management
- **User Registration**: Sign up with username, email, and password
  - Email verification required before account activation
  - Password complexity validation (minimum 8 characters)
  - Username format validation (alphanumeric + underscores, 3-20 characters)
- **Authentication**: Secure login with session management
- **Password Reset**: Forgot password functionality with email-based reset
- **Profile Management**: Update username, email, and password
- **User Preferences**: Toggle email notifications for comments and likes

### Image Editing
- **Webcam Capture**: Take photos directly from your webcam
- **Image Upload**: Upload your own images if webcam is unavailable
- **Image Superposition**: Overlay predefined assets (frames, stickers, etc.) on your photos
- **Server-Side Processing**: All image composition handled securely on the server
- **Image Gallery**: View all your created images in a personal gallery

### Public Gallery
- **Browse All Images**: Public gallery displaying all user-created images
- **Pagination**: Navigate through images with paginated results (12 per page)
- **Like Images**: Authenticated users can like images
- **Comment System**: Add comments on images (authentication required)
- **Email Notifications**: Get notified when someone likes or comments on your images

### Security & Best Practices
- **No Console Logs**: Zero errors, warnings, or logs in console (except getUserMedia)
- **Form Validation**: Comprehensive client and server-side validation
- **Session Management**: Secure cookie-based sessions
- **XSS Protection**: HTML escaping for user-generated content
- **SQL Injection Prevention**: Parameterized queries throughout
- **CSRF Protection**: Session-based protection for form submissions

## 🏗️ Architecture

The application follows a clean, feature-based internal structure:

```
camagru/
├── internal/
│   ├── auth/          # Authentication utilities (hashing, tokens, validation)
│   ├── database/      # Database initialization and schema
│   ├── models/        # Data models (User, Image, Comment, etc.)
│   └── server/        # HTTP handlers organized by feature
│       ├── auth.go    # Authentication handlers
│       ├── email.go   # Email sending functionality
│       ├── handlers.go # Page and API handlers
│       ├── image.go   # Image composition and processing
│       ├── routes.go  # Route definitions
│       └── server.go   # Core server type and middleware
├── data/
│   ├── camagru.db     # SQLite database
│   └── uploads/       # User-uploaded images
├── web/
│   └── static/        # Frontend assets
│       ├── assets/    # Superposable images (overlays)
│       ├── pages/     # HTML pages
│       └── *.js       # Client-side JavaScript
└── main.go            # Application entry point
```

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** (for local development)
- **Docker & Docker Compose** (for containerized deployment)

### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone <repository-url>
cd goguru-2

# Start the application
docker-compose up --build
```

The application will be available at:
- **Application**: http://localhost:8080
- **MailHog** (Email Testing): http://localhost:8025

### Option 2: Local Development

```bash
# Install dependencies
go mod download

# Build the application
go build -o camagru .

# Run the server
./camagru
```

The application will start on `http://localhost:8080` by default.

## ⚙️ Configuration

### Environment Variables

The application supports environment variables via a `.env` file or system environment variables.

#### Setup .env File

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

#### Available Variables

- `PORT` - Server port (default: 8080)
- `SMTP_HOST` - SMTP server host (default: mailhog)
- `SMTP_PORT` - SMTP server port (default: 1025)
- `SMTP_USER` - SMTP username (optional, for authenticated SMTP)
- `SMTP_PASS` - SMTP password (optional)
- `FROM_EMAIL` - From email address (default: noreply@camagru.local)

#### Docker Compose

When using Docker Compose, environment variables are automatically loaded from `.env` file. System environment variables take precedence over `.env` file values.

### Database

The application uses SQLite3, which is automatically initialized on first run. The database file is stored in `data/camagru.db`.

### Email Configuration

For development, MailHog is included in the Docker Compose setup to capture and view all emails. In production, configure your SMTP settings via environment variables.

## 📖 Usage

### Registration

1. Navigate to `/register`
2. Fill in username (3-20 characters, alphanumeric + underscores)
3. Enter a valid email address
4. Create a password (minimum 8 characters)
5. Confirm your password
6. Check your email for verification link
7. Click the verification link to activate your account

### Creating Images

1. Log in to your account
2. Navigate to `/editor`
3. Select a superposable image (overlay) from the list
4. Either:
   - Click "Capture" to take a photo with your webcam, or
   - Click "Upload" to upload an image file
5. Adjust the overlay position and size (optional)
6. Click "Save" to create your image
7. Your image will appear in the gallery

### Gallery

- **Public Access**: Anyone can view the gallery at `/gallery`
- **Interactions**: Logged-in users can like and comment on images
- **Pagination**: Navigate through pages using the pagination controls
- **Notifications**: Receive email notifications when your images are liked or commented on

## 🛠️ Technology Stack

### Backend
- **Go 1.21**: Server-side language
- **SQLite3**: Database (`github.com/mattn/go-sqlite3`)
- **bcrypt**: Password hashing (`golang.org/x/crypto`)
- **Standard Library**: Image processing, HTTP server, email

### Frontend
- **HTML5**: Semantic markup
- **CSS3**: Styling with modern features (flexbox, grid, glassmorphism)
- **Vanilla JavaScript**: No frameworks, pure browser APIs
- **Webcam API**: `getUserMedia()` for camera access
- **Canvas API**: Image manipulation and preview

### Infrastructure
- **Docker**: Containerization
- **Docker Compose**: Multi-container orchestration
- **MailHog**: Email testing and development

## 📁 Project Structure

```
.
├── internal/              # Internal packages (not importable)
│   ├── auth/            # Authentication utilities
│   ├── database/        # Database schema and initialization
│   ├── models/          # Data models
│   └── server/          # HTTP handlers and server logic
├── data/                # Application data
│   ├── camagru.db       # SQLite database
│   └── uploads/        # User-uploaded images
├── web/                 # Frontend assets
│   └── static/
│       ├── assets/      # Superposable images
│       ├── pages/       # HTML pages
│       └── *.js         # JavaScript modules
├── main.go              # Application entry point
├── Dockerfile           # Container build instructions
├── docker-compose.yml   # Docker Compose configuration
├── go.mod               # Go module definition
└── README.md            # This file
```

## 🔌 API Endpoints

### Authentication
- `POST /register` - User registration
- `POST /login` - User login
- `POST /logout` - User logout
- `GET /verify?token=...` - Email verification
- `POST /forgot-password` - Request password reset
- `POST /reset-password` - Reset password with token
- `POST /resend-verification` - Resend verification email

### User Management
- `GET /api/current-user` - Get current user info
- `POST /api/user/update` - Update user profile
- `GET /api/user/preferences` - Get user preferences
- `POST /api/user/preferences` - Update user preferences
- `GET /api/user/images` - Get user's images

### Image Operations
- `GET /api/assets` - Get list of superposable images
- `POST /api/compose` - Create new image (requires auth)
- `GET /api/gallery` - Get gallery images (paginated)
- `POST /api/gallery/like` - Like an image (requires auth)
- `POST /api/gallery/comment` - Comment on image (requires auth)
- `POST /api/gallery/delete` - Delete own image (requires auth)

## 🎨 Features in Detail

### Responsive Design
- Mobile-first approach
- Glassmorphism effects on mobile
- Touch-friendly interface
- Adaptive navigation

### Image Processing
- Server-side image composition using Go's `image` package
- Alpha channel blending for overlays
- JPEG encoding with quality control
- Automatic resizing and positioning

### Security Features
- Password hashing with bcrypt
- Session token management
- CSRF protection via session cookies
- XSS prevention through HTML escaping
- SQL injection prevention with parameterized queries

## 🧪 Development

### Running Tests

```bash
# Run Go tests (if any)
go test ./...
```

### Building for Production

```bash
# Build binary
CGO_ENABLED=1 go build -o camagru .

# Or with Docker
docker build -t camagru .
```

### Code Structure

The codebase follows Go best practices:
- Feature-based organization in `internal/`
- Clear separation of concerns
- No external dependencies beyond standard library equivalents
- Comprehensive error handling (without logging)

## 📝 Notes

- **No Logging**: The application follows strict requirements with no console logs, warnings, or errors (except `getUserMedia` related)
- **Client-Side Only**: HTML, CSS, and vanilla JavaScript - no frameworks
- **Server Flexibility**: Go implementation with PHP standard library equivalents
- **Containerization**: Full Docker support for easy deployment

## 📄 License

© 2025 MMAN

## 🙏 Acknowledgments

Built as a learning project focusing on:
- Responsive web design
- DOM manipulation
- SQL debugging
- Cross-Site Request Forgery (CSRF)
- Cross-Origin Resource Sharing (CORS)
- Server-side image processing
- Email verification systems
- Session management

---

**Enjoy creating and sharing your edited photos! 📸**
