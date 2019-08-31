# App description

This app is based on [HighLoad Cup 2017 Task](https://github.com/MailRuChamps/hlcupdocs/blob/master/2017/TECHNICAL_TASK.md) (but the app isn't focused on highload)

But the app isn't focused on highload

It's a REST golang app that works with database.

ORM library: gorm
HTTP router: mux

Used database is SQLite, but it can be easy substituted for another type of database with which GORM can work. 


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


