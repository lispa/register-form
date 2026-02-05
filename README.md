# Register Form (Go + MySQL)

A simple pet project: user registration form with frontend validation and backend API written in Go.

The project demonstrates a complete registration flow:
- HTML/CSS frontend
- JavaScript client-side validation
- Go backend API
- MySQL database
- Secure password storage (bcrypt)
- Environment-based configuration

---

## Features

- User registration with:
  - First name
  - Last name
  - Email
  - Password
- Client-side password validation:
  - Minimum 8 characters
  - At least one uppercase letter
  - At least one lowercase letter
  - At least one digit
  - Password confirmation
- Backend validation
- Password hashing using `bcrypt`
- MySQL persistence
- Configuration via `.env` (no secrets in code)
- Simple REST API (`/api/register`)

---

## Tech Stack

**Frontend**
- HTML5
- CSS3 (custom styles, no frameworks)
- Vanilla JavaScript

**Backend**
- Go
- net/http
- MySQL
- bcrypt

**Database**
- MySQL (via WAMP for local development)

---

## Project Structure

```
register-form/
├── api/
│   ├── go.mod
│   ├── go.sum
│   └── main.go
│
├── css/
│   └── style.css
│
├── index.html
├── .env
├── .gitignore
└── README.md
```

---

## Database Schema

```sql
CREATE TABLE users (
  id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  first_name VARCHAR(100) NOT NULL,
  last_name VARCHAR(100) NOT NULL,
  email VARCHAR(191) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uniq_email (email)
);
```

---

## Environment Variables

Create a `.env` file in the project root:

```
DB_USER=regapp
DB_PASS=your_password
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=userregistr
```

> ⚠️ The `.env` file is intentionally not committed to the repository.

---

## Running the Project

### 1. Start MySQL
Make sure MySQL is running (e.g. via WAMP).

### 2. Run the backend API

```bash
cd api
go run .
```

Expected output:
```
API listening on http://localhost:8080
```

### 3. Open the frontend
Open `index.html` in your browser.

---

## API

### `POST /api/register`

**Request body**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "email": "john@example.com",
  "password": "Aa123456"
}
```

**Success response**
```json
{
  "ok": true,
  "message": "registered"
}
```

**Error examples**
- `400` — validation error
- `409` — email already exists

---

## Security Notes

- Passwords are never stored in plain text.
- Passwords are hashed using `bcrypt`.
- Database credentials are stored in environment variables.
- Root MySQL user is not required (recommended to use a dedicated DB user).

---

## Project Status

✅ **Finished**

This project is intentionally kept minimal and complete as-is.  
No additional features are planned.

---

## Author

Dmitrii Novikov

Pet project created for learning purposes (Go backend + basic frontend).
