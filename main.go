package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"telmax/templatemail"
	"time"

	"github.com/gorilla/mux"
	"github.com/mohamedattahri/mail"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func init() {
	flag.Parse()
	lvl, _ := log.ParseLevel(*LogLevel)
	log.SetLevel(lvl)
	// Connect to the database
	DBClient = DBConnect(*MongoURI)
	if DBClient != nil {
		CoreDB = DBClient.Database(*CoreDatabase)
	}
}

func main() {
	r := mux.NewRouter()
	r.Methods("OPTIONS").HandlerFunc(HandleOptions)

	r.HandleFunc("/", HandleHome).Methods("GET")
	r.HandleFunc("/newtemplate/", HandleSaveTemplate).Methods("POST")
	r.HandleFunc("/updatetemplate/{id}", HandleUpdateTemplate).Methods("POST")
	r.HandleFunc("/gettemplatebyfilter/", HandleGetTemplatesByFilter).Methods("GET")
	r.HandleFunc("/sendemail/{sender},{template_id}", HandleSendEmail).Methods("POST")

	port := os.Getenv("PORT")
	if port == "" {
		port = "5004"
	}

	log.Warning("Listing on PORT ", port)
	log.Fatal(http.ListenAndServe(":"+port, r))

}

func HandleHome(w http.ResponseWriter, r *http.Request) {
	CORSHeaders(w, r)
	fmt.Fprintf(w, "Homepage")
}

func HandleOptions(w http.ResponseWriter, r *http.Request) {
	CORSHeaders(w, r)
}

func CORSHeaders(w http.ResponseWriter, r *http.Request) {
	headers := w.Header()
	log.Infof("Request Origin is: %v", r.Header.Get("Origin"))
	headers.Add("Access-Control-Allow-Origin", "*")
	headers.Add("Access-Control-Allow-Headers", "Content-Type, Origin, Accept, token, api-key")
	headers.Add("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	headers.Add("Access-Control-Max-Age", "3600")
}

func HandleSaveTemplate(w http.ResponseWriter, r *http.Request) {
	CORSHeaders(w, r)

	reqBody, err := ioutil.ReadAll(r.Body)
	var response Response
	var template templatemail.Template
	var succcess bool

	err = json.Unmarshal(reqBody, &template)
	if err != nil {
		log.Error(err)
	}

	log.Errorf("Template Body is %v", string(reqBody))
	succcess, err = template.Insert(CRMDB)
	if err != nil {
		response.Status = "error"
		response.Error = "Problem creating template " + err.Error()
	}
	if succcess {
		response.Data = template
		response.Status = "ok"
	}

	json.NewEncoder(w).Encode(response)

}

func HandleUpdateTemplate(w http.ResponseWriter, r *http.Request) {
	CORSHeaders(w, r)

	id := mux.Vars(r)["id"]
	var response Response
	var success bool
	log.Infof("Updating Template ID %v ", id)
	template, err := GetSingleTemplate(CRMDB, id)
	if err != nil {
		errstr := "Problem getting template " + id + " " + err.Error()
		log.Error(errstr)
		response.Status = "error"
		response.Error = errstr
	} else {
		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Errorf("Please provide valid update data  %s ", r)
			response.Error = "Invalid update body"
		}
		log.Debug("request body for update is %v ", string(reqBody))
		json.Unmarshal(reqBody, &template)
		log.Infof("New Template object is %v ", template)
		success, err = template.Update(CRMDB)
		if success == true {
			response.Status = "ok"
			response.Data = template
		}
		if err != nil {
			errstr := "Problem getting template " + id + " " + err.Error()
			log.Error(errstr)
			response.Status = "error"
			response.Error = errstr
		} else {

		}
	}

	json.NewEncoder(w).Encode(response)
}

func GetSingleTemplate(database *mongo.Database, id string) (template *templatemail.Template, err error) {
	log.Infof("Fetching template %v ", id)
	filter := bson.D{{"template_id", id}}

	result := database.Collection("emailTemplates").FindOne(context.TODO(), filter)
	if err != nil {
		log.Errorf("Problem getting template %v - %v ", id, err)
	}
	err = result.Decode(&template)
	if err != nil {
		log.Errorf("error decoding BSON %v from template %v ", err, id)
	}

	return
}

