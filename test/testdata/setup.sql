CREATE TABLE teams (
	team_name VARCHAR(255) PRIMARY KEY
);

CREATE TABLE users (
	user_id VARCHAR(255) PRIMARY KEY,
	username VARCHAR(255) NOT NULL,
	is_active BOOLEAN NOT NULL,
	team_name VARCHAR(255) NOT NULL,
	FOREIGN KEY(team_name) REFERENCES teams(team_name)
);

CREATE TABLE pull_requests_status (
	pr_status_id SERIAL PRIMARY KEY,
	status VARCHAR(255) NOT NULL 
);

CREATE TABLE pull_requests (
	pull_request_id VARCHAR(255) PRIMARY KEY,
	pull_request_name VARCHAR(255) NOT NULL,
	author_id VARCHAR(255) NOT NULL,
	status_id INT NOT NULL,
	created_at TIMESTAMP NOT NULL,
	merged_at TIMESTAMP NULL,
	FOREIGN KEY(author_id) REFERENCES users(user_id),
	FOREIGN KEY(status_id) REFERENCES pull_requests_status(pr_status_id)
);



CREATE TABLE reviewers (
	reviewer_id SERIAL PRIMARY KEY,
	user_id VARCHAR(255) NOT NULL,
	pull_request_id VARCHAR(255) NOT NULL,
	FOREIGN KEY(user_id) REFERENCES users(user_id),
	FOREIGN KEY(pull_request_id) REFERENCES pull_requests(pull_request_id)
);

INSERT INTO pull_requests_status(pr_status_id, status) VALUES(1, 'OPEN'), (2, 'MERGED');