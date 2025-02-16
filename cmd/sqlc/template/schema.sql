-- public.m_frame definition

-- Drop table

-- DROP TABLE public.m_frame;

CREATE TABLE public.m_frame (
	id bigserial NOT NULL,
	pray_date date NOT NULL,
	target_hour int4 NOT NULL,
	start_time timestamp NOT NULL,
	closing_time timestamp NOT NULL,
	group_count int4 NOT NULL,
	group_people_count int4 NOT NULL,
	total_people_count int4 NOT NULL,
	unit_minute int4 DEFAULT 0 NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_frame_pkey PRIMARY KEY (id)
);


-- public.m_frame_releases definition

-- Drop table

-- DROP TABLE public.m_frame_releases;

CREATE TABLE public.m_frame_releases (
	pray_date timestamp NOT NULL,
	opening_timestamp timestamp NOT NULL,
	closing_timestamp timestamp NOT NULL,
	CONSTRAINT m_frame_releases_pk PRIMARY KEY (pray_date)
);


-- public.m_pray_category definition

-- Drop table

-- DROP TABLE public.m_pray_category;

CREATE TABLE public.m_pray_category (
	id int8 NOT NULL,
	title varchar NOT NULL,
	description varchar NOT NULL,
	is_corprate bool DEFAULT false NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_pray_category_pkey PRIMARY KEY (id)
);


-- public.m_role definition

-- Drop table

-- DROP TABLE public.m_role;

CREATE TABLE public.m_role (
	id bigserial NOT NULL,
	role_name varchar NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_role_pkey PRIMARY KEY (id)
);


-- public.m_writing_content definition

-- Drop table

-- DROP TABLE public.m_writing_content;

CREATE TABLE public.m_writing_content (
	id bigserial NOT NULL,
	title varchar NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_writing_content_pkey PRIMARY KEY (id)
);


-- public.sys_connecten_settings definition

-- Drop table

-- DROP TABLE public.sys_connecten_settings;

CREATE TABLE public.sys_connecten_settings (
	"key" varchar NOT NULL,
	data_type int4 NOT NULL,
	data_string varchar NULL,
	data_long int8 NULL,
	data_decimal numeric(6, 3) NULL,
	data_boolean bool NULL,
	description varchar NULL,
	CONSTRAINT sys_connecten_settings_pk PRIMARY KEY (key)
);


-- public.t_bulletin_board definition

-- Drop table

-- DROP TABLE public.t_bulletin_board;

CREATE TABLE public.t_bulletin_board (
	id bigserial NOT NULL,
	"content" varchar NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	CONSTRAINT t_bulletin_board_pkey PRIMARY KEY (id)
);


-- public.m_frame_details definition

-- Drop table

-- DROP TABLE public.m_frame_details;

