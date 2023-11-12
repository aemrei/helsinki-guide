package specifications

import "github.com/AndreyAD1/helsinki-guide/internal"

type ActorSpecificationByBuilding struct {
	buildingID int64
}

func NewActorSpecificationByBuilding(buildingID int64) *ActorSpecificationByBuilding {
	return &ActorSpecificationByBuilding{buildingID}
}

func (a *ActorSpecificationByBuilding) ToSQL() (string, map[string]any) {
	query := `SELECT id, name, title_fi, title_en, title_ru, created_at,
	updated_at, deleted_at FROM actors JOIN building_authors ON id = actor_id
	WHERE building_id = @building_id;`
	return query, map[string]any{"building_id": a.buildingID}
}

type ActorSpecificationByName struct {
	actor internal.Actor
}

func NewActorSpecificationByName(a internal.Actor) *ActorSpecificationByName {
	return &ActorSpecificationByName{a}
}

func (a *ActorSpecificationByName) ToSQL() (string, map[string]any) {
	query := `SELECT id, name, title_fi, title_en, title_ru, created_at,
	updated_at, deleted_at FROM actors WHERE name = @name;`
	return query, map[string]any{"name": a.actor.Name}
}
