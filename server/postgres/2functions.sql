--name, ip
create or replace function add_external_user(
    provided_name VARCHAR, 
    provided_ip VARCHAR)
returns INTEGER
as
$$
DECLARE user_id_response INTEGER;
BEGIN
INSERT INTO users (created_at, internal) VALUES (CURRENT_TIMESTAMP, FALSE) RETURNING id INTO user_id_response;
INSERT INTO external_users (user_id, name, ip_address) VALUES (user_id_response, provided_name, provided_ip);
RETURN user_id_response;
END;
$$
language plpgsql;

CREATE OR REPLACE FUNCTION update_external_user_info(
    current_user_id INT,
    new_name VARCHAR,
    new_ip_address VARCHAR,
    new_email VARCHAR
) RETURNS TABLE (
	given_user_id INT,
	updated_name VARCHAR,
	updated_ip_address VARCHAR,
	updated_email VARCHAR
) AS $$
	
DECLARE
	original_name VARCHAR;
	original_ip_address VARCHAR;
	original_email VARCHAR;

BEGIN

  -- Retrieve current values for return
    SELECT COALESCE(name, ''),
           COALESCE(ip_address, ''),
           COALESCE(email, '')
	INTO original_name, original_ip_address, original_email
    FROM external_users WHERE user_id = $1;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'record not found';
    END IF;

  UPDATE external_users
  SET
      name = COALESCE(NULLIF(new_name, ''), original_name),
      ip_address = COALESCE(NULLIF(new_ip_address, ''), original_ip_address),
      email = COALESCE(NULLIF(new_email, ''), original_email)
  WHERE user_id = $1;

  RETURN QUERY (
	  SELECT
	  current_user_id as given_user_id, 
      COALESCE(NULLIF(new_name, '')::VARCHAR, original_name) as updated_name,
      COALESCE(NULLIF(new_ip_address, '')::VARCHAR, original_ip_address) as updated_ip_address,
      COALESCE(NULLIF(new_email, '')::VARCHAR, original_email) as updated_email
  );
END;
$$ LANGUAGE plpgsql;


create or replace function add_internal_user(
	provided_role_id INT,
	provided_firstname VARCHAR,
	provided_surname VARCHAR,
	provided_email VARCHAR,
	hashed_password VARCHAR)
returns INTEGER
as
$$
DECLARE user_id_response INTEGER;
BEGIN
INSERT INTO users (created_at, internal) VALUES (CURRENT_TIMESTAMP, TRUE) RETURNING id INTO user_id_response;
INSERT INTO internal_users (user_id, role_id, firstname, surname, email, password) VALUES (user_id_response, provided_role_id, 
																						   provided_firstname, provided_surname,
																						  provided_email, hashed_password);
RETURN user_id_response;
END;
$$
language plpgsql;

CREATE OR REPLACE FUNCTION update_internal_user_info(
    current_user_id INT,
    new_role_id INT,
    new_firstname VARCHAR,
    new_surname VARCHAR,
    new_email VARCHAR,
    new_password VARCHAR
) RETURNS TABLE (
	given_user_id INT,
    updated_role_id INT,
	updated_firstname VARCHAR,
    updated_surname VARCHAR,
	updated_email VARCHAR,
	updated_password VARCHAR
) AS $$
	
DECLARE
	original_role_id INT;
    original_firstname VARCHAR;
    original_surname VARCHAR;
    original_email VARCHAR;
    original_password VARCHAR;

BEGIN

  -- Retrieve current values for return
    SELECT role_id, firstname, surname, email, password
	INTO original_role_id, original_firstname, original_surname, original_email, original_password
    FROM internal_users WHERE user_id = $1;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'record not found';
    END IF;

  UPDATE internal_users
  SET
      role_id = COALESCE(NULLIF(new_role_id, 0), original_role_id),
      firstname = COALESCE(NULLIF(new_firstname, ''), original_firstname),
      surname = COALESCE(NULLIF(new_surname, ''), original_surname),
      email = COALESCE(NULLIF(new_email, ''), original_email),
      password = COALESCE(NULLIF(new_password, ''), original_password)
  WHERE user_id = $1;

  RETURN QUERY (
	  SELECT
	  current_user_id as given_user_id, 
      COALESCE(NULLIF(new_role_id, 0)::INT, original_role_id) as updated_role_id,
      COALESCE(NULLIF(new_firstname, '')::VARCHAR, original_firstname) as updated_firstname,
      COALESCE(NULLIF(new_surname, '')::VARCHAR, original_surname) as updated_surname,
      COALESCE(NULLIF(new_email, '')::VARCHAR, original_email) as updated_email,
      COALESCE(NULLIF(new_password, '')::VARCHAR, original_password) as updated_password
  );
END;
$$ LANGUAGE plpgsql;