CREATE TABLE uploads (
    id integer NOT NULL,
    slot_hash character(64),
    jid character varying,
    original_name character varying,
    disk_name character varying,
    upload_time timestamp with time zone,
    file_size integer,
    content_type character varying,
    slot_time timestamp with time zone
);


CREATE SEQUENCE uploads_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE uploads_id_seq OWNED BY uploads.id;

ALTER TABLE ONLY uploads
    ADD CONSTRAINT id PRIMARY KEY (id);
