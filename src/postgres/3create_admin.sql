insert into user_roles (id, description) values (1, 'admin')
on conflict do nothing;

insert into users (id, created_at, internal) values (1, CURRENT_TIMESTAMP, true)
on conflict do nothing;

insert into internal_users (user_id, role_id, firstname, surname, email, password)
values (1, 1, 'admin', 'user', 'admin@example.com', 'password')
on conflict do nothing;

SELECT SETVAL((SELECT PG_GET_SERIAL_SEQUENCE('"users"', 'id')), (SELECT (MAX("id") + 1) FROM "users"), FALSE);