
create table pairs (
    id SERIAL PRIMARY KEY ,
    pair VARCHAR(255) UNIQUE NOT NULL);

create table prices (
    id SERIAL PRIMARY KEY,
    pair VARCHAR(255) NOT NULL,
    price float not null,
    price_time  BIGINT);