INSERT INTO countries (code,name,default_language) VALUES ('LK','Sri Lanka','en');
INSERT INTO cities (country_id,name,slug) VALUES ((SELECT id FROM countries WHERE code='LK'),'Colombo','colombo'),((SELECT id FROM countries WHERE code='LK'),'Kandy','kandy'),((SELECT id FROM countries WHERE code='LK'),'Galle','galle');
INSERT INTO categories (name,slug) VALUES ('Dining','dining'),('Travel','travel'),('Electronics','electronics');
INSERT INTO deal_types (code,name) VALUES ('promotion','Promotions / Sales'),('event','Events'),('card_promo','Card Promotions');
INSERT INTO merchants (name,slug,contact,verified) VALUES ('Ceylon Eats','ceylon-eats','+94-11-000000',true),('Lanka Travels','lanka-travels','+94-81-111111',true);
INSERT INTO users (email,password_hash,name,role) VALUES ('admin@example.com','$2a$10$KQw9v4N3zgA6fP8fXe3vSO3A0qqCiwEv6eP9Q5Ytgk7BfQ.IgRjfm','Admin','admin');
INSERT INTO users (email,password_hash,name,role) VALUES ('owner@example.com','$2a$10$KQw9v4N3zgA6fP8fXe3vSO3A0qqCiwEv6eP9Q5Ytgk7BfQ.IgRjfm','Owner','submitter');

INSERT INTO deals (title,slug,description,country_id,city_id,category_id,merchant_id,deal_type_id,start_at,end_at,featured,image_url,status,created_by_user_id)
VALUES
('50% off dinner buffet','50-off-dinner-buffet','Special buffet offer in Colombo.',(SELECT id FROM countries WHERE code='LK'),(SELECT id FROM cities WHERE slug='colombo'),(SELECT id FROM categories WHERE slug='dining'),(SELECT id FROM merchants WHERE slug='ceylon-eats'),(SELECT id FROM deal_types WHERE code='promotion'),NOW()-INTERVAL '1 day',NOW()+INTERVAL '5 days',true,'/static/uploads/sample1.jpg','published',(SELECT id FROM users WHERE email='admin@example.com')),
('Visa card travel promo','visa-card-travel-promo','Use Visa and get travel benefits',(SELECT id FROM countries WHERE code='LK'),(SELECT id FROM cities WHERE slug='kandy'),(SELECT id FROM categories WHERE slug='travel'),(SELECT id FROM merchants WHERE slug='lanka-travels'),(SELECT id FROM deal_types WHERE code='card_promo'),NOW(),NOW()+INTERVAL '10 days',false,'/static/uploads/sample2.jpg','published',(SELECT id FROM users WHERE email='admin@example.com')),
('Tech expo 2026','tech-expo-2026','Electronics event in Galle',(SELECT id FROM countries WHERE code='LK'),(SELECT id FROM cities WHERE slug='galle'),(SELECT id FROM categories WHERE slug='electronics'),NULL,(SELECT id FROM deal_types WHERE code='event'),NOW(),NOW()+INTERVAL '15 days',true,'/static/uploads/sample3.jpg','published',(SELECT id FROM users WHERE email='admin@example.com'));

INSERT INTO deal_translations (deal_id,lang,title,description)
SELECT id,'en',title,description FROM deals;
INSERT INTO deal_translations (deal_id,lang,title,description)
SELECT id,'si','සිංහල '||title,'සිංහල විස්තර: '||description FROM deals;
INSERT INTO deal_translations (deal_id,lang,title,description)
SELECT id,'ta','தமிழ் '||title,'தமிழ் விளக்கம்: '||description FROM deals;

INSERT INTO admin_config (key,value) VALUES
('featured_categories','["dining","travel"]'),
('homepage_sections','{"showFeatured":true,"showEndingSoon":true}');
