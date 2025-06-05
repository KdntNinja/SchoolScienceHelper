package utils

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID       interface{} `bson:"_id,omitempty"`
	Username string      `bson:"username"`
	Password string      `bson:"password"`
}

func CreateUser(ctx context.Context, username, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = UsersCollection.InsertOne(ctx, bson.M{"username": username, "password": string(hash)})
	if mongo.IsDuplicateKeyError(err) {
		return errors.New("username already exists")
	}
	return err
}

func AuthenticateUser(ctx context.Context, username, password string) (bool, error) {
	var user User
	err := UsersCollection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return false, nil
	}
	return true, nil
}
