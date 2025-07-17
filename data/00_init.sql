
-- create a database user

DO $$
BEGIN
   IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'mydbuser') THEN
      CREATE ROLE mydbuser WITH LOGIN ENCRYPTED PASSWORD 'myDbPwd'; 
   END IF;
END $$;

CREATE EXTENSION Postgis;
CREATE EXTENSION pgcrypto;

grant all privileges on database mydatabase to mydbuser;
