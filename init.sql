CREATE TABLE IF NOT EXISTS public.news (
	news_id serial NOT NULL,
	title text NOT NULL,
	date timestamptz NOT NULL,
	CONSTRAINT pkey_suppliers PRIMARY KEY (news_id)
);

ALTER TABLE public.news OWNER TO role_1;
GRANT ALL ON TABLE public.news TO role_1;

CREATE OR REPLACE FUNCTION public.create_news(title_ text, date_ text)
 RETURNS integer
 LANGUAGE plpgsql
AS $function$
declare
id int4;
begin
  WITH upd as(INSERT into news(title, date) values ($1, $2::timestamp) RETURNING news_id as new_id)
select new_id into id from upd;
return id;
END;
$function$
;

CREATE OR REPLACE FUNCTION public.get_news(id integer)
 RETURNS SETOF json
 LANGUAGE plpgsql
AS $function$
begin
 return query
	WITH news as(select title, date from news where news_id= id)
select row_to_json(news.*)::json from news;
END;
$function$
;
