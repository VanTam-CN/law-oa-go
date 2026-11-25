-- Run through compare-main-browser-db.py seed on the isolated v14+86 database.
\getenv password NP_BROWSER_PASSWORD
SELECT gen_random_uuid() AS target, gen_random_uuid() AS control \gset
SELECT 'np.' || left(:'target', 18) AS username, '虚构接案·' || left(:'target', 8) AS target_title \gset
INSERT INTO users(username, email, name, password, role, status)
VALUES (:'username', :'username' || '@np.invalid', '虚构试验律师',
        crypt(:'password', gen_salt('bf', 10)), 'lawyer', 'active') RETURNING id AS actor_id \gset
INSERT INTO clients(name, type, created_by) VALUES ('虚构试验客户', '企业', :actor_id)
RETURNING id AS client_id \gset
INSERT INTO case_intakes(id,intake_code,title,client_id,created_by,case_type,description,status,metadata)
SELECT id, 'NP-' || id, title, :client_id, :actor_id, 'civil', '虚构初始摘要', 'draft',
       jsonb_build_object('draft_revision', gen_random_uuid(), 'lawyer_id', :'actor_id',
                          'np_unrelated', jsonb_build_object('big', 9007199254740993::bigint))
FROM (VALUES (:'target'::uuid, :'target_title'), (:'control'::uuid, '虚构对照草稿')) d(id,title);
INSERT INTO case_intake_parties(intake_id,entity_name,entity_type,party_role)
SELECT id,'虚构对方','LEGAL_PERSON','opposing_party' FROM unnest(ARRAY[:'target',:'control']::uuid[]) id;
INSERT INTO case_materials(intake_id,name,material_type)
VALUES (:'target','虚构清单甲','证据'),(:'target','虚构清单乙','证据'),(:'control','虚构对照清单','证据');
SELECT jsonb_build_object('target', :'target', 'control', :'control', 'target_title', :'target_title',
    'username', :'username', 'actor_id', :'actor_id', 'client_id', :'client_id',
    'parties', (SELECT jsonb_agg(id ORDER BY id) FROM case_intake_parties WHERE intake_id IN (:'target',:'control')),
    'materials', (SELECT jsonb_agg(id ORDER BY id) FROM case_materials WHERE intake_id IN (:'target',:'control')));
