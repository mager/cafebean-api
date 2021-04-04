package handler

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"

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

// postRoasterToDiscord posts a webhook to Discord when a roaster is added or updated
func (h *Handler) postRoasterToDiscord(req RoasterReq, userEmail string, action string) (*discordgo.Message, error) {
	content := "A roaster was added!"
	if action == "edit" {
		content = "A roaster was updated!"
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
				Title:       req.Name,
				Description: req.City,
				URL:         fmt.Sprintf("https://cafebean.org/roasters/%s", req.Slug),
				Fields: []*discordgo.MessageEmbedField{
					{
						Name:  "Twittter",
						Value: req.Twitter,
					},
					{
						Name:  "Instagram",
						Value: req.Instagram,
					},
				},
				Provider: &discordgo.MessageEmbedProvider{
					URL:  req.URL,
					Name: req.URL,
				},
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL:   req.Logo,
					Width: 32,
				},
			}},
		},
	)
}
