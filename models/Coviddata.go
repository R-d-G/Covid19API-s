package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Coviddata struct {
	ID             primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StateCode      string             `json:"stateCode,omitempty" bson:"stateCode,omitempty"`
	ConfirmedCases string             `json:"confirmedCases,omitempty" bson:"confirmedCases,omitempty"`
	LastUpdated    string             `json:"lastUpdated,omitempty" bson:"lastUpdated,omitempty"`
}
