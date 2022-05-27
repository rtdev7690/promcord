package discord

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	gp "github.com/exsocial/goperspective"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var insultScore = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "insult_score",
	Help:    "A histogram of message insult score.",
	Buckets: prometheus.LinearBuckets(0, .1, 10),
}, []string{"guild", "user"})

var toxicScore = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "toxicity_score",
	Help:    "A histogram of message toxicity score.",
	Buckets: prometheus.LinearBuckets(0, .1, 10),
}, []string{"guild", "user"})

var severeToxicScore = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Name:    "severe_toxicity_score",
	Help:    "A histogram of message severe toxicity score.",
	Buckets: prometheus.LinearBuckets(0, .1, 10),
}, []string{"guild", "user"})

var messageCount = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "message_count",
	Help: "A count of messages",
}, []string{"guild", "user"})

func StartBot(ctx context.Context, token string, perspectiveKey string) (*discordgo.Session, error) {
	client := gp.NewClient(perspectiveKey)

	d, err := discordgo.New("Bot " + token)
	if err != nil {
		return d, err
	}
	d.ShouldReconnectOnError = true
	d.ShouldRetryOnRateLimit = true
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	d.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot {
			return // ignore bots for now
		}
		messageCount.WithLabelValues(m.GuildID, m.Author.String()).Inc()
		log.Println("Message: ", m.Author.String(), " -> ", m.Message.Content)

		if m.Message.Content != "" {
			resp, err := client.AnalyzeComment(gp.AnalyzeRequest{
				Comment: gp.AnalyzeRequestComment{
					Text: m.Message.Content,
					Type: "PLAIN_TEXT",
				},
				DoNotStore:  true,
				CommunityID: "discord:" + m.GuildID,
				ReqAttr: map[gp.Attribute]gp.AnalyzeRequestAttr{
					"TOXICITY":        {},
					"SEVERE_TOXICITY": {},
					"INSULT":          {},
				},
			})

			if err != nil {
				log.Println("err: ", err.Error())
			} else {
				// _ = encoder.Encode(&resp)
				toxicScore.WithLabelValues(m.GuildID, m.Author.String()).Observe(float64(resp.AttributeScores["TOXICITY"].SummaryScore.Value))
				severeToxicScore.WithLabelValues(m.GuildID, m.Author.String()).Observe(float64(resp.AttributeScores["SEVERE_TOXICITY"].SummaryScore.Value))
				insultScore.WithLabelValues(m.GuildID, m.Author.String()).Observe(float64(resp.AttributeScores["INSULT"].SummaryScore.Value))
			}
		}
	})

	return d, nil
}
