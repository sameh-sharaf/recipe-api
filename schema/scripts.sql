-- Build schema
CREATE SCHEMA app;

-- Recipe table and sequence
CREATE SEQUENCE app.recipes_id_seq;

CREATE TABLE IF NOT EXISTS app.recipes (
  id            INT           PRIMARY KEY   DEFAULT NEXTVAL('app.recipes_id_seq'),
  name          VARCHAR(128)  NOT NULL,
  prep_time     SMALLINT      NOT NULL,
  difficulty    SMALLINT      NOT NULL,
  vegeterian    BOOLEAN       NOT NULL,
  createdat     TIMESTAMP     NOT NULL,
  updatedat     TIMESTAMP     NOT NULL
);

ALTER SEQUENCE app.recipes_id_seq OWNED BY app.recipes.id;

-- Recipes rates and sequence
CREATE SEQUENCE app.rates_id_seq;

CREATE TABLE IF NOT EXISTS app.rates (
  id        INT         PRIMARY KEY   DEFAULT NEXTVAL('app.rates_id_seq'),
  recipeID  INT         NOT NULL,
  rate      SMALLINT    NOT NULL,
  createdat TIMESTAMP   NOT NULL,
  FOREIGN KEY (recipeID) REFERENCES app.recipes(id)
);

ALTER SEQUENCE app.rates_id_seq OWNED BY app.rates.id;

-- Users table
CREATE TABLE IF NOT EXISTS app.users (
 id             SERIAL        PRIMARY KEY,
 username       VARCHAR(128)  NOT NULL,
 fullName       VARCHAR(128)  NOT NULL,
 passwordHash   VARCHAR(128)  NOT NULL,
 isDisabled     bool          DEFAULT FALSE,
 createdat      TIMESTAMP     NOT NULL
);

-- User session
CREATE TABLE IF NOT EXISTS app.userSessions (
 sessionKey   VARCHAR(128)  PRIMARY KEY,
 userID       INT           NOT NULL,
 LoginTime    TIMESTAMP     NOT NULL,
 FOREIGN KEY (userID) REFERENCES app.users(id)
);
