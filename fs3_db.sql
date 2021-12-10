--
-- PostgreSQL database dump
--

-- Dumped from database version 10.19 (Ubuntu 10.19-1.pgdg18.04+1)
-- Dumped by pg_dump version 14.1 (Ubuntu 14.1-1.pgdg18.04+1)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

--
-- Name: psql_volume_backup_car_csvs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.psql_volume_backup_car_csvs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    uuid text,
    source_file_name text,
    source_file_path text,
    source_file_md5 text,
    source_file_size bigint,
    car_file_name text,
    car_file_path text,
    car_file_md5 text,
    car_file_url text,
    car_file_size bigint,
    deal_cid text,
    data_cid text,
    piece_cid text,
    miner_fid text,
    start_epoch bigint,
    source_id bigint,
    cost text
);


ALTER TABLE public.psql_volume_backup_car_csvs OWNER TO root;

--
-- Name: psql_volume_backup_car_csvs_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.psql_volume_backup_car_csvs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.psql_volume_backup_car_csvs_id_seq OWNER TO root;

--
-- Name: psql_volume_backup_car_csvs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.psql_volume_backup_car_csvs_id_seq OWNED BY public.psql_volume_backup_car_csvs.id;


--
-- Name: psql_volume_backup_jobs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.psql_volume_backup_jobs (
    id bigint NOT NULL,
    name text,
    uuid text,
    source_file_name text,
    miner_id text,
    deal_cid text,
    payload_cid text,
    file_source_url text,
    md5 text,
    start_epoch bigint,
    piece_cid text,
    file_size bigint,
    cost text,
    duration text,
    status text,
    created_on text,
    updated_on text,
    volume_backup_plan_id bigint
);


ALTER TABLE public.psql_volume_backup_jobs OWNER TO root;

--
-- Name: psql_volume_backup_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.psql_volume_backup_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.psql_volume_backup_jobs_id_seq OWNER TO root;

--
-- Name: psql_volume_backup_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.psql_volume_backup_jobs_id_seq OWNED BY public.psql_volume_backup_jobs.id;


--
-- Name: psql_volume_backup_metadata_csvs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.psql_volume_backup_metadata_csvs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    uuid text,
    source_file_name text,
    source_file_path text,
    source_file_md5 text,
    source_file_size bigint,
    car_file_name text,
    car_file_path text,
    car_file_md5 text,
    car_file_url text,
    car_file_size bigint,
    deal_cid text,
    data_cid text,
    piece_cid text,
    miner_fid text,
    start_epoch bigint,
    source_id bigint,
    cost text
);


ALTER TABLE public.psql_volume_backup_metadata_csvs OWNER TO root;

--
-- Name: psql_volume_backup_metadata_csvs_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.psql_volume_backup_metadata_csvs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.psql_volume_backup_metadata_csvs_id_seq OWNER TO root;

--
-- Name: psql_volume_backup_metadata_csvs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.psql_volume_backup_metadata_csvs_id_seq OWNED BY public.psql_volume_backup_metadata_csvs.id;


--
-- Name: psql_volume_backup_plans; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.psql_volume_backup_plans (
    id bigint NOT NULL,
    name text,
    "interval" text,
    miner_region text,
    price text,
    duration text,
    verified_deal boolean,
    fast_retrieval boolean,
    status text,
    last_backup_on text,
    created_on text,
    updated_on text
);


ALTER TABLE public.psql_volume_backup_plans OWNER TO root;

--
-- Name: psql_volume_backup_plans_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.psql_volume_backup_plans_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.psql_volume_backup_plans_id_seq OWNER TO root;

--
-- Name: psql_volume_backup_plans_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.psql_volume_backup_plans_id_seq OWNED BY public.psql_volume_backup_plans.id;


--
-- Name: psql_volume_backup_task_csvs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.psql_volume_backup_task_csvs (
    id bigint NOT NULL,
    created_at timestamp with time zone,
    updated_at timestamp with time zone,
    deleted_at timestamp with time zone,
    uuid text,
    source_file_name text,
    miner_id text,
    deal_cid text,
    payload_cid text,
    file_source_url text,
    md5 text,
    start_epoch bigint,
    piece_cid text,
    file_size bigint,
    cost text
);


ALTER TABLE public.psql_volume_backup_task_csvs OWNER TO root;

--
-- Name: psql_volume_backup_task_csvs_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.psql_volume_backup_task_csvs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.psql_volume_backup_task_csvs_id_seq OWNER TO root;

--
-- Name: psql_volume_backup_task_csvs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.psql_volume_backup_task_csvs_id_seq OWNED BY public.psql_volume_backup_task_csvs.id;


--
-- Name: psql_volume_rebuild_jobs; Type: TABLE; Schema: public; Owner: root
--

