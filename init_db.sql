DROP TABLE users;
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    id_tele     BIGINT NOT NULL,
    id_chat     BIGINT,
    uname       VARCHAR(50),
    fname       VARCHAR(50),
    lname       VARCHAR(50),
    phone       VARCHAR(15),
    visit       TIMESTAMP WITHOUT TIME ZONE,
    is_admin    BOOLEAN NOT NULL DEFAULT FALSE,
    is_banned   BOOLEAN NOT NULL DEFAULT FALSE
);
DROP TABLE orders;
CREATE TABLE orders (
    id          SERIAL PRIMARY KEY,
    id_customer BIGINT,
    frame       INTEGER, 
    net         INTEGER, 
    datetime    TIMESTAMP WITHOUT TIME ZONE,
    sizes       VARCHAR(128),
    is_pickup   BOOLEAN NOT NULL DEFAULT FALSE,
    id_worker   BIGINT,
    assigned    TIMESTAMP WITHOUT TIME ZONE
);


DROP Function UpdateOrder;
CREATE FUNCTION UpdateOrder(_id_customer bigint, _frame integer, _net integer, _sizes varchar(128), _is_pickup boolean)
RETURNS INTEGER
LANGUAGE plpgsql AS
$func$
DECLARE
    _ID integer;
BEGIN
    SELECT id INTO _ID FROM orders WHERE id_customer=_id_customer AND frame=_frame AND net=_net LIMIT 1;
    IF _ID > 0 THEN
        UPDATE orders SET sizes=_sizes, is_pickup=_is_pickup, datetime=Now()
        WHERE id=_ID;
    ELSE
        INSERT INTO orders 
        (id_customer, frame, net, sizes, is_pickup, datetime) 
        VALUES
        (_id_customer, _frame, _net, _sizes, _is_pickup, Now()) 
        RETURNING id INTO _ID;
    END IF;
    RETURN _ID;
END
$func$;