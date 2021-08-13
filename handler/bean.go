package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/bwmarrin/discordgo"
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
	DirectSun   bool       `firestore:"direct_sun" json:"direct_sun"`
	FairTrade   bool       `firestore:"fair_trade" json:"fair_trade"`
	Flavors     []string   `firestore:"flavors" json:"flavors"`
	Name        string     `firestore:"name" json:"name"`
	Organic     bool       `firestore:"organic" json:"organic"`
	Photo       string     `firestore:"photo" json:"photo"`
	Roaster     RoasterMap `firestore:"roaster" json:"roaster"`
	Shade       string     `firestore:"shade" json:"shade"`
	Slug        string     `firestore:"slug" json:"slug"`
	URL         string     `firestore:"url" json:"url"`
	Year        int64      `firestore:"year" json:"year"`
}

type BeanSimple struct {
	Name    string `json:"name"`
	Roaster string `json:"roaster"`
	Slug    string `json:"slug"`
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

// BeansListResp returns a list of unique beans
type BeansListResp struct {
	Beans []BeanSimple `json:"beans"`
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

// postBeanToDiscord posts a webhook to Discord when a bean is added or updated
func (h *Handler) postBeanToDiscord(req BeanReq, userEmail string, action string) (*discordgo.Message, error) {
	content := "A bean was added!"
	if action == "edit" {
		content = "A bean was updated!"
	}

	countries := ""
	if len(req.Countries) > 0 {
		countries = strings.Join(req.Countries, ", ")
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
						Value: countries,
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
