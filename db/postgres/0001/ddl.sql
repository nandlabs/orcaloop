-- DROP SCHEMA public;

CREATE SCHEMA public AUTHORIZATION pg_database_owner;

COMMENT ON SCHEMA public IS 'standard public schema';

-- DROP TYPE public.status;

CREATE TYPE public.status AS ENUM (
	'Pending',
	'Running',
	'Completed',
	'Failed',
	'Skipped',
	'Unknown');
-- public.actions definition

-- Drop table

-- DROP TABLE public.actions;

CREATE TABLE public.actions (
	id varchar NOT NULL,
	"name" varchar NOT NULL,
	description text NULL,
	is_deleted bool DEFAULT false NOT NULL,
	param jsonb NULL,
	"returns" jsonb NULL,
	endpoint jsonb NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT actions_pkey PRIMARY KEY (id)
);


-- public.workflow_data definition

-- Drop table

-- DROP TABLE public.workflow_data;

CREATE TABLE public.workflow_data (
	instance_id varchar NOT NULL,
	workflow_id varchar NOT NULL,
	workflow_version int4 NOT NULL,
	pipeline_data jsonb NULL,
	is_deleted bool DEFAULT false NOT NULL,
	is_locked bool DEFAULT false NOT NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT workflow_data_pkey PRIMARY KEY (instance_id)
);


-- public.workflows definition

-- Drop table

-- DROP TABLE public.workflows;

CREATE TABLE public.workflows (
	workflow_id varchar NOT NULL,
	"version" int4 NOT NULL,
	"name" varchar NOT NULL,
	description text NULL,
	workflow_document jsonb NULL,
	is_deleted bool DEFAULT false NOT NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT workflows_pkey PRIMARY KEY (workflow_id, version)
);


-- public.pending_steps definition

-- Drop table

-- DROP TABLE public.pending_steps;

CREATE TABLE public.pending_steps (
	id varchar NOT NULL,
	instance_id varchar NOT NULL,
	step_id varchar NOT NULL,
	"data" jsonb NULL,
	is_deleted bool DEFAULT false NOT NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT pending_steps_pkey PRIMARY KEY (id),
	CONSTRAINT fk_pending_steps_instance FOREIGN KEY (instance_id) REFERENCES public.workflow_data(instance_id) ON DELETE CASCADE
);


-- public.step_change_event definition

-- Drop table

-- DROP TABLE public.step_change_event;

CREATE TABLE public.step_change_event (
	instance_id varchar NOT NULL,
	event_id varchar NOT NULL,
	step_id varchar NOT NULL,
	status public.status NULL,
	"data" jsonb NULL,
	is_deleted bool DEFAULT false NOT NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT step_change_event_pkey PRIMARY KEY (instance_id, event_id),
	CONSTRAINT fk_step_change_event_instance FOREIGN KEY (instance_id) REFERENCES public.workflow_data(instance_id) ON DELETE CASCADE
);


-- public.step_state definition

-- Drop table

-- DROP TABLE public.step_state;

CREATE TABLE public.step_state (
	instance_id varchar NOT NULL,
	step_id varchar NOT NULL,
	iteration int4 DEFAULT 0 NOT NULL,
	parent_step varchar NULL,
	child_count int4 NULL,
	status public.status NULL,
	step_state jsonb NULL,
	CONSTRAINT step_state_pkey PRIMARY KEY (instance_id, step_id, iteration),
	CONSTRAINT fk_step_state_instance FOREIGN KEY (instance_id) REFERENCES public.workflow_data(instance_id) ON DELETE CASCADE
);


-- public.workflow_state definition

-- Drop table

-- DROP TABLE public.workflow_state;

CREATE TABLE public.workflow_state (
	instance_id varchar NOT NULL,
	workflow_id varchar NOT NULL,
	workflow_version int4 NOT NULL,
	instance_version int4 NULL,
	status public.status NULL,
	"error" text NULL,
	created_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	updated_at timestamp DEFAULT CURRENT_TIMESTAMP NOT NULL,
	CONSTRAINT workflow_state_pkey PRIMARY KEY (instance_id, workflow_id, workflow_version),
	CONSTRAINT fk_workflow_state_instance FOREIGN KEY (instance_id) REFERENCES public.workflow_data(instance_id) ON DELETE CASCADE,
	CONSTRAINT fk_workflow_state_workflow FOREIGN KEY (workflow_id,workflow_version) REFERENCES public.workflows(workflow_id,"version") ON DELETE CASCADE
);