package specifications

import i "github.com/AndreyAD1/helsinki-guide/internal"

type NeighbourhoodSpecificationByName struct {
	neigbourhood i.Neighbourhood
}

func NewNeighbourhoodSpecificationByName(n i.Neighbourhood) *NeighbourhoodSpecificationByName {
	return &NeighbourhoodSpecificationByName{n}
}

func (a *NeighbourhoodSpecificationByName) ToSQL() (string, map[string]any) {
	query := `SELECT id, name, municipality, created_at, updated_at, 
	deleted_at FROM neighbourhoods WHERE name = @name AND `
	params := map[string]any{
		"name": a.neigbourhood.Name,
	}
	if a.neigbourhood.Municipality == nil {
		query := query + "municipality is NULL;"
		return query, params
	}

	query = query + "municipality = @municipality;"
	params["municipality"] = a.neigbourhood.Municipality
	return query, params
}