CREATE TABLE public.psql_volume_rebuild_jobs (
    id bigint NOT NULL,
    miner_id text,
    deal_cid text,
    payload_cid text,
    status text,
    created_on text,
    updated_on text,
    backup_job_id bigint
);


ALTER TABLE public.psql_volume_rebuild_jobs OWNER TO root;

--
-- Name: psql_volume_rebuild_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: root
--

CREATE SEQUENCE public.psql_volume_rebuild_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER TABLE public.psql_volume_rebuild_jobs_id_seq OWNER TO root;

--
-- Name: psql_volume_rebuild_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: root
--

ALTER SEQUENCE public.psql_volume_rebuild_jobs_id_seq OWNED BY public.psql_volume_rebuild_jobs.id;


--
-- Name: psql_volume_backup_car_csvs id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_car_csvs ALTER COLUMN id SET DEFAULT nextval('public.psql_volume_backup_car_csvs_id_seq'::regclass);


--
-- Name: psql_volume_backup_jobs id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_jobs ALTER COLUMN id SET DEFAULT nextval('public.psql_volume_backup_jobs_id_seq'::regclass);


--
-- Name: psql_volume_backup_metadata_csvs id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_metadata_csvs ALTER COLUMN id SET DEFAULT nextval('public.psql_volume_backup_metadata_csvs_id_seq'::regclass);


--
-- Name: psql_volume_backup_plans id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_plans ALTER COLUMN id SET DEFAULT nextval('public.psql_volume_backup_plans_id_seq'::regclass);


--
-- Name: psql_volume_backup_task_csvs id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_task_csvs ALTER COLUMN id SET DEFAULT nextval('public.psql_volume_backup_task_csvs_id_seq'::regclass);


--
-- Name: psql_volume_rebuild_jobs id; Type: DEFAULT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_rebuild_jobs ALTER COLUMN id SET DEFAULT nextval('public.psql_volume_rebuild_jobs_id_seq'::regclass);


--
-- Name: psql_volume_backup_car_csvs psql_volume_backup_car_csvs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_car_csvs
    ADD CONSTRAINT psql_volume_backup_car_csvs_pkey PRIMARY KEY (id);


--
-- Name: psql_volume_backup_jobs psql_volume_backup_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_jobs
    ADD CONSTRAINT psql_volume_backup_jobs_pkey PRIMARY KEY (id);


--
-- Name: psql_volume_backup_metadata_csvs psql_volume_backup_metadata_csvs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_metadata_csvs
    ADD CONSTRAINT psql_volume_backup_metadata_csvs_pkey PRIMARY KEY (id);


--
-- Name: psql_volume_backup_plans psql_volume_backup_plans_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_plans
    ADD CONSTRAINT psql_volume_backup_plans_pkey PRIMARY KEY (id);


--
-- Name: psql_volume_backup_task_csvs psql_volume_backup_task_csvs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_task_csvs
    ADD CONSTRAINT psql_volume_backup_task_csvs_pkey PRIMARY KEY (id);


--
-- Name: psql_volume_rebuild_jobs psql_volume_rebuild_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_rebuild_jobs
    ADD CONSTRAINT psql_volume_rebuild_jobs_pkey PRIMARY KEY (id);


--
-- Name: idx_psql_volume_backup_car_csvs_deleted_at; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_psql_volume_backup_car_csvs_deleted_at ON public.psql_volume_backup_car_csvs USING btree (deleted_at);


--
-- Name: idx_psql_volume_backup_metadata_csvs_deleted_at; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_psql_volume_backup_metadata_csvs_deleted_at ON public.psql_volume_backup_metadata_csvs USING btree (deleted_at);


--
-- Name: idx_psql_volume_backup_task_csvs_deleted_at; Type: INDEX; Schema: public; Owner: root
--

CREATE INDEX idx_psql_volume_backup_task_csvs_deleted_at ON public.psql_volume_backup_task_csvs USING btree (deleted_at);


--
-- Name: psql_volume_backup_jobs fk_psql_volume_backup_jobs_volume_backup_plan; Type: FK CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_backup_jobs
    ADD CONSTRAINT fk_psql_volume_backup_jobs_volume_backup_plan FOREIGN KEY (volume_backup_plan_id) REFERENCES public.psql_volume_backup_plans(id);


--
-- Name: psql_volume_rebuild_jobs fk_psql_volume_rebuild_jobs_backup_job; Type: FK CONSTRAINT; Schema: public; Owner: root
--

ALTER TABLE ONLY public.psql_volume_rebuild_jobs
    ADD CONSTRAINT fk_psql_volume_rebuild_jobs_backup_job FOREIGN KEY (backup_job_id) REFERENCES public.psql_volume_backup_jobs(id);


--
-- PostgreSQL database dump complete
--

