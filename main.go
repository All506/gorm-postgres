package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	"gorm-postgres/models"
	"gorm-postgres/storage"
)

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	book := Book{}

	err := context.BodyParser(&book)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"Message": "Request Failed"})
		return err
	}

	creationErr := r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"Message": "Couldnt create new book instance"})
		return creationErr
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{"message": "Book entity has been created"})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"Message": "Couldnt retrieve any books"})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Books fetch succesfully",
		"data":    bookModels,
	})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{"message": "Id is empty"})
		return nil
	}

	err := r.DB.Delete(bookModel, id)

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{"message": "Coudln't delete book"})
		return err.Error
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Book has been deleted",
	})
	return nil
}

func (r *Repository) GetBookById(context *fiber.Ctx) error {
	id := context.Params("id")
	bookModel := &models.Books{}

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(&fiber.Map{
			"message": "Id is empty",
		})
		return nil
	}

	fmt.Println("Looking for book with id: ", id)

	err := r.DB.Where("id = ?", id).First(bookModel).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(&fiber.Map{
			"message": "Couldnt retrieve info from book",
		})
		return err
	}

	context.Status(http.StatusOK).JSON(&fiber.Map{
		"message": "Book fetch succesfully",
		"data":    bookModel,
	})
	return nil
}

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

func (r *Repository) SetUpRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_books/:id", r.GetBookById)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Password: os.Getenv("DB_PASS"),
		User:     os.Getenv("DB_USER"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("Couldnt load the database")
	}

	bookMigrationErr := models.MigrateBooks(db)

	if bookMigrationErr != nil {
		log.Fatal("Couldnt migrate db")
	}

	r := Repository{
		DB: db,
	}

	app := fiber.New()
	r.SetUpRoutes(app)
	app.Listen(":8000")
}
