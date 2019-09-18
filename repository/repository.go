package repository

import (
	"DB-worker-test/models"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/proto"
	stan "github.com/nats-io/go-nats-streaming"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io/ioutil"
	"log"
	"time"

	"DB-worker-test/provider"
)

// Repository interface
type Repository interface {
	getNatsStreamingConn() (stan.Conn, error)
	publishMessage(dataToSend []byte, subjectTheme string) error
	subscribeSimple(subject string, handler func(m *stan.Msg)) error
	subscribeQueue(subject string, handler func(m *stan.Msg)) error
	GetNewsHandler()
	CreateNewsHandler()
	SubscribeNewsGet()
	SubscribeNewsCreate()
	DatabaseGetNews(newsID int32) ([]byte, error)
	DatabaseCreateNews(title, date string) ([]byte, error)
	MongoDatabaseGetNews(title string) ([]byte, error)
	MongoDatabaseGetAllNews() ([]byte, error)
	MongoDatabaseCreateOneNews(news models.News) ([]byte, error)
	MongoDatabaseCreateManyNews(news []interface{}) ([]byte, error)
	MongoUpdateNews(oldTitle string, newTitle string) ([]byte, error)
}

type repository struct {
	provider   provider.Provider
	timeout    time.Duration
	newsCreate chan []byte
	newsGet    chan []byte
}

// New func
// NewNats
func New(pr provider.Provider) Repository {
	rep := &repository{
		provider:   pr,
		newsCreate: make(chan []byte, 100),
		newsGet:    make(chan []byte, 100),
	}
	err := rep.initDB()
	if err != nil {
		log.Fatalf("initialization failed - %s", err.Error())
	}
	return rep
}
func (r *repository) initDB() error {
	query, err := ioutil.ReadFile("init.sql")
	if err != nil {
		panic(err)
	}
	db, err := r.provider.GetConn()
	if err != nil {
		return err
	}
	if _, err := db.Exec(string(query)); err != nil {
		return err
	}
	fmt.Println("initialization completed successfully")
	return nil
}

func (r *repository) getNatsStreamingConn() (stan.Conn, error) {
	return r.provider.GetNatsConnectionStreaming()
}

func (r *repository) publishMessage(dataToSend []byte, subjectTheme string) error {
	nc, err := r.getNatsStreamingConn()
	if err != nil {
		return err
	}
	if err := nc.Publish(subjectTheme, dataToSend); err != nil {
		return err
	}
	return nil
}

func (r *repository) subscribeSimple(subject string, handler func(m *stan.Msg)) error {
	nats, err := r.getNatsStreamingConn()
	if err != nil {
		return err
	}
	optErrorsWorker := []stan.SubscriptionOption{
		stan.DurableName("remember" + subject),
		stan.MaxInflight(1),
		stan.SetManualAckMode()}

	_, err = nats.Subscribe(subject, handler, optErrorsWorker...)

	if err != nil {
		return err
	}
	return nil
}

func (r *repository) subscribeQueue(subject string, handler func(m *stan.Msg)) error {
	nats, err := r.getNatsStreamingConn()
	if err != nil {
		return err
	}
	optErrorsWorker := []stan.SubscriptionOption{
		stan.DurableName("remember" + subject),
		stan.MaxInflight(1),
		stan.SetManualAckMode()}

	queue := fmt.Sprintf("%s_queue", subject)
	_, err = nats.QueueSubscribe(subject, queue, handler, optErrorsWorker...)

	if err != nil {
		return err
	}
	return nil
}

func (r *repository) SubscribeNews(subject string, channel chan []byte) {
	handlerErrLog := func(m *stan.Msg) {
		defer func() {
			err := m.Ack()
			if err != nil {
				fmt.Println("err", err)
			}
			fmt.Println("get", subject)
			select {
			case channel <- m.Data:
			default:
			}
		}()
	}
	r.subscribeQueue(subject, handlerErrLog)
}

func (r *repository) SubscribeNewsGet() {
	subject := "create_news"
	fmt.Println("subscribe", subject)
	r.SubscribeNews(subject, r.newsCreate)
}

