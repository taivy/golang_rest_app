# App description

This app is based on [HighLoad Cup 2017 Task](https://github.com/MailRuChamps/hlcupdocs/blob/master/2017/TECHNICAL_TASK.md) (but the app isn't focused on highload)

But the app isn't focused on highload

It's a REST golang app that works with database.

ORM library: gorm
HTTP router: mux

Used database is SQLite, but it can be easy substituted for another type of database with which GORM can work. 


# Requirements

go >= 1.10


# Install and run locally

```
git clone https://github.com/taivy/golang_rest_app
cd golang_rest_app
go get -d -v
go run main.go
```


# Deploy with Docker
Go to repo directory in Docker shell and run:

```
docker build -t rest_app:latest . -f Dockerfile.multi
docker run -p 8000:8000 --name rest_app --rm rest_app 
```

Press Ctrl+C to detach from container.

App is running on port 8000 of your docker machine. You can make requests with curl, Postman etc. For example, curl:
```
curl 192.168.99.100:8080/users/1 -X GET
```
where 192.168.99.100 is docker machine address

You can view files in container with sh:
```
docker exec -it rest_app /bin/sh  
```

And copy files from container to host machine:
```
docker cp container_id:/container/file/path /host/file/path
```

For example:
```
docker cp rest_app:/app/log.log C:\\Users\\User1\\go_rest_files
docker cp rest_app:/app/data.db C:\\Users\\User1\\go_rest_files
```
to copy logs and database.

# Run tests
Go to repo directory and run
`go test`

# Entities
- users
- locations
- visits

# Entities' Models

## User
- id
- email
- first_name
- last_name
- gender - "f" for female or "m" for male
- birth_date - timestamp

## Location
- id
- place - location description
- country
- city
- distance - distance from city in km

## Visit
- id
- location - id of visit location
- user - id of user who made visit
- visited_at - timestamp
- mark - 0 to 5

# Endpoints:

## GET

### `/<entity>/<id>` - get info about entity

### `/users/<id>/visits` - get list of places user has visited

### `/locations/<id>/avg` - get average location mark
Get parameters:
- fromAge - consider marks only from users with age more than specified in parameter
- toAge - consider marks only from users with age less than specified in parameter
- gender - consider marks only from users with specified gender (m or f)
- fromDate - consider marks only from visits with date more than specified in parameter
- toDate - consider marks only from visits with date less than specified in parameter


## POST

### `/<entity>/<id>`
Update info about entity. New values for fields are specified in JSON body.

### `/<entity>/new`
Create new entity. All fields (from entities' models) are required. Fields are specified in JSON body.


