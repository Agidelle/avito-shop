CREATE DATABASE IF NOT EXISTS Avito;
CREATE DATABASE IF NOT EXISTS test_db;

GRANT ALL PRIVILEGES ON Avito.* TO 'user'@'%';
GRANT ALL PRIVILEGES ON test_db.* TO 'user'@'%';


USE Avito;
CREATE TABLE IF NOT EXISTS users (
                                     id INT AUTO_INCREMENT PRIMARY KEY,
                                     username VARCHAR(255) UNIQUE NOT NULL,
                                     password_hash VARCHAR(255) NOT NULL,
                                     coins INT DEFAULT 1000
);
CREATE TABLE IF NOT EXISTS transactions (
                                            id INT AUTO_INCREMENT PRIMARY KEY,
                                            from_user_id INT,
                                            to_user_id INT NOT NULL,
                                            amount INT NOT NULL,
                                            FOREIGN KEY (from_user_id) REFERENCES users(id),
                                            FOREIGN KEY (to_user_id) REFERENCES users(id)
);
CREATE TABLE IF NOT EXISTS inventory (
                                         id INT AUTO_INCREMENT PRIMARY KEY,
                                         user_id INT NOT NULL,
                                         item_name VARCHAR(255) NOT NULL,
                                         quantity INT DEFAULT 0,
                                         FOREIGN KEY (user_id) REFERENCES users(id),
                                         UNIQUE unique_user_item (user_id, item_name)
);

USE test_db;
CREATE TABLE IF NOT EXISTS users (
                                     id INT AUTO_INCREMENT PRIMARY KEY,
                                     username VARCHAR(255) UNIQUE NOT NULL,
                                     password_hash VARCHAR(255) NOT NULL,
                                     coins INT DEFAULT 1000
);
CREATE TABLE IF NOT EXISTS transactions (
                                            id INT AUTO_INCREMENT PRIMARY KEY,
                                            from_user_id INT,
                                            to_user_id INT NOT NULL,
                                            amount INT NOT NULL
);
CREATE TABLE IF NOT EXISTS inventory (
                                         id INT AUTO_INCREMENT PRIMARY KEY,
                                         user_id INT NOT NULL,
                                         item_name VARCHAR(255) NOT NULL,
                                         quantity INT DEFAULT 0,
                                         UNIQUE unique_user_item (user_id, item_name)
);
USE Avito;