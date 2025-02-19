package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"store/model"
	"store/view"
	"strconv"

	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client
var productCollection *mongo.Collection
var logger = logrus.New()

func init() {
	logFile, err := os.OpenFile("logging.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		return
	}

	logger.SetOutput(logFile)

	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logger.WithFields(logrus.Fields{
		"action": "initialize_logger",
		"status": "success",
	}).Info("Logger initialized and writing to logging.txt")
}

func InitializeProduct(mongoClient *mongo.Client) {
	client = mongoClient
	productCollection = client.Database("storeDB").Collection("products")
	log.WithFields(logrus.Fields{
		"action": "initialize",
		"status": "success",
	}).Info("Product collection initialized")
}

func validateProductFields(reqData map[string]interface{}, requiredFields []string) (map[string]string, bool) {
	for _, field := range requiredFields {
		if _, exists := reqData[field]; !exists {
			logger.Warnf("Validation failed: Field '%s' is required", field)
			return map[string]string{
				"status":  "fail",
				"message": fmt.Sprintf("Field '%s' is required", field),
			}, false
		}
	}

	if id, ok := reqData["id"].(float64); !ok {
		if idInt, ok := reqData["id"].(int); ok && idInt > 0 {
			id = float64(idInt)
		} else {
			logger.Warn("Validation failed: 'id' must be a positive number")
			return map[string]string{
				"status":  "fail",
				"message": "'id' must be a positive number",
			}, false
		}
	} else if id <= 0 {
		logger.Warn("Validation failed: 'id' must be a positive number")
		return map[string]string{
			"status":  "fail",
			"message": "'id' must be a positive number",
		}, false
	}

	if name, ok := reqData["name"].(string); !ok || name == "" {
		logger.Warn("Validation failed: 'name' must be a non-empty string")
		return map[string]string{
			"status":  "fail",
			"message": "'name' must be a non-empty string",
		}, false
	}

	if price, ok := reqData["price"].(float64); !ok || price <= 0 {
		logger.Warn("Validation failed: 'price' must be a positive number")
		return map[string]string{
			"status":  "fail",
			"message": "'price' must be a positive number",
		}, false
	}

	return nil, true
}

func getPaginationParams(r *http.Request) (int, int) {
	page := r.URL.Query().Get("page")
	limit := 0
	skip := 0

	if p, err := strconv.Atoi(page); err == nil && p > 1 {
		skip = (p - 1) * limit
	} else {
		page = "1"
	}
	return skip, limit
}
func getSortingParams(r *http.Request) (string, int) {
	sortField := r.URL.Query().Get("sort")
	var sortOrder int
	if sortField != "" {
		sortOrder = 1
		if sortField[0] == '-' {
			sortField = sortField[1:]
			sortOrder = -1
		}
	} else {
		sortField = "price"
		sortOrder = 1
	}
	return sortField, sortOrder
}

