package service

import (
	"DB-worker-test/models"
	"DB-worker-test/repository"
)

// Service interface
type Service interface {
	Close()
	Subscribe()
	GetNews(title string) ([]byte, error)
	GetNewsAll() ([]byte, error)
	CreateOneNews(news models.News) ([]byte, error)
	CreateManyNews(news []models.News) ([]byte, error)
	UpdateNews(oldTitle string, newTitle string) ([]byte, error)
}

type service struct {
	repNatsStreaming repository.Repository
}

// New return new service
func New(rNatsStreaming repository.Repository) Service {
	return &service{
		repNatsStreaming: rNatsStreaming,
	}
}

func (s *service) Close() {}

func (s *service) Subscribe() {
	s.repNatsStreaming.SubscribeNewsCreate()
	s.repNatsStreaming.SubscribeNewsGet()
	go s.repNatsStreaming.GetNewsHandler()
	go s.repNatsStreaming.CreateNewsHandler()
}

func (s *service) GetNews(title string) ([]byte, error) {
	return s.repNatsStreaming.MongoDatabaseGetNews(title)
}
func (s *service) CreateOneNews(news models.News) ([]byte, error) {
	return s.repNatsStreaming.MongoDatabaseCreateOneNews(news)
}
func (s *service) CreateManyNews(news []models.News) ([]byte, error) {
	var newsNew []interface{}
	for _, newsItem := range news {
		newsNew = append(newsNew, newsItem)
	}
	return s.repNatsStreaming.MongoDatabaseCreateManyNews(newsNew)
}
func (s *service) UpdateNews(oldTitle string, newTitle string) ([]byte, error) {
	return s.repNatsStreaming.MongoUpdateNews(oldTitle, newTitle)
}

func (s *service) GetNewsAll() ([]byte, error) {
	return s.repNatsStreaming.MongoDatabaseGetAllNews()
}