func GetEmails(database *mongo.Database, filters []Filter) (templates []templatemail.Template, err error) {
	log.Info("Fetching Emails using Filter %v", filters)

	options := options.Find()
	options.SetSort(bson.D{{"created_date", -1}})

	filter := bson.D{{}}
	layout := "2006-01-02T15:04:05Z"

	if len(filters) > 0 {
		for _, filterEntry := range filters {
			log.Debugf("Filtering on %v = %v", filterEntry.Key, filterEntry.Value)
			oper := "$in"
			if filterEntry.Value[:1] == "$" {
				var value interface{}
				log.Info("Using special filter operator")
				if filterEntry.Value[:4] == "$lte" {
					oper = "$lte"
					value = filterEntry.Value[4:]
				} else if filterEntry.Value[:4] == "$gte" {
					oper = "$gte"
					value = filterEntry.Value[4:]
				} else if filterEntry.Value[:3] == "$lt" {
					oper = "$lte"
					value = filterEntry.Value[3:]
				} else if filterEntry.Value[:3] == "$gt" {
					oper = "$gte"
					value = filterEntry.Value[3:]
				}
				t, err := time.Parse(layout, value.(string))
				if err != nil {
					log.Error(err)
				} else {
					value = t
				}
				filter = append(filter, bson.E{filterEntry.Key, bson.D{{oper, value}}})
			} else {
				filter = append(filter, bson.E{filterEntry.Key, bson.D{{oper, bson.A{filterEntry.Value}}}})
			}
		}
	}
	log.Info(filter)

	cur, err := database.Collection("emailTemplates").Find(context.TODO(), filter, options)
	if err != nil {
		log.Errorf("Problem Finding templates where filter is %v - %v ", filter, err)
	}

	for cur.Next(context.TODO()) {
		var template templatemail.Template
		err = cur.Decode(&template)
		if err != nil {
			log.Errorf("Error decoding BSON %v", err)
		} else {
			templates = append(templates, template)
		}
	}

	cur.Close(context.TODO())
	return

}

func HandleGetTemplatesByFilter(w http.ResponseWriter, r *http.Request) {
	CORSHeaders(w, r)
	var response Response

	var filters []Filter
	requestVars := r.URL.Query()
	for filterkey, filterval := range requestVars {
		filters = append(filters, Filter{Key: filterkey, Value: filterval[0]})
	}

	templates, err := GetEmails(CRMDB, filters)
	if err != nil {
		errstr := "Problem fetching templates " + err.Error()
		log.Errorf("%v filter - %v", errstr, filters)
		response.Status = "error"
		response.Error = errstr
	} else {
		response.Status = "ok"
		response.Data = templates
	}

	json.NewEncoder(w).Encode(response)
}

func HandleSendEmail(w http.ResponseWriter, r *http.Request) {
	CORSHeaders(w, r)
	var response Response
	var success bool
	Sender := mux.Vars(r)["sender"]
	TemplateID := mux.Vars(r)["template_id"]

	//log.Infof("Lead object for sending Email %v", lead)
	success, _ = sendEmail(Sender, TemplateID)
	if success == true {
		response.Status = "ok"
	}

	json.NewEncoder(w).Encode(response)
}

func sendEmail(sender string, id string) (success bool, err error) {
	template := templatemail.GetTemplate(CRMDB.Collection("emailTemplates"), id, "crm")
	recipients := []mail.Address{
		mail.Address{
			Name:    "Customer",
			Address: sender,
		},
	}

	Sender := mail.Address{
		Name:    "telmax CRM",
		Address: sender,
	}

	//log.Infof("Using HTML template %v - code is - \n%v", template.Description, template.Text)
	message := NewTemplateMail(SMTP, Sender, template.Subject, recipients)
	err = message.AddHtml(template.Text, "Hello")
	if err != nil {
		log.Errorf("Problem generating HTML part %v", err)
	} else {
		message.Send()
		success = true
	}

	return
}
