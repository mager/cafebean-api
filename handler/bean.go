package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
	"github.com/gorilla/mux"
	"google.golang.org/api/iterator"
)

// RoasterMap represents the roaster
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
	Photo       string     `firestore:"photo" json:"photo"`
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

// BeanBQ represents a coffee bean
type BeanBQ struct {
	Countries   string
	Description string
	Flavors     string
	Name        string
	Photo       string
	Roaster     RoasterMap
	Shade       string
	Slug        string
	URL         string
	Year        int64
}

type BeanBQItem struct {
	Bean      BeanBQ
	UpdatedBy string
	UpdatedAt string
}

// BeanReq is the request body for adding or updating a Bean
type BeanReq struct {
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
	if b.Countries == nil {
		b.Countries = []string{}
	}
	return b
}

// recordBeanChange posts a changelog event to BigQuery
func (h *Handler) recordBeanChange(ctx context.Context, req BeanReq, userEmail string) {
	dataset := h.bq.DatasetInProject("cafebean", "bean")
	table := dataset.Table("changelog")

	u := table.Inserter()
	items := []*BeanBQItem{
		{
			Bean: BeanBQ{
				Countries:   strings.Join(req.Countries, ", "),
				Description: req.Description,
				Flavors:     strings.Join(req.Flavors, ", "),
				Name:        req.Name,
				Photo:       req.Photo,
				Roaster:     req.Roaster,
				Shade:       req.Shade,
				Slug:        req.Slug,
				Year:        req.Year,
				URL:         req.URL,
			},
			UpdatedBy: userEmail,
			UpdatedAt: time.Now().Format(time.RFC3339),
		},
	}
	if err := u.Put(ctx, items); err != nil {
		h.logger.Error(err)
	}
}

func (h *Handler) postBeanToDiscord(req BeanReq, userEmail string, action string) (*discordgo.Message, error) {
	content := "A bean was added!"
	if action == "edit" {
		content = "A bean was updated!"
	}
	return h.discord.WebhookExecute(
		h.cfg.DiscordBeansWebhookID,
		h.cfg.DiscordBeansWebhookToken,
		false,
		&discordgo.WebhookParams{
			Content: content,
			Embeds: []*discordgo.MessageEmbed{{
				Author: &discordgo.MessageEmbedAuthor{
					Name: userEmail,
				},
				Title:       fmt.Sprintf("%s - %s", req.Roaster.Name, req.Name),
				Description: req.Description,
				URL:         fmt.Sprintf("https://cafebean.org/beans/%s", req.Slug),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Flavors",
						Value: strings.Join(req.Flavors, ", "),
					},
					{
						Name:  "Countries",
						Value: strings.Join(req.Countries, ", "),
					},
				},
				Provider: &discordgo.MessageEmbedProvider{
					URL:  req.URL,
					Name: req.Roaster.Name,
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL:   req.Photo,
					Width: 32,
				},
			}},
		},
	)
}

func (h *Handler) getBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx  = context.TODO()
		resp = &BeanResp{}
		vars = mux.Vars(r)
		slug = vars["slug"]
	)

	// Get the bean
	iter := h.database.Collection("beans").Where("slug", "==", slug).Documents(ctx)
	for {
		doc, err := iter.Next()
		if doc == nil {
			http.Error(w, "invalid bean", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp.Bean = docToBean(doc)

		break
	}

	json.NewEncoder(w).Encode(resp)
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

// EditBeanResp is the response from the POST /beans/{slug} endpoint
type EditBeanResp struct {
	Bean
}

func (h *Handler) editBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		docID     string
		vars      = mux.Vars(r)
		slug      = vars["slug"]
		err       error
		req       BeanReq
		resp      = &EditBeanResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch the bean
	q := h.database.Collection("beans").Where("slug", "==", slug)
	iter := q.Documents(ctx)
	defer iter.Stop()
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			// TODO: Handle error.
		}
		docID = doc.Ref.ID
	}

	bean := h.database.Collection("beans").Doc(docID)
	docsnap, err := bean.Get(ctx)
	if err != nil {
		h.logger.Error(err)
		http.Error(w, "invalid bean slug", http.StatusBadRequest)
		return
	}

	// Update the bean
	result, _ := bean.Update(
		ctx,
		[]firestore.Update{
			{Path: "countries", Value: req.Countries},
			{Path: "flavors", Value: req.Flavors},
			{Path: "description", Value: req.Description},
			{Path: "name", Value: req.Name},
			{Path: "photo", Value: req.Photo},
			{Path: "roaster.name", Value: req.Roaster.Name},
			{Path: "roaster.slug", Value: req.Roaster.Slug},
			{Path: "slug", Value: req.Slug},
			{Path: "url", Value: req.URL},
		},
	)
	h.logger.Infow(
		"Bean updated",
		"id", docsnap.Ref.ID,
		"updated_at", result.UpdateTime,
		"updated_by", userEmail,
	)

	// Publish an entry in BigQuery
	h.recordBeanChange(ctx, req, userEmail)

	// Send a webhook event to Discord
	msg, err := h.postBeanToDiscord(req, userEmail, "edit")
	h.logger.Info(msg)
	h.logger.Error(err)

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
	Bean
}

// AddBeanResp is the response from the POST /beans endpoint
type AddBeanResp struct {
	ID string `json:"id"`
}

func (h *Handler) addBean(w http.ResponseWriter, r *http.Request) {
	var (
		ctx       = context.TODO()
		err       error
		req       BeanReq
		resp      = &AddBeanResp{}
		userEmail = r.Header.Get("X-User-Email")
	)

	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Make sure roaster exists
	iter := h.database.Collection("roasters").Where("name", "==", req.Roaster.Name).Documents(ctx)
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
	h.logger.Infow(
		"Bean added",
		"id", doc.ID,
		"updated_by", userEmail,
	)

	resp.ID = doc.ID

	// Publish an entry in BigQuery
	h.recordBeanChange(ctx, req, userEmail)

	// Send a webhook event to Discord
	msg, err := h.postBeanToDiscord(req, userEmail, "add")
	h.logger.Info(msg)
	h.logger.Error(err)

	w.WriteHeader(http.StatusAccepted)

	json.NewEncoder(w).Encode(resp)
}
