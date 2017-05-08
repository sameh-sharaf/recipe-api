# recipe-api

Language: Golang
Database: Postgresql

This repo is a sample of web server written in Golang 1.7. This is the list of endpoints it can support:


| Name   | Method      | URL                  | Protected |
| ---    | ---         | ---                  | ---       |
| List   | `GET`       | `/recipes`           | ✘         |
| Create | `POST`      | `/recipes`           | ✓         |
| Get    | `GET`       | `/recipes/{id}`      | ✘         |
| Update | `PUT/PATCH` | `/recipes/{id}`      | ✓         |
| Delete | `DELETE`    | `/recipes/{id}`      | ✓         |
| Rate   | `PUT/PATCH` | `/recipes/{id}/rate` | ✘         |
| Search | `GET`       | `/search`            | ✘         |


Directories & Files:
- `schema/:` Contains necessary SQL scripts (schema, tables, sequences) for web server to work.
- `src/`: Contains .go files with `main_test.go` for unit testing.
- `Makefile`: To build web server bin file.
- `example.env`: Contains all necessary environment variables.
- `searchTemplate.json`: Contains some search JSON examples.

Authentication:
This PR provides three endpoints for user control:
/register: To create a new user.
- username
- password
- fullname
/login: To open a new session to access web server and use protected endpoints. It will create a session key (token) written in a
cookie with max age can be configured by COOKIE_MAX_AGE
- username
- password
/logout: To end a session.

BCrypt is used for password hashing.

How to build the web server docker container:

Simply by running:
`docker-compose build`
`docker-compose up -d`

Search endpoint
Search takes the URL parameter `query` which is JSON. The structure of JSON object is as follows:

- Groups:
This represents the filters which are grouped with AND relationship.

- Filters:
Each filter represents a condition which it filters recipes based on one of the columns (types): Name, difficulty, prep time or avg. rating.
A filter can be two types:
- String filter: (name)
It has the following attributes:
-- type: which column/field to filter based on (here we have only recipe name)
-- operation: String filters have the following operations:
--- 'match' or '='
--- 'start'
--- 'end'
--- 'contain'
-- value: the value which filter will be based on.
-- case_sensitive (bool): Can be case sensitive or insensitive

- Numeric filter: (difficulty, prep_time, rating)
It has the following attributes:
-- type: We have difficulty, prep_time & avg. rating (rating).
-- operation: Numeric filters have the the known mathematic operations (==, >, <, !=, >=, <=) for comparison.
-- value: the value which filter will be based on.
```

Ex1; Search for recipes whose name contains 'lasagna' and difficulty is medium (2) or lower case insensitive
```
{
  "groups": [
    {
      "filters": [
        {
          "type": "name",
          "operation": "contain",
          "value": "lasagna",
          "case_sensitive": false
        },
        {
          "type": "difficulty",
          "operation": "<=",
          "value": "2"
        }
      ]
    }
  ]
}
```
Note that we have one group with two filters. Search engine will consider it an 'AND' relationship.

Ex2: Search for recipes whose prep time is between 30 and 60 minutes.
```
{
  "groups": [
    {
      "filters": [
        {
          "type": "prep_time",
          "operation": ">=",
          "value": "1800"
        },
        {
          "type": "prep_time",
          "operation": "<=",
          "value": "3600"
        }
      ]
    }
  ]
}
```
Ex2: Search recipes whose name either starts with 'Grilled' or ends with 'Pork' case sensitive.
{
  "groups": [
    {
      "filters": [
        {
          "type": "name",
          "operation": "start",
          "value": "Grilled",
          "case_sensitive": true
        }
      ]
    },
    {
      "filters": [
        {
          "type": "name",
          "operation": "end",
          "value": "Pork",
          "case_sensitive": true
        }
      ]
    }
  ]
}
Note we created two groups: First group is for first condition (starts with 'Grilled') and the other for the condition (ends with 'Pork').
This will consider groups as 'OR' relationship.
