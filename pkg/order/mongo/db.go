package mongo

import (
	"context"
	"errors"
	"shopping-cart/pkg/db"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DbName         = "order"
	CollectionName = "orders"
)

type OrdersConnection struct {
	db.MongoConnection
	OrderCollection *mongo.Collection
	//ctx             context.Context
}

func Init(client *mongo.Client) *OrdersConnection {
	database := client.Database(DbName)
	//todo cancel fun
	return &OrdersConnection{
		MongoConnection: db.MongoConnection{
			Database: database,
			Client:   client,
		},
		OrderCollection: database.Collection(CollectionName),
	}
}

/*
Deletes an order with a specific id
@param id id in hex format without the '0x' prefix
*/
func (orderConn *OrdersConnection) DeleteOrder(id string) error {
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	query := bson.D{
		primitive.E{
			Key:   OrderId,
			Value: objId,
		},
	}
	res, err := orderConn.OrderCollection.DeleteOne(context.Background(), query)
	if err != nil {
		return err
	}
	if res.DeletedCount != 1 {
		return errors.New("document was not deleted")
	}
	return nil
}

/*
Create a new empty order for the user
The id is in hex format
*/
func (orderConn *OrdersConnection) EmptyOrder(userId string) (string, error) {
	objId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return "", err
	}

	order := Order{
		OrderId:   primitive.NewObjectID(),
		Paid:      false,
		Items:     []primitive.ObjectID{},
		UserId:    objId,
		TotalCost: 0,
	}

	res, err := orderConn.OrderCollection.InsertOne(context.Background(),
		order)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

/*
Find one order by id
the id is in hex format
*/
func (orderConn *OrdersConnection) FindOrder(id string) (*Order, error) {
	// TODO: Remove totalcost from this table, is retrieved from the stock service #Rahim
	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	query := bson.D{
		primitive.E{
			Key:   OrderId,
			Value: objId,
		},
	}
	res := orderConn.OrderCollection.FindOne(context.Background(), query)

	if res.Err() != nil {
		return nil, res.Err()
	}
	order := &Order{}
	err = res.Decode(order)
	if err != nil {
		return nil, err
	}
	return order, nil
}

/*
Add an item to an order
the ids are in hex format
*/
func (orderConn *OrdersConnection) AddItem(orderId string, itemId string) error {
	objOrderId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return err
	}

	objItemId, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return err
	}

	push := bson.D{
		primitive.E{
			Key: "$push",
			Value: bson.D{
				primitive.E{
					Key:   Items,
					Value: objItemId,
				},
			},
		},
	}

	query := bson.D{
		primitive.E{
			Key:   OrderId,
			Value: objOrderId,
		},
	}

	res, err := orderConn.OrderCollection.UpdateOne(context.Background(), query, push)

	if err != nil {
		return err
	}
	if res.ModifiedCount > 1 {
		return errors.New("updated multiple documents")
	}
	if res.ModifiedCount == 0 {
		return errors.New("updated 0 documents")
	}
	return nil
}

/*
Remove an item from an order
the ids are in hex format
*/
func (orderConn *OrdersConnection) RemoveItem(orderId string, itemId string) error {
	objOrderId, err := primitive.ObjectIDFromHex(orderId)
	if err != nil {
		return err
	}

	objItemId, err := primitive.ObjectIDFromHex(itemId)
	if err != nil {
		return err
	}
	//todo, only remove one item of the same id
	pull := bson.D{
		primitive.E{
			Key: "$pull",
			Value: bson.D{
				primitive.E{
					Key:   Items,
					Value: objItemId,
				},
			},
		},
	}
	//pull := bson.D{{"$unset", bson.D{{"items.$[]", objItemId}}}}

	query := bson.D{
		primitive.E{
			Key:   OrderId,
			Value: objOrderId,
		},
	}

	res, err := orderConn.OrderCollection.UpdateOne(context.Background(), query, pull)

	if err != nil {
		return err
	}
	if res.ModifiedCount > 1 {
		return errors.New("updated multiple documents")
	}
	//todo think about this, I guess we probably want to delete if it is there and not error when it isn't there
	//if res.ModifiedCount == 0 {
	//	return errors.New("updated 0 documents")
	//}
	return nil
}
