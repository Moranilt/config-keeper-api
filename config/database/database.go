package database

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBCreds struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"db_name"`
}

func New(ctx context.Context, dbCreds *DBCreds) (*mongo.Client, *mongo.Database, error) {
	client, err := mongo.Connect(ctx, options.Client().SetAuth(options.Credential{
		Username: dbCreds.Username,
		Password: dbCreds.Password,
	}).ApplyURI(makeURI(dbCreds)))
	if err != nil {
		return nil, nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, nil, err
	}

	database := client.Database(dbCreds.DBName)

	return client, database, nil
}

func makeURI(dbCreds *DBCreds) string {
	return fmt.Sprintf("mongodb://%s:%s", dbCreds.Host, dbCreds.Port)
}

type Checker struct {
	db *mongo.Client
}

func MakeChecker(db *mongo.Client) *Checker {
	return &Checker{
		db: db,
	}
}

func (c *Checker) Check(ctx context.Context) error {
	return c.db.Ping(ctx, nil)
}
