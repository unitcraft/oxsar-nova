package economy

// Unit ID-константы для зданий и исследований.
// Источник: configs/construction.yml (поле id).
const (
	// Здания (mode=1)
	IDMetalmine        = 1
	IDSiliconLab       = 2
	IDHydrogenLab      = 3
	IDSolarPlant       = 4
	IDHydrogenPlant    = 5
	IDNanoFactory      = 7
	IDMetalStorage     = 9
	IDSiliconStorage   = 10
	IDHydrogenStorage  = 11
	IDRoboticFactory   = 6
	IDMoonHydrogenLab  = 326

	// Корабли/оборона
	IDSolarSatellite   = 39
	IDGravi            = 28

	// Исследования (mode=2): tech ID в formula.Context.Tech
	IDTechEnergy       = 18 // energy_tech   — energy_ratio
	IDTechLaser        = 23 // laser_tech     — metalmine prod
	IDTechSilicon      = 24 // silicon_tech   — silicon_lab prod
	IDTechHydrogen     = 25 // hydrogen_tech  — hydrogen_lab prod

	// Боевые техи
	IDTechGun          = 15 // gun_tech
	IDTechShield       = 16 // shield_tech
	IDTechShell        = 17 // shell_tech
	IDTechBallistics   = 103 // ballistics_tech
	IDTechMasking      = 104 // masking_tech
	IDShipyard         = 8   // shipyard
	IDDefenseFactory   = 100 // repair_factory / defense_factory
	IDResearchLab      = 12
)

// ProfessionKeyToID — карта ключей профессии к ID юнита/исследования.
// Используется для применения бонусов/штрафов профессии к tech-картe.
var ProfessionKeyToID = map[string]int{
	"metalmine":       IDMetalmine,
	"silicon_lab":     IDSiliconLab,
	"solar_plant":     IDSolarPlant,
	"gun":             IDTechGun,
	"shield_weapon":   IDTechShield,
	"shell_weapon":    IDTechShell,
	"ballistics":      IDTechBallistics,
	"masking":         IDTechMasking,
	"shipyard":        IDShipyard,
	"defense_factory": IDDefenseFactory,
	"computer_tech":   14, // computer_tech
	"gravi":           IDGravi,
	"combustion_drive":  20, // combustion_engine
	"impulse_drive":     21, // impulse_engine
	"hyperspace_drive":  22, // hyperspace_engine
	"rocket_station":    53, // missile_silo
}
