package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Todo Data Structure
type Todo struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"id"`
	Body      string             `json:"body" bson:"body"`
	Completed bool               `json:"completed" bson:"completed"`
}

var collection *mongo.Collection

func main() {

	// Load the environment variable
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	MONGODB_URI := os.Getenv("MONGODB_URI")

	clientOption := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOption)
	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	// Test connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MONGODB ATLAS")

	collection = client.Database("golang_db").Collection("todos")

	app := fiber.New()

	app.Get("/api/todos", getTodos)
	app.Post("/api/todos", createTodos)
	app.Patch("/api/todos/:id", editTodos)
	app.Delete("/api/todos/:id", deleteTodos)

	PORT := os.Getenv("PORT")

	err = app.Listen(fmt.Sprintf(":%v", PORT))
	if err != nil {
		log.Fatal(err)
	}
}

func getTodos(c *fiber.Ctx) error {
	var todos []Todo

	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var todo Todo

		err := cursor.Decode(&todo)
		if err != nil {
			log.Fatal(err)
		}

		todos = append(todos, todo)
	}

	return c.Status(200).JSON(todos)
}

func createTodos(c *fiber.Ctx) error {
	todo := new(Todo)

	err := c.BodyParser(todo)
	if err != nil {
		return err
	}

	if todo.Body == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "The Body can not be empty",
		})
	}

	insertedOne, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		return err
	}

	todo.ID = insertedOne.InsertedID.(primitive.ObjectID)

	return c.Status(201).JSON(todo)
}

func editTodos(c *fiber.Ctx) error {
	id := c.Params("id")

	primitiveObjectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid Todo ID",
		})
	}

	filter := bson.M{"_id": primitiveObjectID}
	update := bson.M{"$set": bson.M{"completed": true}}

	_, err = collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{
		"success": true,
	})
}

func deleteTodos(c *fiber.Ctx) error {
	id := c.Params("id")

	primitiveIdObject, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid id format",
		})
	}

	filter := bson.M{"_id": primitiveIdObject}

	deleteResult, err := collection.DeleteOne(context.Background(), filter)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete document",
		})
	}

	// Checking if any Documented was deleted
	if deleteResult.DeletedCount == 0{
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error":"Document not found",
		})
	}

	return c.JSON(fiber.Map{
		"success": "Document Deleted Successfully!",
	})
}
