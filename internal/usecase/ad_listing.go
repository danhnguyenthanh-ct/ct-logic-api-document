package usecase

import (
	"context"

	"github.com/carousell/ct-go/pkg/logger"
	"github.com/ct-logic-standard/internal/entity"
)

type AdListingRepository interface {
	GetByListID(ctx context.Context, listID int64) (entity.AdListing, error)
}

type AdListingUC interface {
	GetAdByListID(ctx context.Context, adID int64) (entity.AdListing, error)
}

type adListing struct {
	log                 *logger.Logger
	AdListingRepository AdListingRepository
}

func NewAdListingUC(
	adListingClient AdListingRepository,
) AdListingUC {
	return &adListing{
		log:                 logger.MustNamed("use_case"),
		AdListingRepository: adListingClient,
	}
}

func (al *adListing) GetAdByListID(ctx context.Context, adID int64) (res entity.AdListing, err error) {
	ad, err := al.AdListingRepository.GetByListID(ctx, adID)
	al.log.Infof("resp: %v, err: %s", ad, err)

	return ad, err
}
