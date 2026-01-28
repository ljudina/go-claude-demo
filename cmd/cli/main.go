package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"claude-test/internal/database"
	"claude-test/internal/repository"
	"claude-test/internal/service"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "users.db")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		fmt.Fprintf(os.Stderr, "failed to run migrations: %v\n", err)
		os.Exit(1)
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)

	command := os.Args[1]
	args := os.Args[2:]

	switch command {
	case "user":
		if err := userCmd(userService, args); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: cli <command> <subcommand> [options]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  user    Manage users")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  cli user list")
	fmt.Println("  cli user create --email user@example.com --name \"John Doe\"")
	fmt.Println("  cli user get --id 1")
	fmt.Println("  cli user update --id 1 --email new@example.com --name \"Jane Doe\"")
	fmt.Println("  cli user delete --id 1")
}