func AllProducts(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"action": "start_all_products",
	}).Info("Start AllProducts Handler")

	// Фильтрация по названию
	filterName := r.URL.Query().Get("name")
	filter := bson.M{}
	if filterName != "" {
		filter["name"] = bson.M{"$regex": filterName, "$options": "i"}
	}

	// Фильтрация по цене
	minPriceStr := r.URL.Query().Get("minPrice")
	maxPriceStr := r.URL.Query().Get("maxPrice")

	// Преобразуем строки в float64
	var minPrice, maxPrice float64
	var err error

	if minPriceStr != "" {
		minPrice, err = strconv.ParseFloat(minPriceStr, 64)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err.Error()}).Warn("Invalid minPrice format")
			http.Error(w, "Invalid minPrice format", http.StatusBadRequest)
			return
		}
	}

	if maxPriceStr != "" {
		maxPrice, err = strconv.ParseFloat(maxPriceStr, 64)
		if err != nil {
			log.WithFields(logrus.Fields{"error": err.Error()}).Warn("Invalid maxPrice format")
			http.Error(w, "Invalid maxPrice format", http.StatusBadRequest)
			return
		}
	}

	// Добавляем фильтр по цене, если переданы параметры
	priceFilter := bson.M{}
	if minPrice > 0 {
		priceFilter["$gte"] = minPrice
	}
	if maxPrice > 0 {
		priceFilter["$lte"] = maxPrice
	}
	if len(priceFilter) > 0 {
		filter["price"] = priceFilter
	}

	// Сортировка и пагинация
	sortField, sortOrder := getSortingParams(r)
	skip, limit := getPaginationParams(r)

	cursor, err := productCollection.Find(
		context.TODO(),
		filter,
		options.Find().SetSort(bson.D{{Key: sortField, Value: sortOrder}}).
			SetLimit(int64(limit)).
			SetSkip(int64(skip)),
	)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "all_products",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to fetch products from database")
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to fetch products from database"})
		return
	}
	defer cursor.Close(context.TODO())

	var products model.Products
	if err = cursor.All(context.TODO(), &products); err != nil {
		log.WithFields(logrus.Fields{
			"action": "all_products",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Error decoding product data")
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Error decoding product data"})
		return
	}

	view.RenderProducts(w, products)
	log.WithFields(logrus.Fields{
		"action": "all_products",
		"status": "success",
		"count":  len(products),
	}).Info("Fetched products successfully")
}

func HandleProductPostRequest(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"action": "start_create_product",
	}).Info("Start HandleProductPostRequest Handler")

	if r.Method != http.MethodPost {
		log.WithFields(logrus.Fields{
			"action": "method_not_allowed",
			"status": "fail",
			"method": r.Method,
		}).Warn("Only POST methods are allowed!")
		http.Error(w, "Only POST methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "invalid_json",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if resp, valid := validateProductFields(reqData, []string{"id", "name", "price"}); !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	newProduct := model.Product{
		ID:    int(reqData["id"].(float64)),
		Name:  reqData["name"].(string),
		Price: reqData["price"].(float64),
	}

	_, err = productCollection.InsertOne(context.TODO(), newProduct)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "create_product",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to create product")
		jsonResponse(w, http.StatusInternalServerError, map[string]string{"status": "fail", "message": "Failed to create product"})
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": "Product data successfully received",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.WithFields(logrus.Fields{
		"action": "create_product",
		"status": "success",
		"id":     newProduct.ID,
		"name":   newProduct.Name,
	}).Info("Successfully added new product")
}

func DeleteProductByID(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"action": "start_delete_product",
	}).Info("Start DeleteProductByID Handler")

	if r.Method != http.MethodDelete {
		log.WithFields(logrus.Fields{
			"action": "method_not_allowed",
			"status": "fail",
			"method": r.Method,
		}).Warn("Only DELETE methods are allowed!")
		http.Error(w, "Only DELETE methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "invalid_json",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, idExists := reqData["id"]
	if !idExists || id.(float64) <= 0 {
		log.WithFields(logrus.Fields{
			"action": "validation",
			"status": "fail",
			"field":  "id",
		}).Warn("Field 'id' is required and must be positive")
		response := map[string]string{
			"status":  "fail",
			"message": "Field 'id' is required and must be positive",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	idFloat := int(id.(float64))
	filter := bson.M{"id": idFloat}
	deleteResult, err := productCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "delete_product",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to delete product from database")
		http.Error(w, "Failed to delete product from the database", http.StatusInternalServerError)
		return
	}

	if deleteResult.DeletedCount == 0 {
		log.WithFields(logrus.Fields{
			"action": "product_not_found",
			"status": "fail",
			"id":     idFloat,
		}).Warn("No product found with the given ID")
		response := map[string]string{
			"status":  "fail",
			"message": fmt.Sprintf("No product found with ID %d", idFloat),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Product with ID %d successfully deleted", idFloat),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.WithFields(logrus.Fields{
		"action": "delete_product",
		"status": "success",
		"id":     idFloat,
	}).Info("Successfully deleted product")
}

func UpdateProductByID(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"action": "start_update_product",
		"status": "initiated",
	}).Info("Start: UpdateProductByID Handler")

	if r.Method != http.MethodPut {
		log.WithFields(logrus.Fields{
			"action": "method_not_allowed",
			"status": "fail",
		}).Warn("Only PUT methods are allowed!")
		http.Error(w, "Only PUT methods are allowed!", http.StatusMethodNotAllowed)
		return
	}

	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "invalid_json",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	if resp, valid := validateProductFields(reqData, []string{"id", "name", "price"}); !valid {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	id := int(reqData["id"].(float64))
	update := bson.M{
		"$set": bson.M{
			"name":  reqData["name"].(string),
			"price": reqData["price"].(float64),
		},
	}

	filter := bson.M{"id": id}
	result, err := productCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "update_product",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Failed to update product")
		http.Error(w, "Failed to update product", http.StatusInternalServerError)
		return
	}

	if result.MatchedCount == 0 {
		log.WithFields(logrus.Fields{
			"action": "product_not_found",
			"status": "fail",
			"id":     id,
		}).Warn("No product found with the given ID")
		http.Error(w, "No product found with the given ID", http.StatusNotFound)
		return
	}

	response := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Product with ID %d successfully updated", id),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	log.WithFields(logrus.Fields{
		"action": "update_product",
		"status": "success",
		"id":     id,
	}).Info("Successfully updated product")
}

