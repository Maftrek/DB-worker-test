package application

import (
	"DB-worker-test/models"
	"bytes"
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

// хелф хендлер для кубера
func (app *Application) HealthHandler(w http.ResponseWriter, r *http.Request) {
	sum := models.GetHashSum()
	if !bytes.Equal(sum, app.hashSum) {
		app.logger.Log("msg", "New Configuration")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (app *Application) getNews(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	news, err := app.svc.GetNews(vars["title"])
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = w.Write(news)
	if err != nil {
		app.logger.Log("err", err)
	}
}

func (app *Application) updateNews(w http.ResponseWriter, r *http.Request) {
	newsInfo, err := ioutil.ReadAll(r.Body)
	if err != nil {
		app.logger.Log("err", err, "body", r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}

	defer r.Body.Close()
	// инициализируем переменную для получения информации по поставщику
	var news models.NewsUpdate
	// анмаршалим в структуру
	err = json.Unmarshal(newsInfo, &news)
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	resp, err := app.svc.UpdateNews(news.TitleOld, news.TitleNew)
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		app.logger.Log("err", err)
	}
}

func (app *Application) createOneNews(w http.ResponseWriter, r *http.Request) {
	newsInfo, err := ioutil.ReadAll(r.Body)
	if err != nil {
		app.logger.Log("err", err, "body", r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}

	defer r.Body.Close()
	// инициализируем переменную для получения информации по поставщику
	var news models.News
	// анмаршалим в структуру
	err = json.Unmarshal(newsInfo, &news)
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	resp, err := app.svc.CreateOneNews(news)
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		app.logger.Log("err", err)
	}
}

func (app *Application) createManyNews(w http.ResponseWriter, r *http.Request) {
	newsInfo, err := ioutil.ReadAll(r.Body)
	if err != nil {
		app.logger.Log("err", err, "body", r.Body)
		w.WriteHeader(http.StatusInternalServerError)
	}

	defer r.Body.Close()
	// инициализируем переменную для получения информации по поставщику
	var news []models.News
	// анмаршалим в структуру
	err = json.Unmarshal(newsInfo, &news)
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
	resp, err := app.svc.CreateManyNews(news)
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		app.logger.Log("err", err)
	}
}

func (app *Application) getNewsAll(w http.ResponseWriter, r *http.Request) {
	resp, err := app.svc.GetNewsAll()
	if err != nil {
		app.logger.Log("err", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, err = w.Write(resp)
	if err != nil {
		app.logger.Log("err", err)
	}
}
