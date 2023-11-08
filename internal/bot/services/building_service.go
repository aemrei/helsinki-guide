package services

import (
	"context"
	"strings"

	"github.com/AndreyAD1/helsinki-guide/internal"
	"github.com/AndreyAD1/helsinki-guide/internal/infrastructure/repositories"
	s "github.com/AndreyAD1/helsinki-guide/internal/infrastructure/specifications"
)

type BuildingService struct {
	storage repositories.BuildingRepository
}

func NewBuildingService(storage repositories.BuildingRepository) BuildingService {
	return BuildingService{storage}
}

type BuildingPreview struct {
	Address string
	Name    string
}

type BuildingDTO struct {
	NameFi         *string `valueLanguage:"fi" nameFi:"Nimi" nameEn:"Name" nameRu:"Имя"`
	NameEn         *string `valueLanguage:"en" nameFi:"Nimi" nameEn:"Name" nameRu:"Имя"`
	NameRu         *string `valueLanguage:"ru" nameFi:"Nimi" nameEn:"Name" nameRu:"Имя"`
	Address        string  `valueLanguage:"all" nameFi:"Katuosoite" nameEn:"Address" nameRu:"Адрес"`
	CompletionYear *int    `valueLanguage:"all" nameFi:"Käyttöönottovuosi" nameEn:"Completion_year" nameRu:"Год_постройки"`
	HistoryFi      *string `valueLanguage:"fi" nameFi:"Rakennushistoria" nameEn:"Building_history" nameRu:"История_здания"`
	HistoryEn      *string `valueLanguage:"en" nameFi:"Rakennushistoria" nameEn:"Building_history" nameRu:"История_здания"`
	HistoryRu      *string `valueLanguage:"ru" nameFi:"Rakennushistoria" nameEn:"Building_history" nameRu:"История_здания"`
}

func NewBuildingDTO(b internal.Building, address string) BuildingDTO {
	return BuildingDTO{
		b.NameFi,
		b.NameEn,
		b.NameRu,
		address,
		b.CompletionYear,
		b.HistoryFi,
		b.HistoryEn,
		b.HistoryRu,
	}
}

func (bs BuildingService) GetBuildingPreviews(
	ctx context.Context,
	addressPrefix string,
	limit,
	offset int,
) ([]BuildingPreview, error) {
	addressPrefix = strings.TrimLeft(addressPrefix, " ")
	spec := s.NewBuildingSpecificationByAlikeAddress(addressPrefix, limit, offset)
	buildings, err := bs.storage.Query(ctx, spec)
	if err != nil {
		return nil, err
	}

	previews := make([]BuildingPreview, len(buildings))
	for i, building := range buildings {
		name := ""
		if building.NameFi != nil {
			name = *building.NameFi
		}
		previews[i] = BuildingPreview{building.Address.StreetAddress, name}
	}
	return previews, nil
}

func (bs BuildingService) GetBuildingsByAddress(
	ctx context.Context,
	address string,
) ([]BuildingDTO, error) {
	address = strings.TrimSpace(address)
	spec := s.NewBuildingSpecificationByAddress(address)
	buildings, err := bs.storage.Query(ctx, spec)
	if err != nil {
		return nil, err
	}
	buildingsDto := make([]BuildingDTO, len(buildings))
	for i, building := range buildings {
		buildingsDto[i] = NewBuildingDTO(building, address)
	}
	return buildingsDto, nil
}
