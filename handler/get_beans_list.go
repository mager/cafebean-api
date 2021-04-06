package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"google.golang.org/api/iterator"
)

func (h *Handler) getBeansList(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &BeansListResp{}
	)

	// Call Firestore API
	iter := h.database.Collection("beans").Documents(context.TODO())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		name, err := doc.DataAt("name")
		if err != nil {
			h.logger.Error(err.Error())
		}

		roaster, err := doc.DataAt("roaster.name")
		if err != nil {
			h.logger.Error(err.Error())
		}

		slug, err := doc.DataAt("slug")
		if err != nil {
			h.logger.Error(err.Error())
		}

		bean := BeanSimple{
			Name:    name.(string),
			Roaster: roaster.(string),
			Slug:    slug.(string),
		}

		resp.Beans = append(resp.Beans, bean)
	}

	json.NewEncoder(w).Encode(resp)
}
