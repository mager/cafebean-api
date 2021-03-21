package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"

	"google.golang.org/genproto/googleapis/type/latlng"
)

// Roaster represents an organization that roasts beans
type Roaster struct {
	City      string         `firestore:"city" json:"city"`
	Instagram string         `firestore:"instagram" json:"instagram"`
	Location  *latlng.LatLng `firestore:"location" json:"location"`
	Logo      string         `firestore:"logo" json:"logo"`
	Name      string         `firestore:"name" json:"name"`
	Slug      string         `firestore:"slug" json:"slug"`
	Twitter   string         `firestore:"twitter" json:"twitter"`
	URL       string         `firestore:"url" json:"url"`
}

// RoasterDB represents a Roaster in firestore
type RoasterDB struct {
	Roaster
	Verified bool `firestore:"verified"`
}

// RoasterBQ represents a coffee roaster
type RoasterBQ struct {
	City      string
	Instagram string
	Location  string
	Logo      string
	Name      string
	Slug      string
	Twitter   string
	URL       string
}

type RoasterBQItem struct {
	Roaster   RoasterBQ
	UpdatedBy string
	UpdatedAt string
}

// RoasterReq is the request body for adding and updating a Roaster
type RoasterReq struct {
	Roaster
}

// RoasterResp is the response for the GET /roaster/{slug} endpoint
type RoasterResp struct {
	Roaster Roaster `json:"roaster"`
	Beans   []Bean  `json:"beans"`
}

// RoastersResp is the response for the GET /roasters endpoint
type RoastersResp struct {
	Roasters []Roaster `json:"roasters"`
}

// RoastersListResp returns a list of unique roasters
type RoastersListResp struct {
	Roasters []RoasterMap `json:"roasters"`
}

func docToRoaster(doc *firestore.DocumentSnapshot) Roaster {
	var r Roaster
	doc.DataTo(&r)
	return r
}

// recordRoasterChange posts a changelog event to BigQuery
func (h *Handler) recordRoasterChange(ctx context.Context, req RoasterReq, userEmail string) {
	dataset := h.bq.DatasetInProject("cafebean", "roaster")
	table := dataset.Table("changelog")

	u := table.Inserter()
	items := []*RoasterBQItem{
		{
			Roaster: RoasterBQ{
				City:      req.City,
				Instagram: req.Instagram,
				Location:  fmt.Sprintf("POINT(%f %f)", req.Location.Longitude, req.Location.Latitude),
				Logo:      req.Logo,
				Name:      req.Name,
				Slug:      req.Slug,
				URL:       req.URL,
				Twitter:   req.Twitter,
			},
			UpdatedBy: userEmail,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}
	if err := u.Put(ctx, items); err != nil {
		h.logger.Error(err)
	}
}

func (h *Handler) getRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoasterResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
		ctx  = context.TODO()
	)

	// Get the roaster
	roasterIter := h.database.Collection("roasters").Where("slug", "==", slug).Documents(ctx)
	for {
		doc, err := roasterIter.Next()
		if doc == nil {
			http.Error(w, "invalid roaster", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Roaster = docToRoaster(doc)

		break
	}

	// Get the beans for that roaster
	beansIter := h.database.Collection("beans").Where("roaster.slug", "==", resp.Roaster.Slug).Documents(ctx)
	for {
		doc, err := beansIter.Next()
		if doc == nil {
			break
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Beans = append(resp.Beans, docToBean(doc))
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getRoasters(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoastersResp{}
	)

	// Call Firestore API
	iter := h.database.Collection("roasters").Documents(context.TODO())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		var r Roaster
		doc.DataTo(&r)

		resp.Roasters = append(resp.Roasters, r)
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) getRoastersList(w http.ResponseWriter, r *http.Request) {
	var (
		resp = &RoastersListResp{}
	)

	// Call Firestore API
	iter := h.database.Collection("roasters").Documents(context.TODO())
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Failed to iterate: %v", err)
		}

		var r RoasterMap
		doc.DataTo(&r)

		resp.Roasters = append(resp.Roasters, r)
	}

	json.NewEncoder(w).Encode(resp)
}

// AddRoasterResp is the response from the POST /roasters/{slug} endpoint
type AddRoasterResp struct {
	ID string `json:"id"`
}

func (h *Handler) addRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       RoasterReq
		resp      = &AddRoasterResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure the roaster doesn't already exist
	roasterIter := h.database.Collection("roasters").Where("slug", "==", req.Slug).Documents(ctx)
	for {
		doc, err := roasterIter.Next()

		if doc != nil {
			http.Error(w, "roaster already exists", http.StatusBadRequest)
			return
		}
		if err != nil && err.Error() != "no more items in iterator" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		break
	}

	// Add the roaster
	doc, _, err := h.database.Collection("roasters").Add(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Infow(
		"Roaster added",
		"id", doc.ID,
		"updated_by", userEmail,
	)

	// Publish an entry in BigQuery
	h.recordRoasterChange(ctx, req, userEmail)

	// Send updated roaster response
	w.WriteHeader(http.StatusAccepted)

	resp.ID = doc.ID

	json.NewEncoder(w).Encode(resp)
}

// EditRoasterResp is the response from the POST /roasters/{slug} endpoint
type EditRoasterResp struct {
	Roaster
}

func (h *Handler) editRoaster(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		docID     string
		vars      = mux.Vars(r)
		slug      = vars["slug"]
		err       error
		req       RoasterReq
		resp      = &EditRoasterResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the roaster
	q := h.database.Collection("roasters").Where("slug", "==", slug)
	roasterIter := q.Documents(ctx)
	defer roasterIter.Stop()
	for {
		doc, err := roasterIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		docID = doc.Ref.ID
	}

	roaster := h.database.Collection("roasters").Doc(docID)
	docsnap, err := roaster.Get(ctx)
	if err != nil {
		http.Error(w, "invalid roaster slug", http.StatusBadRequest)
		return
	}

	// Update the roaster
	result, err := roaster.Update(
		ctx,
		[]firestore.Update{
			{Path: "city", Value: req.City},
			{Path: "instagram", Value: req.Instagram},
			{Path: "location", Value: req.Location},
			{Path: "logo", Value: req.Logo},
			{Path: "name", Value: req.Name},
			{Path: "slug", Value: req.Slug},
			{Path: "twitter", Value: req.Twitter},
			{Path: "url", Value: req.URL},
		},
	)
	h.logger.Infow(
		"Roaster updated",
		"id", docsnap.Ref.ID,
		"updated_at", result.UpdateTime,
		"updated_by", userEmail,
	)

	// Publish an entry in BigQuery
	h.recordRoasterChange(ctx, req, userEmail)

	// Send updated roaster response
	w.WriteHeader(http.StatusAccepted)

	updated, err := roaster.Get(ctx)
	if err != nil {
		h.logger.Errorw(
			"Error fetching roaster after updating it",
			"id", updated.Ref.ID,
		)
	}
	h.logger.Debug(updated)
	resp.Roaster = docToRoaster(updated)

	json.NewEncoder(w).Encode(resp)
}
