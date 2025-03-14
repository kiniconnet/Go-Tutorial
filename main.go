package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

// Todo Data Structure
type Todo struct {
	ID        int    `json:"id"`
	Task      string `json:"task"`
	Completed bool   `json:"completed"`
}

func main() {
	app := fiber.New()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading dot env file.")
	}

	PORT := os.Getenv("PORT")

	todos := []Todo{}

	// Get all the todos
	app.Get("/api/todos", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(todos)
	})

	//Adding a new todo
	app.Post("/api/todos", func(c *fiber.Ctx) error {
		todo := &Todo{}

		err := c.BodyParser(todo)
		if err != nil {
			return err
		}

		if todo.Task == "" {
			return c.Status(400).JSON(fiber.Map{
				"error": "The Task can not be empty",
			})
		}

		todo.ID = len(todos) + 1
		todos = append(todos, *todo)

		return c.Status(201).JSON(todos)
	})

	// Edit a todo
	app.Patch("/api/todos/:id", func(c *fiber.Ctx) error {
		id := c.Params("id")

		IntId, err := strconv.Atoi(id)
		if err != nil {
			return err
		}

		for i, todo := range todos {

			if todo.ID == IntId {
				todos[i].Completed = true
				return c.Status(200).JSON(todos[i])
			}
		}
		return c.Status(400).JSON(fiber.Map{
			"error": "Task not found",
		})
	})

	// Delete a todo
	app.Delete("/api/todos/:id", func(c *fiber.Ctx) error {
		// Get the requested Id
		id := c.Params("id")

		// Convert the id to an int
		IntId, _ := strconv.Atoi(id)

		// Range over all the Id's
		for i, todo := range todos {
			if todo.ID == IntId {
				// Remove the todo item
				todos = append(todos[:i], todos[i+1:]...)

				// Log the deletion
				log.Printf("Deleted todo with ID: %d", IntId)

				// Return success response
				return c.Status(fiber.StatusOK).JSON(fiber.Map{
					"success": true,
					"message": "Todo deleted successfully",
				})
			}
		}

		return c.Status(400).JSON(fiber.Map{
			"error": "Todo not found",
		})
	})

	app.Listen(fmt.Sprintf(":%s", PORT))
}
