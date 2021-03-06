CREATE TABLE uploads (
    id SERIAL,
    slot_hash character(64),
    jid character varying,
    original_name character varying,
    disk_name character varying,
    upload_time timestamp with time zone,
    file_size integer,
    content_type character varying,
    slot_time timestamp with time zone
);


ALTER TABLE ONLY uploads
    ADD CONSTRAINT id PRIMARY KEY (id);
