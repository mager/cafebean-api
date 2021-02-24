package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

type RoasterMap struct {
	Name string `firestore:"name" json:"name"`
	Slug string `firestore:"slug" json:"slug"`
}

// Bean represents a coffee bean
type Bean struct {
	Countries   []string   `firestore:"countries" json:"countries"`
	Description string     `firestore:"description" json:"description"`
	Flavors     []string   `firestore:"flavors" json:"flavors"`
	Name        string     `firestore:"name" json:"name"`
	Roaster     RoasterMap `firestore:"roaster" json:"roaster"`
	Shade       string     `firestore:"shade" json:"shade"`
	Slug        string     `firestore:"slug" json:"slug"`
	URL         string     `firestore:"url" json:"url"`
	Year        int64      `firestore:"year" json:"year"`
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

func docToBean(doc *firestore.DocumentSnapshot) Bean {
	var b Bean
	doc.DataTo(&b)
	b.Slug = doc.Ref.ID
	return b
}

func (h *Handler) getBean(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &BeanResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
		ctx  = context.TODO()
	)

	// Get the bean
	doc, err := h.database.Collection("beans").Doc(slug).Get(ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(
			&ErrorMessage{
				Message: fmt.Sprintf("Failed to get document: %s", slug),
			},
		)
	} else {
		resp.Bean = docToBean(doc)

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
			h.logger.Fatalf("Failed to iterate: %v", err)
		}

		resp.Beans = append(resp.Beans, docToBean(doc))
	}

	json.NewEncoder(w).Encode(resp)
}

// EditBeanReq is the request body for adding a Bean
// NOTE: Currently you can only update a bean name
type EditBeanReq struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Flavors     []string `json:"flavors"`
}

// EditBeanResp is the response from the POST /beans endpoint
type EditBeanResp struct {
	Bean
}

func (h *Handler) editBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		vars      = mux.Vars(r)
		slug      = vars["slug"]
		err       error
		req       EditBeanReq
		resp      = &EditBeanResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the bean
	bean := h.database.Collection("beans").Doc(slug)
	docsnap, err := bean.Get(ctx)
	if err != nil {
		http.Error(w, "invalid bean slug", http.StatusBadRequest)
		return
	}

	// Update the bean
	result, err := bean.Update(
		ctx,
		[]firestore.Update{
			{Path: "flavors", Value: req.Flavors},
			{Path: "description", Value: req.Description},
			{Path: "name", Value: req.Name},
			{Path: "url", Value: req.URL},
		},
	)
	h.logger.Infow(
		"Bean updated",
		"id", docsnap.Ref.ID,
		"updated_at", result.UpdateTime,
		"updated_by", userEmail,
	)

	// Send event
	// TODO: Send updated fields
	t := h.events.Topic("bean")
	res := t.Publish(ctx, &pubsub.Message{
		Data: []byte("Bean updated"),
		Attributes: map[string]string{
			"id":         docsnap.Ref.ID,
			"user_email": userEmail,
		},
	})
	msgID, err := res.Get(ctx)
	if err != nil {
		log.Fatal(err)
	}
	h.logger.Infow("Pubsub message succeeded", "msgId", msgID)

	// Send updated bean response
	w.WriteHeader(http.StatusAccepted)

	updated, err := bean.Get(ctx)
	if err != nil {
		h.logger.Errorw(
			"Error fetching bean after updating it",
			"id", updated.Ref.ID,
		)
	}
	resp.Bean = docToBean(updated)

	json.NewEncoder(w).Encode(resp)
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

	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(resp)
}
