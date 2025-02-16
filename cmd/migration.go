/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"database/sql"
	"fmt"

	"avito-shop/internal/config"

	"github.com/spf13/cobra"
)

// migrationCmd represents the migration command
var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.MustLoad(cfgFile)
		if err != nil {
			return err
		}

		connect := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", "root", "5555", cfg.Host, cfg.Port, cfg.Database)
		db, err := sql.Open("mysql", connect)
		if err != nil {
			fmt.Println("Connection Error")
			return err
		}

		err = db.Ping()
		if err != nil {
			fmt.Println("Connection Error")
			return err
		}

		queries := []string{
			`CREATE TABLE IF NOT EXISTS users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            username VARCHAR(255) UNIQUE NOT NULL,
    		password_hash VARCHAR(255) NOT NULL,
            coins INT DEFAULT 1000
        );`,
			`CREATE TABLE IF NOT EXISTS transactions (
            id INT AUTO_INCREMENT PRIMARY KEY,
            from_user_id INT,
            to_user_id INT NOT NULL,
            amount INT NOT NULL,
            FOREIGN KEY (from_user_id) REFERENCES users(id),
            FOREIGN KEY (to_user_id) REFERENCES users(id)
        );`,
			`CREATE TABLE IF NOT EXISTS inventory (
            id INT AUTO_INCREMENT PRIMARY KEY,
            user_id INT NOT NULL,
            item_name VARCHAR(255) NOT NULL,
            quantity INT DEFAULT 0,
            FOREIGN KEY (user_id) REFERENCES users(id),
    		UNIQUE unique_user_item (user_id, item_name)
        );`,
			//`GRANT ALL PRIVILEGES ON test_db.* TO 'user'@'%';`,
			`CREATE DATABASE IF NOT EXISTS test_db;`,
			`GRANT ALL PRIVILEGES ON test_db.* TO 'user'@'%';`,
			`USE test_db;`,
			`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			coins INT DEFAULT 1000
		);`,
			`CREATE TABLE IF NOT EXISTS transactions (
			id INT AUTO_INCREMENT PRIMARY KEY,
			from_user_id INT,
			to_user_id INT NOT NULL,
			amount INT NOT NULL
);`,
			`CREATE TABLE IF NOT EXISTS inventory (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			item_name VARCHAR(255) NOT NULL,
			quantity INT DEFAULT 0,
			UNIQUE unique_user_item (user_id, item_name)
);`,
		}

		for _, query := range queries {
			_, err = db.Exec(query)
			if err != nil {
				fmt.Println("Query Error")
				fmt.Println(err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(migrationCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// migrationCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// migrationCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
