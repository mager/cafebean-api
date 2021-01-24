package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

// Bean represents a coffee bean
type Bean struct {
	Countries   []string `firestore:"countries" json:"countries"`
	Description string   `firestore:"description" json:"description"`
	Flavors     []string `firestore:"flavors" json:"flavors"`
	Name        string   `firestore:"name" json:"name"`
	Roaster     string   `firestore:"roaster" json:"roaster"`
	Shade       string   `firestore:"shade" json:"shade"`
	Slug        string   `firestore:"slug" json:"slug"`
	URL         string   `firestore:"url" json:"url"`
	Year        int64    `firestore:"year" json:"year"`
}

// BeanDB represents a Bean in firestore
type BeanDB struct {
	Bean
}

// BeanResp is the response for the GET /bean/{slug} endpoint
type BeanResp struct {
	Bean Bean `json:"bean"`
}

// BeansResp is the response for the GET /beans endpoint
type BeansResp struct {
	Beans []Bean `json:"beans"`
}

// AddBeanReq is the request body for adding a Bean
type AddBeanReq struct {
	Flavors     []string `json:"flavors"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Roaster     string   `json:"roaster"`
	Shade       string   `json:"shade"`
}

// AddBeanResp is the response from the POST /beans endpoint
type AddBeanResp struct {
	ID string `json:"id"`
}

func (h *Handler) getBean(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &BeanResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
		ctx  = context.TODO()
	)
	// Call Firestore API
	doc, err := h.database.Collection("beans").Doc(slug).Get(ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			&ErrorMessage{
				Message: fmt.Sprintf("Failed to get document: %s", slug),
			},
		)
	} else {
		var b Bean
		doc.DataTo(&b)
		resp.Bean = b

		json.NewEncoder(w).Encode(resp)
	}
}

func (h *Handler) getBeans(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &BeansResp{}
		ctx  = context.TODO()
	)

	// Call Firestore API
	iter := h.database.Collection("beans").Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		var b Bean
		doc.DataTo(&b)
		b.Slug = doc.Ref.ID
		resp.Beans = append(resp.Beans, b)
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) addBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = context.TODO()
		err  error
		req  AddBeanReq
		resp = &AddBeanResp{}
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure roaster exists
	iter := h.database.Collection("roasters").Where("name", "==", req.Roaster).Documents(ctx)
	for {
		doc, err := iter.Next()
		if doc == nil {
			http.Error(w, "invalid roaster", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}

	// Add the bean
	doc, _, err := h.database.Collection("beans").Add(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.ID = doc.ID

	json.NewEncoder(w).Encode(resp)
}