func GetProductByID(w http.ResponseWriter, r *http.Request) {
	log.WithFields(logrus.Fields{
		"action": "start_get_product",
		"status": "initiated",
	}).Info("Start: GetProductByID Handler")

	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "invalid_json",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	id, idExists := reqData["id"]
	if !idExists || id.(float64) <= 0 {
		log.WithFields(logrus.Fields{
			"action": "validation",
			"status": "fail",
			"error":  "'id' must be positive",
		}).Warn("Validation failed")
		response := map[string]string{
			"status":  "fail",
			"message": "Field 'id' is required and must be positive",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	idFloat := int(id.(float64))
	filter := bson.M{"id": idFloat}
	var product model.Product
	err = productCollection.FindOne(context.TODO(), filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.WithFields(logrus.Fields{
				"action": "product_not_found",
				"status": "fail",
				"id":     idFloat,
			}).Warn("No product found with the given ID")
			http.Error(w, fmt.Sprintf("No product found with ID %d", idFloat), http.StatusNotFound)
			return
		}
		log.WithFields(logrus.Fields{
			"action": "fetch_product",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Error fetching product")
		http.Error(w, "Error fetching product from the database", http.StatusInternalServerError)
		return
	}

	view.RenderProducts(w, product)
	log.WithFields(logrus.Fields{
		"action": "fetch_product",
		"status": "success",
		"id":     idFloat,
	}).Info("Successfully fetched product")
}

func GetProductByName(w http.ResponseWriter, r *http.Request) {
	var reqData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "invalid_json",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Invalid JSON format")
		http.Error(w, "Invalid JSON format: "+err.Error(), http.StatusBadRequest)
		return
	}

	name, nameExists := reqData["name"]
	if !nameExists || name == "" {
		log.WithFields(logrus.Fields{
			"action": "validation",
			"status": "fail",
			"error":  "'name' is required and must be non-empty",
		}).Warn("Validation failed")
		response := map[string]string{
			"status":  "fail",
			"message": "Field 'name' is required and must be non-empty",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	filter := bson.M{"name": bson.M{"$regex": name.(string), "$options": "i"}}
	var product model.Product
	err = productCollection.FindOne(context.TODO(), filter).Decode(&product)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.WithFields(logrus.Fields{
				"action": "product_not_found",
				"status": "fail",
				"name":   name,
			}).Warn("No product found with the given name")
			http.Error(w, fmt.Sprintf("No product found with name %s", name), http.StatusNotFound)
			return
		}
		log.WithFields(logrus.Fields{
			"action": "fetch_product",
			"status": "fail",
			"error":  err.Error(),
		}).Error("Error fetching product")
		http.Error(w, "Error fetching product from the database", http.StatusInternalServerError)
		return
	}

	view.RenderProducts(w, product)
	log.WithFields(logrus.Fields{
		"action": "fetch_product",
		"status": "success",
		"name":   name,
	}).Info("Successfully fetched product")
}
