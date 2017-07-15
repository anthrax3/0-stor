package rest

//This file is auto-generated by go-raml
//Do not edit this file by hand since it will be overwritten during the next generation

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/zero-os/0-stor/store/config"
	"github.com/zero-os/0-stor/store/db"
	"github.com/zero-os/0-stor/store/rest/models"
)

// NamespacesInterface is interface for /namespaces root endpoint
type NamespacesInterface interface { // nsidaclPost is the handler for POST /namespaces/{nsid}/acl
	// Create an dataAccessToken for a user. This token gives this user access to the data in this namespace
	nsidaclPost(http.ResponseWriter, *http.Request)
	// DeleteObject is the handler for DELETE /namespaces/{nsid}/objects/{id}
	// Delete object from the KV
	DeleteObject(http.ResponseWriter, *http.Request)
	// HeadObject is the handler for HEAD /namespaces/{nsid}/objects/{id}
	// Test object exists in the KV
	HeadObject(http.ResponseWriter, *http.Request)
	// GetObject is the handler for GET /namespaces/{nsid}/objects/{id}
	// Retrieve object from the KV
	GetObject(http.ResponseWriter, *http.Request)
	// Listobjects is the handler for GET /namespaces/{nsid}/objects
	// List keys of the namespaces
	Listobjects(http.ResponseWriter, *http.Request)
	// Createobject is the handler for POST /namespaces/{nsid}/objects
	// Set an object into the namespace
	Createobject(http.ResponseWriter, *http.Request)
	// nsidreservationidGet is the handler for GET /namespaces/{nsid}/reservation/{id}
	// Return information about a reservation
	nsidreservationidGet(http.ResponseWriter, *http.Request)
	// UpdateReservation is the handler for PUT /namespaces/{nsid}/reservation/{id}
	// Renew an existing reservation
	UpdateReservation(http.ResponseWriter, *http.Request)
	// ListReservations is the handler for GET /namespaces/{nsid}/reservation
	// Return a list of all the existing reservation for the give resource
	ListReservations(http.ResponseWriter, *http.Request)
	// CreateReservation is the handler for POST /namespaces/{nsid}/reservation
	// Create a reservation for the given resource.
	CreateReservation(http.ResponseWriter, *http.Request)
	// StatsNamespace is the handler for GET /namespaces/{nsid}/stats
	// Return usage statistics about this namespace
	StatsNamespace(http.ResponseWriter, *http.Request)
	// Deletensid is the handler for DELETE /namespaces/{nsid}
	// Delete nsid
	Deletensid(http.ResponseWriter, *http.Request)
	// Getnsid is the handler for GET /namespaces/{nsid}
	// Get detail view about nsid
	Getnsid(http.ResponseWriter, *http.Request)
	// Listnamespaces is the handler for GET /namespaces
	// List all namespaces
	Listnamespaces(http.ResponseWriter, *http.Request)
	// Createnamespace is the handler for POST /namespaces
	// Create a new namespace
	Createnamespace(http.ResponseWriter, *http.Request)
	// GetStoreStats is the handler for GET /namespaces/stats
	// Return usage statistics about the whole KV
	GetStoreStats(http.ResponseWriter, *http.Request)
	// UpdateStoreStats is the handler for POST /namespaces/stats
	// Update Global Store statistics and available space
	UpdateStoreStats(http.ResponseWriter, *http.Request)

	// Get db object
	DB() db.DB

	// Get Settings object
	Config() config.Settings
}

