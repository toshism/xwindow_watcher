CREATE TABLE IF NOT EXISTS apps (
       id integer PRIMARY KEY,
       name text NOT NULL,
       created_at datetime
);


CREATE TABLE IF NOT EXISTS activity (
       id integer PRIMARY KEY,
       name text NOT NULL,
       started_time datetime NOT NULL,
       end_time datetime NOT NULL,
       duration INTEGER NOT NULL,
       app_id INTEGER NOT NULL,
       FOREIGN KEY (app_id) REFERENCES apps (id)
);
