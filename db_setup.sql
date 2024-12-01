CREATE TABLE users (
	github_email VARCHAR PRIMARY KEY,
	name VARCHAR,
	registered_at TIMESTAMP,
	last_login TIMESTAMP
);

CREATE TABLE bookings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	registrar_email VARCHAR,
	outbound_trip INTEGER,
	vacation INTEGER,
	vacation_day_count INTEGER,
	inbound_trip INTEGER,
	total_days INTEGER,
	total_price INTEGER,
	FOREIGN KEY (registrar_email) REFERENCES users(github_email),
	FOREIGN KEY (outbound_trip) REFERENCES trip(id),
	FOREIGN KEY (inbound_trip) REFERENCES trip(id),
	FOREIGN KEY (vacation) REFERENCES vacations(id)
);

CREATE TABLE bookings_persons (
	booking_id INTEGER,
	person_id INTEGER,
	PRIMARY KEY (booking_id, person_id),
	FOREIGN KEY (booking_id) REFERENCES bookings(id),
	FOREIGN KEY (person_id) REFERENCES persons(id)
);

CREATE TABLE persons (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	nationality VARCHAR,
	passport_number VARCHAR,
	first_name VARCHAR,
	last_name VARCHAR
);

CREATE TABLE vacations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	city VARCHAR,
	hotel_budget VARCHAR,
	sightseeing_budget VARCHAR,
	total_price INTEGER
);

CREATE TABLE trip (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	departure_feeder INTEGER,
	trunk INTEGER,
	arrival_feeder INTEGER,
	total_price INTEGER,
	FOREIGN KEY (departure_feeder) REFERENCES leg(id),
	FOREIGN KEY (trunk) REFERENCES leg(id),
	FOREIGN KEY (arrival_feeder) REFERENCES leg(id)
);

CREATE TABLE leg (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type VARCHAR,
	budget VARCHAR,
	origin_city VARCHAR,
	destination_city VARCHAR,
	price INTEGER
);