# Forum Project

## Table of Contents

1. [Project Overview](#project-overview)
2. [Features](#features)
3. [Technologies Used](#technologies-used)
4. [Setup Instructions](#setup-instructions)
5. [Database Structure](#database-structure)
6. [User Authentication](#user-authentication)
7. [Error Handling and Testing](#error-handling-and-testing)
8. [Group Members](#group-members)

## Project Overview

This project involves creating a fully functional web forum in Go. The forum allows users to communicate, organize posts by category, and interact with content through likes, dislikes, and filtering options.

Our forum leverages **SQLite** as the database to handle and store user data, posts, and comments. The application is also containerized with Docker to ensure easy deployment and dependency management.

## Features

### Core Forum Functionalities

- **User Registration and Authentication**  
    - Register new users with email, username, and password.
    - Only registered users can post, comment, like, and dislike content.
    - Single session management per user via cookies with a set expiration time.

- **Content Organization and Interaction**
    - Users can create posts, associate posts with categories, and add comments.
    - Visible likes and dislikes for both posts and comments.

- **Filtering Options**
    - Filter posts by categories, user-created posts, and liked posts (available to registered users only).

### Additional Requirements

- **SQLite Database** for data management
- **Docker** for application containerization
- **HTTP and Error Handling** to manage web responses and website errors

## Technologies Used

- **Backend**: Go
- **Database**: SQLite
- **Containerization**: Docker
- **Languages**: HTML, SQL
- **Libraries**: sqlite3, bcrypt, UUID

## Setup Instructions

1. **Clone the repository**  
   ```bash
   git clone https://learn.reboot01.com/git/hasahmed/forum.git
   cd forum


## Build and Run with Docker

Ensure Docker is installed on your system, then use the Dockerfile to build and run the project:


docker build -t forum-app .
docker run -p 8080:8080 forum-app

## Database Structure

We use SQLite to manage user data, posts, comments, and categories. Key queries include:

    SELECT: Retrieve user posts or comments.
    CREATE: Create user, post, and comment tables.
    INSERT: Add new users, posts, comments, likes, and dislikes.

Consider structuring your tables with an Entity-Relationship Diagram to optimize performance.
User Authentication

    Registration: Users register with an email, username, and password.
        Password encryption: Stored passwords are encrypted (bonus feature).
        Session Management: Sessions are managed using cookies with a set expiration date.
    Login: Users can log in, provided they have correct credentials.

### Important Note

    Users must have unique emails; attempts to register with an existing email will return an error.
    Passwords are securely encrypted when stored.

### Error Handling and Testing

Our code includes:

    HTTP Status Codes: Responses are appropriately handled with HTTP status codes.
    Error Handling: All technical and logical errors are captured, with meaningful responses to guide users.
    Unit Testing: It is recommended to include unit test files for better code reliability.

## Group Members

   ### Captain:
   - Hashem Ahmed (GitHub: hasahmed)

   ### Members:

   - Hashem Ahmed (GitHub: hasahmed)
   - Faisal Almarzouqi (GitHub: falmarzo)
   - Zainab Alansari (GitHub: zalansari)
   - Mudhi BaniHammad (GitHub: mbaniham)
  -  Zainab Nasser (GitHub: zanasser)