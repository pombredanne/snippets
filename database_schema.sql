CREATE TABLE users (
	user_name STRING NOT NULL,
	real_name STRING NOT NULL,
	email_address STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (user_name ASC),
	FAMILY "primary" (user_name, real_name, email_address)
);

CREATE TABLE posts (
	user_name STRING NOT NULL,
	year INT NOT NULL,
	week INT NOT NULL,
	body_this_week STRING NOT NULL,
	body_next_week STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (user_name ASC, year ASC, week ASC),
	INDEX posts_user_name_idx (user_name ASC),
	CONSTRAINT fk_user_name_ref_users FOREIGN KEY (user_name) REFERENCES users (user_name),
	FAMILY "primary" (user_name, year, week, body_this_week, body_next_week),
	CONSTRAINT check_week_week CHECK ((week >= 1) AND (week <= 53)),
	CONSTRAINT check_body_this_week_body_next_week CHECK ((body_this_week != '') OR (body_next_week != ''))
);

CREATE TABLE subscriptions (
	subscriber STRING NOT NULL,
	subscribee STRING NOT NULL,
	CONSTRAINT "primary" PRIMARY KEY (subscriber ASC, subscribee ASC),
	CONSTRAINT fk_subscribee_ref_users FOREIGN KEY (subscribee) REFERENCES users (user_name),
	INDEX subscriptions_subscribee_idx (subscribee ASC),
	CONSTRAINT fk_subscriber_ref_users FOREIGN KEY (subscriber) REFERENCES users (user_name),
	FAMILY "primary" (subscriber, subscribee),
	CONSTRAINT check_subscriber_subscribee CHECK (subscriber != subscribee)
);
