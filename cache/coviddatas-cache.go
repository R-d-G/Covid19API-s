package cache

import "example.com/go-demo/models"

type CoviddataCache interface {
	Set(key string, value *models.Coviddata)
	Get(key string) *models.Coviddata
}
