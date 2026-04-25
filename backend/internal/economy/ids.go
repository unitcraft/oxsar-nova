package economy

// Unit ID-константы для зданий и исследований.
// Источник: configs/units.yml (поле id).
const (
	// Здания (mode=1)
	IDMetalmine       = 1
	IDSiliconLab      = 2
	IDHydrogenLab     = 3
	IDSolarPlant      = 4
	IDHydrogenPlant   = 5
	IDRoboticFactory  = 6
	IDNanoFactory     = 7
	IDShipyard        = 8
	IDMetalStorage    = 9
	IDSiliconStorage  = 10
	IDHydrogenStorage = 11
	IDResearchLab     = 12
	IDTerraformer     = 58  // terra_former — увеличивает поля планеты (план 23)
	IDDefenseFactory  = 100 // defense_factory
	IDMoonHydrogenLab = 326
	IDMoonLab         = 350 // moon_lab (план 22 Ф.2.2 — отложен)

	// Корабли/оборона
	IDSolarSatellite = 39
	IDGravi          = 28

	// Двигатели (research, mode=2)
	IDCombustionEngine = 20
	IDImpulseEngine    = 21
	IDHyperspaceEngine = 22

	// Исследования (mode=2): tech ID в formula.Context.Tech
	IDTechEnergy   = 18 // energy_tech   — energy_ratio
	IDTechLaser    = 23 // laser_tech     — metalmine prod
	IDTechSilicon  = 24 // silicon_tech   — silicon_lab prod
	IDTechHydrogen = 25 // hydrogen_tech  — hydrogen_lab prod

	// Боевые техи
	IDTechGun        = 15  // gun_tech
	IDTechShield     = 16  // shield_tech
	IDTechShell      = 17  // shell_tech
	IDTechBallistics = 103 // ballistics_tech
	IDTechMasking    = 104 // masking_tech
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