func (r *repository) SubscribeNewsCreate() {
	subject := "get_news"
	fmt.Println("subscribe", subject)
	r.SubscribeNews(subject, r.newsGet)
}

func (r *repository) GetNewsHandler() {
	for news := range r.newsGet {
		var data models.Id
		err := proto.Unmarshal(news, &data)
		if err != nil {
			fmt.Println("err", err)
			continue
		}

		// go to base
		res, err := r.DatabaseGetNews(data.Id)
		if err != nil {
			fmt.Println("err", err)
			continue
		}

		resp := models.Response{
			Response: res,
			Request:  string(news),
		}

		dataPb, err := proto.Marshal(&resp)
		if err != nil {
			fmt.Println("err", err)
			continue
		}

		err = r.publishMessage(dataPb, "news_get_resp")
		if err != nil {
			fmt.Println("err", err)
			continue
		}
	}
}

func (r *repository) CreateNewsHandler() {
	for news := range r.newsCreate {
		var data models.Data
		err := proto.Unmarshal(news, &data)
		if err != nil {
			fmt.Println("err", err)
			continue
		}
		// go to base
		res, err := r.DatabaseCreateNews(data.Title, data.Date)
		if err != nil {
			fmt.Println("err", err)
			continue
		}

		resp := models.Response{
			Response: res,
			Request:  string(news),
		}

		dataPb, err := proto.Marshal(&resp)
		if err != nil {
			fmt.Println("err", err)
			continue
		}

		err = r.publishMessage(dataPb, "news_create_resp")
		if err != nil {
			fmt.Println("err", err)
			continue
		}
	}
}

func (r *repository) DatabaseGetNews(newsID int32) ([]byte, error) {
	db, err := r.provider.GetConn()
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("select * from get_news($1)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var res []byte
	if err := stmt.QueryRow(newsID).Scan(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repository) DatabaseCreateNews(title, date string) ([]byte, error) {
	db, err := r.provider.GetConn()
	if err != nil {
		return nil, err
	}
	stmt, err := db.Prepare("select create_news($1, $2)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	var res []byte
	if err := stmt.QueryRow(title, date).Scan(&res); err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repository) MongoDatabaseGetNews(title string) ([]byte, error) {
	collection := r.provider.GetConnMongo().Collection("news")
	filter := bson.M{"title": title}

	var result models.News

	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		return nil, err
	}

	res, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repository) MongoDatabaseCreateManyNews(news []interface{}) ([]byte, error) {
	collection := r.provider.GetConnMongo().Collection("news")
	insertManyResult, err := collection.InsertMany(context.TODO(), news)
	if err != nil {
		return nil, err
	}

	fmt.Println("Inserted multiple documents: ", insertManyResult.InsertedIDs)

	res, err := json.Marshal(insertManyResult)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *repository) MongoDatabaseCreateOneNews(news models.News) ([]byte, error) {
	collection := r.provider.GetConnMongo().Collection("news")
	fmt.Println(collection)
	insertResult, err := collection.InsertOne(context.TODO(), news)
	if err != nil {
		return nil, err
	}

	fmt.Println("Inserted a single document: ", insertResult.InsertedID)

	res, err := json.Marshal(insertResult)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repository) MongoUpdateNews(oldTitle string, newTitle string) ([]byte, error) {
	collection := r.provider.GetConnMongo().Collection("news")
	filter := bson.M{"title": oldTitle}

	update := bson.D{
		{"$set", bson.D{
			{"title", newTitle},
		}},
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return nil, err
	}

	res, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *repository) MongoDatabaseGetAllNews() ([]byte, error) {
	collection := r.provider.GetConnMongo().Collection("news")
	options := options.Find()
	filter := bson.M{}
	cur, err := collection.Find(context.TODO(), filter, options)
	if err != nil {
		return nil, err
	}
	var results []models.News

	for cur.Next(context.TODO()) {
		var elem models.News
		err := cur.Decode(&elem)
		if err != nil {
			return nil, err
		}

		results = append(results, elem)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	cur.Close(context.TODO())

	res, err := json.Marshal(results)
	if err != nil {
		return nil, err
	}

	return res, nil
}