// NamespacesInterfaceRoutes is routing for /namespaces root endpoint
func NamespacesInterfaceRoutes(r *mux.Router, i NamespacesInterface) {

	r.Handle("/namespaces/{nsid}/acl",
		//alice.New(NewReservationValidMiddleware(i.DB()).Handler).
		alice.New().
			Then(http.HandlerFunc(i.nsidaclPost))).Methods("POST")

	r.Handle("/namespaces/{nsid}/objects/{id}",
		alice.New(
			NewDataTokenValidMiddleware(models.ACLEntry{ // At least user should have Delete permissions
				Read:   false,
				Write:  false,
				Delete: true,
				Admin:  false,
			}).Handler,
			NewReservationValidMiddleware(i.DB()).Handler).
			Then(http.HandlerFunc(i.DeleteObject))).Methods("DELETE")

	r.Handle("/namespaces/{nsid}/objects/{id}",
		alice.New(
			NewDataTokenValidMiddleware(models.ACLEntry{ // At least user should have  Read permissions
				Read:   true,
				Write:  false,
				Delete: false,
				Admin:  false,
			}).Handler,
			NewReservationValidMiddleware(i.DB()).Handler).
			Then(http.HandlerFunc(i.HeadObject))).Methods("HEAD")

	r.Handle("/namespaces/{nsid}/objects/{id}",
		alice.New(
			NewDataTokenValidMiddleware(models.ACLEntry{ // At least user should have  Read permissions
				Read:   true,
				Write:  false,
				Delete: false,
				Admin:  false,
			}).Handler,
			NewReservationValidMiddleware(i.DB()).Handler).
			Then(http.HandlerFunc(i.GetObject))).Methods("GET")

	r.Handle("/namespaces/{nsid}/objects",
		alice.New(
			NewDataTokenValidMiddleware(models.ACLEntry{ // At least user should have  Read permissions
				Read:   true,
				Write:  false,
				Delete: false,
				Admin:  false,
			}).Handler,
			NewReservationValidMiddleware(i.DB()).Handler).
			Then(http.HandlerFunc(i.Listobjects))).Methods("GET")

	r.Handle("/namespaces/{nsid}/objects",
		alice.New(
			NewDataTokenValidMiddleware(models.ACLEntry{ // At least user should have  write permissions
				Read:   false,
				Write:  true,
				Delete: false,
				Admin:  false,
			}).Handler,
			NewReservationValidMiddleware(i.DB()).Handler).
			Then(http.HandlerFunc(i.Createobject))).Methods("POST")

	r.Handle("/namespaces/{nsid}/reservation/{id}",
		alice.New(
			NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.nsidreservationidGet))).Methods("GET")

	r.Handle("/namespaces/{nsid}/reservation/{id}",
		alice.New(
			NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.UpdateReservation))).Methods("PUT")

	r.Handle("/namespaces/{nsid}/reservation",
		alice.New(
			NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.ListReservations))).Methods("GET")

	r.Handle("/namespaces/{nsid}/reservation",
		alice.New(
			NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.CreateReservation))).Methods("POST")

	r.Handle("/namespaces/{nsid}/stats",
		alice.New(
			NewDataTokenValidMiddleware(models.ACLEntry{ // Admin permissions
				Read:   true,
				Write:  true,
				Delete: true,
				Admin:  true,
			}).Handler).
			Then(http.HandlerFunc(i.StatsNamespace))).Methods("GET")

	/* namespaces/stats Must come before /namespaces/{nsid}
	   Otherwise it won't match!
	*/

	r.Handle("/namespaces/stats",
		alice.New(NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.UpdateStoreStats))).Methods("POST")

	r.Handle("/namespaces/stats",
		alice.New(NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.GetStoreStats))).Methods("GET")

	r.Handle("/namespaces/{nsid}",
		http.HandlerFunc(i.Deletensid)).Methods("DELETE")

	r.Handle("/namespaces/{nsid}",
		http.HandlerFunc(i.Getnsid)).Methods("GET")

	r.Handle("/namespaces",
		alice.New(NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.Listnamespaces))).Methods("GET")

	r.Handle("/namespaces",
		alice.New(NewOauth2itsyouonlineMiddleware([]string{"user:name"}).Handler).
			Then(http.HandlerFunc(i.Createnamespace))).Methods("POST")
}