CREATE TABLE public.m_frame_details (
	id bigserial NOT NULL,
	frame_id int8 NOT NULL,
	pray_date date NOT NULL,
	target_hour int4 NOT NULL,
	start_time timestamp NOT NULL,
	closing_time timestamp NOT NULL,
	group_count int4 NOT NULL,
	group_people_count int4 NOT NULL,
	total_people_count int4 NOT NULL,
	is_closed bool NOT NULL,
	close_reason varchar NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	is_not_decide_datetime bool DEFAULT false NOT NULL,
	CONSTRAINT m_frame_details_pkey PRIMARY KEY (id),
	CONSTRAINT m_frame_details_frame_id_fkey FOREIGN KEY (frame_id) REFERENCES public.m_frame(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_mail_template_placeholder definition

-- Drop table

-- DROP TABLE public.m_mail_template_placeholder;

CREATE TABLE public.m_mail_template_placeholder (
	id bigserial NOT NULL,
	pray_category_id int8 NOT NULL,
	placeholder_display varchar NOT NULL,
	field_name varchar NOT NULL,
	formatter varchar NOT NULL,
	view_order int8 NOT NULL,
	CONSTRAINT m_mail_template_placeholder_pkey PRIMARY KEY (id),
	CONSTRAINT m_mail_template_placeholder_pray_category_id_fkey FOREIGN KEY (pray_category_id) REFERENCES public.m_pray_category(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_mail_templates definition

-- Drop table

-- DROP TABLE public.m_mail_templates;

CREATE TABLE public.m_mail_templates (
	id bigserial NOT NULL,
	pray_category_id int8 NOT NULL,
	process_classification int4 NOT NULL,
	mail_subject varchar NOT NULL,
	mail_body varchar NOT NULL,
	is_active bool DEFAULT false NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_mail_templates_pkey PRIMARY KEY (id),
	CONSTRAINT m_mail_templates_pray_category_id_fkey FOREIGN KEY (pray_category_id) REFERENCES public.m_pray_category(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_user definition

-- Drop table

-- DROP TABLE public.m_user;

CREATE TABLE public.m_user (
	id bigserial NOT NULL,
	role_id int8 NOT NULL,
	auth_key varchar NOT NULL,
	login_id varchar NOT NULL,
	password_hash varchar NOT NULL,
	username varchar DEFAULT ''::character varying NOT NULL,
	mailaddress varchar DEFAULT ''::character varying NOT NULL,
	authority_level int4 DEFAULT 0 NOT NULL,
	last_logedin_at timestamp NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_user_auth_key_key UNIQUE (auth_key),
	CONSTRAINT m_user_login_id_key UNIQUE (login_id),
	CONSTRAINT m_user_pkey PRIMARY KEY (id),
	CONSTRAINT m_user_role_id_fkey FOREIGN KEY (role_id) REFERENCES public.m_role(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_wishes definition

-- Drop table

-- DROP TABLE public.m_wishes;

CREATE TABLE public.m_wishes (
	id bigserial NOT NULL,
	pray_category_id int8 NOT NULL,
	wish_name varchar NOT NULL,
	description varchar NOT NULL,
	is_date_designation bool NOT NULL,
	is_manual_input bool DEFAULT false NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_wishes_pkey PRIMARY KEY (id),
	CONSTRAINT m_wishes_pray_category_id_fkey FOREIGN KEY (pray_category_id) REFERENCES public.m_pray_category(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_wishes_fee definition

-- Drop table

-- DROP TABLE public.m_wishes_fee;

CREATE TABLE public.m_wishes_fee (
	id bigserial NOT NULL,
	wishes_id int8 NOT NULL,
	fee int8 DEFAULT 0 NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_wishes_fee_pkey PRIMARY KEY (id),
	CONSTRAINT m_wishes_fee_wishes_id_fkey FOREIGN KEY (wishes_id) REFERENCES public.m_wishes(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_wishes_option definition

-- Drop table

-- DROP TABLE public.m_wishes_option;

CREATE TABLE public.m_wishes_option (
	id bigserial NOT NULL,
	wishes_id int8 NOT NULL,
	content_id varchar NOT NULL,
	is_visible bool NOT NULL,
	description varchar NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_wishes_option_pkey PRIMARY KEY (id),
	CONSTRAINT m_wishes_option_wishes_id_fkey FOREIGN KEY (wishes_id) REFERENCES public.m_wishes(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_wishes_text definition

-- Drop table

-- DROP TABLE public.m_wishes_text;

CREATE TABLE public.m_wishes_text (
	id bigserial NOT NULL,
	wishes_id int8 NOT NULL,
	content_id varchar NOT NULL,
	field_kind int4 NOT NULL,
	display_text varchar NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_wishes_text_pk PRIMARY KEY (id),
	CONSTRAINT m_wishes_text_wishes_id_fkey FOREIGN KEY (wishes_id) REFERENCES public.m_wishes(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_wishes_writing_content definition

-- Drop table

-- DROP TABLE public.m_wishes_writing_content;

CREATE TABLE public.m_wishes_writing_content (
	id bigserial NOT NULL,
	wishes_id int8 NOT NULL,
	writing_content_id int8 NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_wishes_writing_content_pkey PRIMARY KEY (id),
	CONSTRAINT m_wishes_writing_content_wishes_id_fkey FOREIGN KEY (wishes_id) REFERENCES public.m_wishes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
	CONSTRAINT m_wishes_writing_content_writing_content_id_fkey FOREIGN KEY (writing_content_id) REFERENCES public.m_writing_content(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_reserve definition

-- Drop table

-- DROP TABLE public.t_reserve;

CREATE TABLE public.t_reserve (
	id bigserial NOT NULL,
	frame_details_id int8 NOT NULL,
	pray_category_id int8 NOT NULL,
	pray_date date NOT NULL,
	applicant_name varchar NOT NULL,
	applicant_email_address varchar NOT NULL,
	applicant_tel_no varchar NOT NULL,
	participants_number int4 NOT NULL,
	memo text DEFAULT ''::text NOT NULL,
	reservation_key varchar NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	booking_key varchar NULL,
	CONSTRAINT t_reserve_pkey PRIMARY KEY (id),
	CONSTRAINT t_reserve_frame_details_id_fkey FOREIGN KEY (frame_details_id) REFERENCES public.m_frame_details(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
	CONSTRAINT t_reserve_pray_category_id_fkey FOREIGN KEY (pray_category_id) REFERENCES public.m_pray_category(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_reserve_history definition

-- Drop table

-- DROP TABLE public.t_reserve_history;

CREATE TABLE public.t_reserve_history (
	id bigserial NOT NULL,
	reserve_id int8 NOT NULL,
	history_added_at timestamp NOT NULL,
	vm_json text NOT NULL,
	CONSTRAINT t_reserve_history_pkey PRIMARY KEY (id),
	CONSTRAINT t_reserve_history_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_reserve(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_reserve_overbooking definition

-- Drop table

-- DROP TABLE public.t_reserve_overbooking;

CREATE TABLE public.t_reserve_overbooking (
	reserve_id int8 NOT NULL,
	booking_key varchar NOT NULL,
	frame_detail_id int8 NOT NULL,
	expiration_datetime timestamp NOT NULL,
	CONSTRAINT t_reserve_overbooking_pk PRIMARY KEY (reserve_id),
	CONSTRAINT t_reserve_overbooking_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_reserve(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);
CREATE INDEX t_reserve_overbooking_booking_key_idx ON public.t_reserve_overbooking USING btree (booking_key);


-- public.t_souvenir definition

-- Drop table

-- DROP TABLE public.t_souvenir;

CREATE TABLE public.t_souvenir (
	reserve_id int8 NOT NULL,
	souvenir01 int4 DEFAULT 0 NOT NULL,
	souvenir02 int4 DEFAULT 0 NOT NULL,
	souvenir03 int4 DEFAULT 0 NOT NULL,
	souvenir04 int4 DEFAULT 0 NOT NULL,
	souvenir05 int4 DEFAULT 0 NOT NULL,
	CONSTRAINT t_souvenir_pk PRIMARY KEY (reserve_id),
	CONSTRAINT t_souvenir_fk FOREIGN KEY (reserve_id) REFERENCES public.t_reserve(id)
);


-- public.m_calendar definition

-- Drop table

-- DROP TABLE public.m_calendar;

CREATE TABLE public.m_calendar (
	id bigserial NOT NULL,
	display_name varchar NOT NULL,
	bind_user bool NOT NULL,
	owner_user_id int8 NULL,
	is_deletable bool DEFAULT true NOT NULL,
	is_visible bool DEFAULT true NOT NULL,
	is_reserve_calendar bool DEFAULT false NOT NULL,
	background_color varchar DEFAULT ''::character varying NOT NULL,
	foreground_color varchar DEFAULT ''::character varying NOT NULL,
	view_order int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT m_calendar_pkey PRIMARY KEY (id),
	CONSTRAINT m_calendar_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.m_user(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.m_frame_details_wishes definition

-- Drop table

-- DROP TABLE public.m_frame_details_wishes;

CREATE TABLE public.m_frame_details_wishes (
	id bigserial NOT NULL,
	frame_details int8 NOT NULL,
	wishes_id int8 NOT NULL,
	CONSTRAINT m_frame_details_wishes_pkey PRIMARY KEY (id),
	CONSTRAINT m_frame_details_wishes_frame_details_fkey FOREIGN KEY (frame_details) REFERENCES public.m_frame_details(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
	CONSTRAINT m_frame_details_wishes_wishes_id_fkey FOREIGN KEY (wishes_id) REFERENCES public.m_wishes(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_calendar_plan definition

-- Drop table

-- DROP TABLE public.t_calendar_plan;

CREATE TABLE public.t_calendar_plan (
	id bigserial NOT NULL,
	calendar_id int8 NOT NULL,
	title varchar NOT NULL,
	start_timestamp timestamp NOT NULL,
	end_timestamp timestamp NOT NULL,
	is_all_day bool DEFAULT false NOT NULL,
	plan_address varchar NOT NULL,
	descriptions varchar NOT NULL,
	owner_id int8 NOT NULL,
	is_private bool DEFAULT false NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_calendar_plan_pkey PRIMARY KEY (id),
	CONSTRAINT t_calendar_plan_calendar_id_fkey FOREIGN KEY (calendar_id) REFERENCES public.m_calendar(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_fuda_size definition

-- Drop table

-- DROP TABLE public.t_fuda_size;

CREATE TABLE public.t_fuda_size (
	reserve_id int8 NOT NULL,
	size_xl int4 DEFAULT 0 NOT NULL,
	size_lg int4 DEFAULT 0 NOT NULL,
	size_md int4 DEFAULT 0 NOT NULL,
	size_sm int4 DEFAULT 0 NOT NULL,
	size_xs int4 DEFAULT 0 NOT NULL,
	size_car int4 DEFAULT 0 NOT NULL,
	CONSTRAINT t_fuda_size_pk PRIMARY KEY (reserve_id),
	CONSTRAINT t_fuda_size_fk FOREIGN KEY (reserve_id) REFERENCES public.t_reserve(id)
);


-- public.t_mail_history definition

-- Drop table

-- DROP TABLE public.t_mail_history;

CREATE TABLE public.t_mail_history (
	id bigserial NOT NULL,
	reserve_id int8 NOT NULL,
	subject varchar NOT NULL,
	body text NOT NULL,
	is_sent bool DEFAULT false NOT NULL,
	is_draft bool DEFAULT false NOT NULL,
	sent_by int8 NOT NULL,
	sent_at timestamp NOT NULL,
	last_saved_at timestamp NOT NULL,
	CONSTRAINT t_mail_history_pkey PRIMARY KEY (id),
	CONSTRAINT t_mail_history_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_reserve(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_basic definition

-- Drop table

-- DROP TABLE public.t_pray_basic;

CREATE TABLE public.t_pray_basic (
	reserve_id int8 NOT NULL,
	main_wish_id int8 NOT NULL,
	main_wish_name varchar NOT NULL,
	sub_wish_id int8 NULL,
	sub_wish_name varchar NOT NULL,
	post_code varchar(7) NOT NULL,
	address1 varchar NOT NULL,
	address2 varchar NOT NULL,
	address3 varchar NOT NULL,
	address1_kana varchar NOT NULL,
	address2_kana varchar NOT NULL,
	address3_kana varchar NOT NULL,
	receipt_addressee varchar NOT NULL,
	pray_fee int8 DEFAULT 0 NOT NULL,
	remarks varchar NOT NULL,
	slip_remarks text DEFAULT ''::text NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	disable_receipt bool DEFAULT false NOT NULL,
	receipt_summary_amount int8 DEFAULT 0 NOT NULL,
	CONSTRAINT t_pray_basic_pkey PRIMARY KEY (reserve_id),
	CONSTRAINT t_pray_basic_main_wish_id_fkey FOREIGN KEY (main_wish_id) REFERENCES public.m_wishes(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
	CONSTRAINT t_pray_basic_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_reserve(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
	CONSTRAINT t_pray_basic_sub_wish_id_fkey FOREIGN KEY (sub_wish_id) REFERENCES public.m_wishes(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_category_accomplishment definition

-- Drop table

-- DROP TABLE public.t_pray_category_accomplishment;

CREATE TABLE public.t_pray_category_accomplishment (
	reserve_id int8 NOT NULL,
	first_aspirations varchar NOT NULL,
	first_aspirations_kana varchar NOT NULL,
	first_aspirations_date date NULL,
	second_aspirations varchar NOT NULL,
	second_aspirations_kana varchar NOT NULL,
	second_aspirations_date date NULL,
	third_aspirations varchar NOT NULL,
	third_aspirations_kana varchar NOT NULL,
	third_aspirations_date date NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_pray_category_accomplishment_pkey PRIMARY KEY (reserve_id),
	CONSTRAINT t_pray_category_accomplishment_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_category_child definition

-- Drop table

-- DROP TABLE public.t_pray_category_child;

CREATE TABLE public.t_pray_category_child (
	reserve_id int8 NOT NULL,
	father_name varchar NOT NULL,
	father_name_kana varchar NOT NULL,
	mother_name varchar NOT NULL,
	mother_name_kana varchar NOT NULL,
	children_count int4 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_pray_category_child_pkey PRIMARY KEY (reserve_id),
	CONSTRAINT t_pray_category_child_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_category_corporate definition

-- Drop table

-- DROP TABLE public.t_pray_category_corporate;

CREATE TABLE public.t_pray_category_corporate (
	reserve_id int8 NOT NULL,
	safety_content varchar NOT NULL,
	safety_period varchar NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_pray_category_corporate_pkey PRIMARY KEY (reserve_id),
	CONSTRAINT t_pray_category_corporate_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_category_couple definition

-- Drop table

-- DROP TABLE public.t_pray_category_couple;

CREATE TABLE public.t_pray_category_couple (
	reserve_id int8 NOT NULL,
	expected_date_of_birth varchar NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	belly_band_div int8 DEFAULT 0 NOT NULL,
	CONSTRAINT t_pray_category_couple_pkey PRIMARY KEY (reserve_id),
	CONSTRAINT t_pray_category_couple_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_category_personal definition

-- Drop table

-- DROP TABLE public.t_pray_category_personal;

CREATE TABLE public.t_pray_category_personal (
	reserve_id int8 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_pray_category_personal_pkey PRIMARY KEY (reserve_id),
	CONSTRAINT t_pray_category_personal_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_category_safety definition

-- Drop table

-- DROP TABLE public.t_pray_category_safety;

CREATE TABLE public.t_pray_category_safety (
	reserve_id int8 NOT NULL,
	safety_content varchar NOT NULL,
	safety_period varchar NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_pray_category_safety_pkey PRIMARY KEY (reserve_id),
	CONSTRAINT t_pray_category_safety_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_pray_writing_content definition

-- Drop table

-- DROP TABLE public.t_pray_writing_content;

CREATE TABLE public.t_pray_writing_content (
	id bigserial NOT NULL,
	reserve_id int8 NOT NULL,
	writing_content_id int8 NOT NULL,
	CONSTRAINT t_pray_writing_content_pkey PRIMARY KEY (id),
	CONSTRAINT t_pray_writing_content_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_reserve(id) ON DELETE RESTRICT ON UPDATE RESTRICT,
	CONSTRAINT t_pray_writing_content_writing_content_id_fkey FOREIGN KEY (writing_content_id) REFERENCES public.m_writing_content(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_prayer_corprate definition

-- Drop table

-- DROP TABLE public.t_prayer_corprate;

CREATE TABLE public.t_prayer_corprate (
	id bigserial NOT NULL,
	reserve_id int8 NOT NULL,
	company_name varchar NOT NULL,
	company_name_kana varchar NOT NULL,
	branch_and_department_name varchar NOT NULL,
	branch_and_department_name_kana varchar NOT NULL,
	representative_job_title varchar NOT NULL,
	representative_job_title_kana varchar NOT NULL,
	representative_name varchar NOT NULL,
	representative_name_kana varchar NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_prayer_corprate_pkey PRIMARY KEY (id),
	CONSTRAINT t_prayer_corprate_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);


-- public.t_prayer_personal definition

-- Drop table

-- DROP TABLE public.t_prayer_personal;

CREATE TABLE public.t_prayer_personal (
	id bigserial NOT NULL,
	reserve_id int8 NOT NULL,
	prayer_name varchar NOT NULL,
	prayer_name_kana varchar NOT NULL,
	gender int4 NOT NULL,
	birthday date NULL,
	"attribute" int8 DEFAULT 0 NOT NULL,
	pray_fee int8 DEFAULT 0 NOT NULL,
	is_deleted bool NOT NULL,
	created_by int8 NOT NULL,
	created_at timestamp NOT NULL,
	updated_by int8 NOT NULL,
	updated_at timestamp NOT NULL,
	CONSTRAINT t_prayer_personal_pkey PRIMARY KEY (id),
	CONSTRAINT t_prayer_personal_reserve_id_fkey FOREIGN KEY (reserve_id) REFERENCES public.t_pray_basic(reserve_id) ON DELETE RESTRICT ON UPDATE RESTRICT
);