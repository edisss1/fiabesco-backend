Fiabesco is a backend service for a social media platform. It allows users to interact with each other through posts, comments, and likes. The backend is built using **Go** and **GoFiber**, with **MongoDB** as the database.

## Features

-   User registration and authentication
    
-   Post creation
    
-   Comment and like systems for posts
 
-   MongoDB for data storage
    

    

## Installation

### Prerequisites

-   [Go 1.23+](https://golang.org/dl/)
    
-   [MongoDB](https://www.mongodb.com/try/download/community)
    

### Clone the repository

```bash
git clone https://github.com/edisss1/fiabesco-backend.git
cd fiabesco-backend
```



### Install dependencies

Run the following command to install the required Go dependencies:



```bash 
go mod tidy
```



### Set up environment variables

Create a `.env` file in the root directory of the project and set the following environment variables:





`MONGO_URI=mongodb+srv://<db_user>:<db_password>@cluster0.ez6e1rn.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0` 
`PORT=3000`

-   `MONGO_URI` should be your MongoDB connection string.
    
-   `PORT` is the port on which the backend will run.
    

### Run the project

To start the server, use the following command:





`go run cmd/main.go` 

Or add this to Makefile
```Makefile
run:
	go run cmd/main.go
```

The backend will be running on `http://localhost:3000`.
