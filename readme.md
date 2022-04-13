# EDM System

EDM System is an electronic document management and task tracking server application.  
It is extremely easy to install and configure.

The application has the following functions:
* Documents: create, upload files, edit, delete
* User profiles, companies, departments: create, edit, delete
* Tasks: create, edit, upload files, assign, forward, change status (mark as done, cancel, etc.), add comments with files attached
* Notifications by e-mail: about user creation to that user, about changes in a task to related users
* Themes and localization support
* UX/UI features bb-code, search results highlighting, etc.
* Some basic bruteforce protection and other features

## How to build and run
To build use Go (Golang) programming language, run `go build`, and then you can run `./edm` app in the current directory. If you build the application and run locally, by default it immediately opens the browser, so you can start using it. Default login: admin, default password: no password.  
To build with docker and run with docker-compose use: `docker build -t edm .` and then `docker-compose up`. If you run it with docker-compose, you can open it at: http://127.0.0.1:8090

## Technical details
The application supports the following RDBMS:
* SQLite
* Microsoft SQL Server
* MySQL(MariaDB)
* Oracle
* PostgreSQL

Config file, logs, uploads, sqlite database are stored in `.edm` directory of user home directory.  
