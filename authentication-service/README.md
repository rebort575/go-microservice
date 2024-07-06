1. 使用前需要在postgres数据库中创建表
    ```
    CREATE SEQUENCE public.user_id_seq
        START WITH 1
        INCREMENT BY 1
        NO MINVALUE
        NO MAXVALUE
        CACHE 1;


    ALTER TABLE public.user_id_seq OWNER TO postgres;

    SET default_tablespace = '';

    SET default_table_access_method = heap;

    --
    -- Name: users; Type: TABLE; Schema: public; Owner: postgres
    --

    CREATE TABLE public.users (
        id integer DEFAULT nextval('public.user_id_seq'::regclass) NOT NULL,
        email character varying(255),
        first_name character varying(255),
        last_name character varying(255),
        password character varying(60),
        user_active integer DEFAULT 0,
        created_at timestamp without time zone,
        updated_at timestamp without time zone
    );


    ALTER TABLE public.users OWNER TO postgres;

    --
    -- Name: user_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
    --

    SELECT pg_catalog.setval('public.user_id_seq', 1, true);


    --
    -- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
    --

    ALTER TABLE ONLY public.users
        ADD CONSTRAINT users_pkey PRIMARY KEY (id);

    ```
2. `main`函数连接postgres数据库成功后会插入一条测试数据，后续可使用测试数据测试