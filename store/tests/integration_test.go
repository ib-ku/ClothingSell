package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"store/controller"
	"store/database"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestAllProductss(t *testing.T) {
	// Подключаем тестовую базу данных
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		t.Fatalf("Ошибка подключения к тестовой базе: %v", err)
	}

	// Проверяем соединение
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		t.Fatalf("Ошибка при проверке соединения с базой: %v", err)
	}

	// Назначаем тестовую коллекцию
	database.ProductCollection = client.Database("test_db").Collection("test_products")

	// Добавляем тестовый продукт
	testProduct := bson.M{"name": "Test Product", "price": 100}
	_, err = database.ProductCollection.InsertOne(context.TODO(), testProduct)
	if err != nil {
		t.Fatalf("Ошибка вставки тестовых данных: %v", err)
	}

	// Создаём тестовый HTTP-запрос
	req, err := http.NewRequest("GET", "/allproducts", nil)
	if err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	handler := http.HandlerFunc(controller.AllProducts)
	handler.ServeHTTP(rec, req)

	// Проверяем код ответа
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", rec.Code)
	}

	// Очищаем тестовую базу после теста
	_, err = database.ProductCollection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		t.Fatalf("Ошибка при очистке тестовых данных: %v", err)
	}

	// Отключаем тестовую базу
	err = client.Disconnect(context.TODO())
	if err != nil {
		t.Fatalf("Ошибка при отключении от базы данных: %v", err)
	}
}
