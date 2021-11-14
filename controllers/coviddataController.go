// Package Fetching of Coviddata API
//
// Documentation for Coviddata API
//
//	Schemes: http
//	BasePath: /
//	Version: 1.0.0
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
// swagger:meta
package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/go-demo/cache"
	"example.com/go-demo/config"
	"example.com/go-demo/models"

	"github.com/gofiber/fiber/v2"
	"github.com/tidwall/gjson"
	"go.mongodb.org/mongo-driver/bson"
)

var coviddataCache cache.CoviddataCache = cache.NewRedisCache("redis-12023.c290.ap-northeast-1-2.ec2.cloud.redislabs.com:12023", 0, 30)

// swagger:route GET /api/Coviddatas Coviddatas fetchAndStoreCoviddata
// Return a success response
// responses:
//	200: successMeassage
//
// FetchAndStoreCoviddata handles GET requests and returns success meassage
func FetchAndStoreCoviddata(c *fiber.Ctx) error {

	coviddataCollection := config.MI.DB.Collection("coviddatas")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	//fetching data from publiclly available api
	resp, err := http.Get("https://data.covid19india.org/v4/min/data.min.json")
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	//Convert the body to type string
	// log.Print(body)
	sb := string(body)
	stateCodes := [...]string{"AN", "AP", "AR", "AS", "BR", "CH", "CT", "DN",
		"DL", "GA", "GJ", "HR", "HP", "JK", "JH", "KA", "KL", "LA", "LD", "MP",
		"MH", "MN", "ML", "MZ", "NL", "OR", "PY", "PB", "RJ", "SK", "TN",
		"TG", "TR", "UP", "UT", "WB"}
	var i int
	for i = 0; i < len(stateCodes); i++ {
		sc := stateCodes[i]
		//fetching required fields from json response from website
		stringConfirmed := sc + ".total.confirmed"
		stringLastUpdated := sc + ".meta.last_updated"
		totalConfirmed := gjson.Get(sb, stringConfirmed)
		lastUpdated := gjson.Get(sb, stringLastUpdated)

		coviddata := models.Coviddata{
			StateCode:      stateCodes[i],
			ConfirmedCases: totalConfirmed.String(),
			LastUpdated:    lastUpdated.String(),
		}

		update := bson.M{
			"$set": coviddata,
		}
		log.Println("updating record for statecode = " + stateCodes[i] + " if it exhists ")
		result, err := coviddataCollection.UpdateOne(ctx, bson.M{"stateCode": stateCodes[i]}, update)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Coviddata failed to update",
				"error":   err.Error(),
			})
		}

		if result.MatchedCount == int64(0) {
			log.Println("Records do not exhist!! Making new record for statecode =" + stateCodes[i])
			_, err := coviddataCollection.InsertOne(ctx, coviddata)
			if err != nil {
				return c.Status(500).JSON(fiber.Map{
					"success": false,
					"message": "Coviddata failed to insert",
					"error":   err,
				})
			}
		}

	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Coviddatafetched and updated to mongoDB Atlas successfully",
	})
}

// swagger:route GET /api/GPSbasedcoviddata GPSbasedcoviddata gPSbasedcoviddata
// Return a coviddata object
// responses:
//		200: coviddataResponse
//		404: errorResponse
//
// FetchAndStoreCoviddata handles GET requests and returns coviddata object
func FetchCoviddataAtLocation(c *fiber.Ctx) error {

	coviddataCollection := config.MI.DB.Collection("coviddatas")
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()
	gpsCoodinates := new(models.GPScoodinates)

	if err := c.BodyParser(gpsCoodinates); err != nil {
		log.Println(err)
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"message": "Failed to parse body",
			"error":   err,
		})
	}

	log.Println("fetched GPS Coordinates : " + gpsCoodinates.Latitude + " , " + gpsCoodinates.Longitude)

	URl := "http://api.positionstack.com/v1/reverse?access_key=" + os.Getenv("YOUR_GEOCODINGAPI_ACCESS_KEY") + "&query=" + gpsCoodinates.Latitude + "," + gpsCoodinates.Longitude + "&limit=1"
	log.Println(URl)

	//fetching data from publiclly available api
	resp, err := http.Get(URl)
	if err != nil {
		log.Fatalln(err)
	}
	//We Read the response body on the line below.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	sb := string(body)
	log.Println("Reverse Geocoding Data from API :\n" + sb)

	//fetching required fields from api json response
	stateCode := gjson.Get(sb, "data.0.region_code").String()

	var coviddata models.Coviddata

	//getting data from cache
	var coviddataFromCache *models.Coviddata = coviddataCache.Get(stateCode)

	//data not fount in cache
	if coviddataFromCache == nil {
		var coviddatas []models.Coviddata
		cursor, err := coviddataCollection.Find(ctx, bson.M{"stateCode": stateCode})
		if err != nil {
			panic(err)
		}
		if err = cursor.All(ctx, &coviddatas); err != nil {
			panic(err)
		}
		if len(coviddatas) == 0 {
			panic("data not returned by api")
		}
		fmt.Println(coviddatas[0])
		coviddata = coviddatas[0]

		log.Println("saving in cache")
		coviddataPointer := &coviddata
		coviddataCache.Set(stateCode, coviddataPointer)
		log.Println("getting from cache")
		var coviddataFromCacheTemp *models.Coviddata = coviddataCache.Get(stateCode)
		if coviddataFromCacheTemp == nil {
			panic("problem occured")
		}
		tempcovid := *coviddataFromCacheTemp
		log.Println(" -- " + tempcovid.ConfirmedCases)

	} else {
		log.Println("data found in cache for statecode = " + stateCode)
		coviddata = *coviddataFromCache
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"Your state code": coviddata.StateCode,
		"total cases":     coviddata.ConfirmedCases,
		"last updated":    coviddata.LastUpdated,
	})
}
