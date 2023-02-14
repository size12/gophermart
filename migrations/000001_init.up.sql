CREATE TABLE users (
    id INT GENERATED BY DEFAULT AS IDENTITY,
    login VARCHAR(255), passw VARCHAR(255),
    balance FLOAT,
    withdrawn FLOAT
);

CREATE TABLE orders (
    userid INT, num BIGINT,
    stat VARCHAR(255),
    accrual FLOAT,
    uploaded TIMESTAMP
);

CREATE TABLE withdrawals (
    userid INT,
    num BIGINT,
    amount FLOAT,
    processed TIMESTAMP
);