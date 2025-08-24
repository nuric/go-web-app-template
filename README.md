# go-api-template

This is a template for creating server-side rendered web applications using Go. It is designed to be simple and minimal with no frameworks. It follows as model-view-controller approach.

- Standard library HTTP server with routing
- Middleware using HTTP handlers including recovery and logging, rate limiting
- Handling of forms with simple explicit validation
- Sessions, login and user management with reset tokens, email verification
- CSRF protection, password hashing and password reset
- Flash messages similar to Django
- File uploads with progress tracking
- Integration tests using chromedp

**Why?** When I start projects, I often have to scaffold a lot of boilerplate code. People argue that's what frameworks are for, but often I need something that's customised down the line. The goal of this template is to provide that initial start with minimal framework overhead.

The only major dependencies I include are for security such as managing user sessions and protecting against CSRF attacks. The rest I try to keep to a minimum.

## Getting Started

You can use this template to create a new Go project. Select use this template in Github to get started. To run the server:

```bash
go run ./
```

or using Docker:

```bash
docker build -t go-api-template .
docker run -p 8080:8080 go-api-template
```

Then navigate to [http://localhost:8080](http://localhost:8080) to get redirected to the login page.

## Structure

The project is organized to keep concerns separated and code maintainable. Below is the top-level structure:

```text
├── auth/           # Authentication logic (login, signup, password reset)
├── controllers/    # HTTP handlers for different pages and actions
├── email/          # Email sending utilities
├── middleware/     # Custom HTTP middleware (rate limiting, error handling)
├── models/         # Data models (e.g., User)
├── static/         # Static assets (CSS, images)
├── templates/      # HTML templates for rendering views
│   ├── components/ # Reusable template components
│   ├── email/      # Email templates
│   ├── layouts/    # Layout templates
│   └── pages/      # Page templates
├── tests/          # Integration and unit tests
├── utils/          # Utility functions (encoding, password hashing)
├── main.go         # Application entry point
```

This structure helps you quickly locate code for authentication, page controllers, templates, and utilities. It is designed for extensibility and clarity, making it easy to add new features or pages.

## Built with

- [env](https://github.com/caarlos0/env) for environment variable based configuration
- [gorilla/csrf](https://github.com/gorilla/csrf) for CSRF protection
- [gorilla/sessions](https://github.com/gorilla/sessions) for session management
- [gorilla/schema](https://github.com/gorilla/schema) for decoding form values into structs
- [lmittmann/tint](https://github.com/lmittmann/tint) for coloured logging
- [gorm](https://github.com/go-gorm/gorm) for ORM and database interactions
