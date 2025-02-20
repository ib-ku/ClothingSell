package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitializeAdmin(client *mongo.Client) {
	transactionCollection = client.Database("storeDB").Collection("transactions")
	productCollection = client.Database("storeDB").Collection("products")
}

func GetAdminMetrics(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Общее количество продаж
	totalSales, _ := transactionCollection.CountDocuments(ctx, bson.M{})

	// 2. Общая сумма выручки
	var revenueResult []bson.M
	cursor, err := transactionCollection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": nil, "total": bson.M{"$sum": "$price"}}}},
	})
	if err == nil {
		_ = cursor.All(ctx, &revenueResult)
	}
	totalRevenue := 0.0
	if len(revenueResult) > 0 {
		totalRevenue, _ = revenueResult[0]["total"].(float64)
	}

	// 3. Самый продаваемый товар
	var bestSelling []bson.M
	cursor, err = transactionCollection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": "$product_name", "count": bson.M{"$sum": 1}}}},
		{{Key: "$sort", Value: bson.M{"count": -1}}},
		{{Key: "$limit", Value: 1}},
	})
	if err == nil {
		_ = cursor.All(ctx, &bestSelling)
	}
	bestSellingProduct := ""
	if len(bestSelling) > 0 {
		bestSellingProduct, _ = bestSelling[0]["_id"].(string)
	}

	// 4. Самый непопулярный товар
	var leastSelling []bson.M
	cursor, err = transactionCollection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": "$product_name", "count": bson.M{"$sum": 1}}}},
		{{Key: "$sort", Value: bson.M{"count": 1}}},
		{{Key: "$limit", Value: 1}},
	})
	if err == nil {
		_ = cursor.All(ctx, &leastSelling)
	}
	leastSellingProduct := ""
	if len(leastSelling) > 0 {
		leastSellingProduct, _ = leastSelling[0]["_id"].(string)
	}

	// 5. Самый просматриваемый товар
	var mostViewed bson.M
	err = productCollection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"views": -1})).Decode(&mostViewed)
	mostViewedProduct := ""
	if err == nil {
		mostViewedProduct, _ = mostViewed["name"].(string)
	}

	// 6. Самый редко просматриваемый товар
	var leastViewed bson.M
	err = productCollection.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.M{"views": 1})).Decode(&leastViewed)
	leastViewedProduct := ""
	if err == nil {
		leastViewedProduct, _ = leastViewed["name"].(string)
	}

	// 7. Средняя цена продаваемых товаров
	var avgPriceResult []bson.M
	cursor, err = transactionCollection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": nil, "avg": bson.M{"$avg": "$price"}}}},
	})
	if err == nil {
		_ = cursor.All(ctx, &avgPriceResult)
	}
	averagePrice := 0.0
	if len(avgPriceResult) > 0 {
		averagePrice, _ = avgPriceResult[0]["avg"].(float64)
	}

	// 8. Число повторных покупателей
	var repeatCustomers []bson.M
	cursor, err = transactionCollection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": "$user_id", "count": bson.M{"$sum": 1}}}},
		{{Key: "$match", Value: bson.M{"count": bson.M{"$gt": 1}}}},
		{{Key: "$count", Value: "count"}},
	})
	if err == nil {
		_ = cursor.All(ctx, &repeatCustomers)
	}
	repeatCustomersCount := 0
	if len(repeatCustomers) > 0 {
		repeatCustomersCount = int(repeatCustomers[0]["count"].(int32)) // Приведение int32 -> int
	}

	// 9. Средний чек (ARPU)
	var avgOrderResult []bson.M
	cursor, err = transactionCollection.Aggregate(ctx, mongo.Pipeline{
		{{Key: "$group", Value: bson.M{"_id": "$user_id", "totalSpent": bson.M{"$sum": "$price"}}}},
		{{Key: "$group", Value: bson.M{"_id": nil, "avg": bson.M{"$avg": "$totalSpent"}}}},
	})
	if err == nil {
		_ = cursor.All(ctx, &avgOrderResult)
	}
	averageOrderValue := 0.0
	if len(avgOrderResult) > 0 {
		averageOrderValue, _ = avgOrderResult[0]["avg"].(float64)
	}

	// 10. Число уникальных покупателей
	uniqueBuyers, _ := transactionCollection.Distinct(ctx, "user_id", bson.M{})

	// Возвращаем JSON с данными
	response := map[string]interface{}{
		"totalSales":         int(totalSales), // Приведение int64 -> int
		"totalRevenue":       totalRevenue,
		"bestSellingProduct": bestSellingProduct,
		"leastSellingProduct": leastSellingProduct,
		"mostViewedProduct":  mostViewedProduct,
		"leastViewedProduct": leastViewedProduct,
		"averagePrice":       averagePrice,
		"repeatCustomers":    repeatCustomersCount,
		"averageOrderValue":  averageOrderValue,
		"uniqueBuyers":       len(uniqueBuyers),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
