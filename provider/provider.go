package provider

import (
	"DB-worker-test/models"
	"context"
	"database/sql"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"

	_ "github.com/lib/pq"
	stan "github.com/nats-io/go-nats-streaming"
)

// Provider interface
type Provider interface {
	Open(driverName, dpURL string) error
	GetConn() (*sql.DB, error)
	GetConnMongo() *mongo.Database
	GetNatsConnectionStreaming() (stan.Conn, error)
}

type provider struct {
	snats       stan.Conn
	db          *sql.DB
	clientMongo *mongo.Client
	dbNameMongo string
	csMongo     string
	cs          string
	idlConns    int
	openConns   int
	lifetime    time.Duration
}

func New(db *models.SQLDataBase, mongoDB *models.NoSQLDataBase) Provider {
	info := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		db.Server, db.Port, db.UserID, db.Password, db.Database)
	infoMongo := fmt.Sprintf("mongodb://%s:%s@%s:%d/?authSource=admin",
		mongoDB.UserID, mongoDB.Password, mongoDB.Server, mongoDB.Port)
	return &provider{
		cs:          info,
		csMongo:     infoMongo,
		dbNameMongo: mongoDB.Database,
		idlConns:    db.MaxIdleConns,
		openConns:   db.MaxOpenConns,
		lifetime:    time.Duration(db.ConnMaxLifetime),
	}
}

// метод для возвращения коннекшена к NATS streaming
func (p *provider) GetNatsConnectionStreaming() (stan.Conn, error) {
	return p.snats, nil
}

// Open connection
func (p *provider) Open(driverName, dpURL string) error {
	var err error
	p.db, err = sql.Open(driverName, p.cs)
	if err != nil {
		return err
	}
	p.db.SetMaxIdleConns(p.idlConns)
	p.db.SetMaxOpenConns(p.openConns)
	p.db.SetConnMaxLifetime(p.lifetime)

	client, err := mongo.NewClient(options.Client().ApplyURI(p.csMongo))
	if err != nil {
		fmt.Println(err)
	}

	// Create connect
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatalf("Can't connect to mongo: %v", err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatalf("Can't ping to mongo: %v", err)
	}

	p.clientMongo = client

	clientID := "nats_1"
	snc, err := stan.Connect("test-cluster", clientID, stan.NatsURL(dpURL), stan.MaxPubAcksInflight(1))
	if err != nil {
		log.Fatalf("Can't connect: %v.\nMake sure a NATS Streaming Server is running at: %s", err, dpURL)
		return err
	}
	p.snats = snc
	fmt.Printf(`{"clientID": %s}`, clientID)

	return nil
}

func (p *provider) GetConn() (*sql.DB, error) {
	return p.db, nil
}

func (p *provider) GetConnMongo() *mongo.Database {
	return p.clientMongo.Database(p.dbNameMongo)
}
