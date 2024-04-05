CREATE TABLE IF NOT EXISTS "user_roles" (
        "id" serial NOT NULL UNIQUE,
        "description" varchar NOT NULL,
        CONSTRAINT "user_roles_pk" PRIMARY KEY ("id")
) WITH (
  OIDS=FALSE
);

CREATE TABLE IF NOT EXISTS "users" (
        "id" serial NOT NULL UNIQUE,
        "created_at" timestamp with time zone NOT NULL,
        "internal" BOOLEAN NOT NULL,
        CONSTRAINT "users_pk" PRIMARY KEY ("id")
) WITH (
  OIDS=FALSE
);

CREATE TABLE IF NOT EXISTS "external_users" (
        "id" serial NOT NULL,
        "user_id" integer NOT NULL,
        "name" varchar,
        "email" varchar,
        "ip_address" varchar,
        CONSTRAINT "external_users_pk" PRIMARY KEY ("id")
) WITH (
  OIDS=FALSE
);

CREATE TABLE IF NOT EXISTS "internal_users" (
        "id" serial NOT NULL,
        "user_id" integer NOT NULL,
        "role_id" integer NOT NULL,
        "firstname" varchar NOT NULL,
        "surname" varchar NOT NULL,
        "email" varchar NOT NULL,
        "password" varchar NOT NULL,
        CONSTRAINT "internal_users_pk" PRIMARY KEY ("id")
) WITH (
  OIDS=FALSE
);

CREATE TABLE IF NOT EXISTS "chat" (
        "uuid" uuid NOT NULL UNIQUE,
        "start_time" timestamp with time zone NOT NULL,
        "end_time" timestamp with time zone,
        CONSTRAINT "chat_pk" PRIMARY KEY ("uuid")
) WITH (
  OIDS=FALSE
);

CREATE TABLE IF NOT EXISTS "chat_messages" (
        "id" serial NOT NULL UNIQUE,
        "chat_uuid" uuid NOT NULL,
        "user_id_from" integer NOT NULL,
        "message" varchar NOT NULL,
        "timestamp" timestamp with time zone,
        CONSTRAINT "chat_messages_pk" PRIMARY KEY ("id")
) WITH (
  OIDS=FALSE
);

CREATE TABLE IF NOT EXISTS "chat_participant"(
    "id" serial NOT NULL UNIQUE,
    "chat_uuid" uuid NOT NULL,
    "user_id" integer NOT NULL,
    "time_joined" timestamp with time zone,
    "time_left" timestamp with time zone,
    CONSTRAINT "chat_participant_pk" PRIMARY KEY ("id")
  ) WITH (
    OIDS=FALSE
);

ALTER TABLE "internal_users" ADD CONSTRAINT "internal_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users"("id");
ALTER TABLE "internal_users" ADD CONSTRAINT "internal_users_role_id" FOREIGN KEY ("role_id") REFERENCES "user_roles"("id");
ALTER TABLE "external_users" ADD CONSTRAINT "external_users_user_id" FOREIGN KEY ("user_id") REFERENCES "users"("id");
ALTER TABLE "chat_messages" ADD CONSTRAINT "chat_messages_uuid" FOREIGN KEY ("chat_uuid") REFERENCES "chat"("uuid");
ALTER TABLE "chat_messages" ADD CONSTRAINT "chat_messages_user_id" FOREIGN KEY ("user_id_from") REFERENCES "users"("id");
ALTER TABLE "chat_participant" ADD CONSTRAINT "chat_participant_chat_uuid" FOREIGN KEY ("chat_uuid") REFERENCES "chat"("uuid");
ALTER TABLE "chat_participant" ADD CONSTRAINT "chat_participant_user_id" FOREIGN KEY ("user_id") REFERENCES "users"("id");