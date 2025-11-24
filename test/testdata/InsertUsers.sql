INSERT INTO teams (team_name) 
VALUES ('test-team');

INSERT INTO users (user_id, username, team_name, is_active) 
VALUES ('u1', 'Alice', 'test-team', true),
('u2', 'Bob', 'test-team', false),
('u3', 'Charlie', 'test-team', true),
('u4', 'David', 'test-team', false),
('u5', 'Ivan', 'test-team', true);
