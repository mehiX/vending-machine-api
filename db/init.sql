create table users (
    id varchar(64) not null,
    username varchar(64) not null,
    password varchar(256) not null,
    deposit int not null default 0,
    role varchar(20) not null default 'BUYER',
    primary key (id)
);

create table products (
    id varchar(64) not null,
    name varchar(128) not null,
    available_amount int not null default 0,
    cost int not null default 5,
    seller_id varchar(64) not null,
    primary key (id)
);

ALTER TABLE products
ADD CONSTRAINT FK_seller
FOREIGN KEY (seller_id) REFERENCES users(id); 