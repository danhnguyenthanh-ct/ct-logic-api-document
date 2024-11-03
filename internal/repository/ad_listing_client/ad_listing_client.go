package ad_listing_client

import (
	"context"
	"fmt"

	"github.com/carousell/ct-go/pkg/logger"
	"github.com/ct-logic-standard/config"
	"github.com/ct-logic-standard/internal/entity"
	"github.com/ct-logic-standard/internal/usecase"
	"github.com/ct-logic-standard/pkg/client"
)

type AdListingResponse struct {
	Ad AdInfo `json:"ad"`
}

type AdInfo struct {
	AdID       int64  `json:"ad_id"`
	ListID     int64  `json:"list_id"`
	CategoryID int64  `json:"category"`
	Body       string `json:"body"`
	Subject    string `json:"subject"`
}

func (adInfo *AdInfo) toEntity() entity.AdListing {
	return entity.AdListing{
		AdId:       adInfo.AdID,
		ListId:     adInfo.ListID,
		CategoryId: adInfo.CategoryID,
		Body:       adInfo.Body,
		Subject:    adInfo.Subject,
	}
}

type adListingClient struct {
	log        *logger.Logger
	conf       *config.Config
	httpClient *client.HTTPClient
}

func NewAdListingClient(
	conf *config.Config,
) usecase.AdListingRepository {
	httpCli := client.NewHttpClient("ad-listing")
	c := &adListingClient{
		log:        logger.MustNamed("ad-listing"),
		conf:       conf,
		httpClient: httpCli,
	}

	return c
}

func (a *adListingClient) GetByListID(ctx context.Context, listID int64) (entity.AdListing, error) {
	var adResp AdListingResponse
	url := fmt.Sprintf("%v/%v", a.conf.Client.AdListingDomain, listID)
	err := a.httpClient.SendHTTPRequest(ctx, "GET", url, nil, &adResp)
	if err != nil {
		return entity.AdListing{}, err
	}

	return adResp.Ad.toEntity(), nil
}
