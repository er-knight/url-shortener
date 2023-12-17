package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB

type addURLRequest struct {
	URL string `json:"URL"`
}

type URLModel struct {
	URL string
	Id  string
}

func (URLModel) TableName() string {
	return "urls"
}

func connectDB() error {
	var dsn = fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"),
	)
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	return nil
}

func generateId() (string, error) {
	// reference: https://encore.dev/docs/tutorials/rest-api#1-create-a-service-and-endpoint

	var data [6]byte // 6 bytes of entropy
	var _, err = rand.Read(data[:])
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data[:]), nil
}

func addURL(c *gin.Context) {

	var request addURLRequest

	{
		var err = c.ShouldBindJSON(&request)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}
	}

	var id, err = generateId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	db.Create(URLModel{
		URL: request.URL,
		Id:  id,
	})

	c.JSON(200, gin.H{
		"URL": request.URL,
		"ID":  id,
	})
}

func getURL(c *gin.Context) {
	var id, exists = c.Params.Get("id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Id not provided",
		})
		return
	}

	var URLRecord URLModel
	var err = db.First(&URLRecord, "id = ?", id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"URL": URLRecord.URL,
	})
}

func main() {

	var err = godotenv.Load()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = connectDB()
	if err != nil {
		log.Fatal(err.Error())
	}

	err = db.AutoMigrate(&URLModel{})
	if err != nil {
		log.Fatal(err.Error())
	}

	r := gin.Default()

	r.GET("/url/:id", getURL)
	r.POST("/url", addURL)

	r.Run()
}
